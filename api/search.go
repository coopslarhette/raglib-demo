package api

import (
	"context"
	"fmt"
	"github.com/go-chi/render"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
	"raglib/api/sse"
	"raglib/lib/document"
	"raglib/lib/generation"
	"raglib/lib/retrieval"
	"raglib/lib/retrieval/exa"
	"raglib/lib/retrieval/qdrant"
	"raglib/lib/retrieval/serp"
	"sync"
)

type SearchResponse struct {
	Answer string `json:"answer"`
}

type TextChunk struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type CitationChunk struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

func validateAndExtractParams(r *http.Request) (query string, corpora []string, err error) {
	queryParams := r.URL.Query()
	query = queryParams.Get("q")
	corpora = queryParams["corpus"]

	if len(corpora) == 0 {
		return "", nil, fmt.Errorf("at least one 'corpus' parameter is required")
	}
	if len(query) == 0 {
		return "", nil, fmt.Errorf("query parameter, 'q', is required")
	}

	return query, corpora, nil
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query, corpora, err := validateAndExtractParams(r)
	if err != nil {
		render.Render(w, r, MalformedRequest(err.Error()))
		return
	}

	ctx := r.Context()

	documents, err := s.doRetrieval(ctx, corpora, query)
	if err != nil {
		render.Render(w, r, InternalServerError(err.Error()))
		return
	}

	g, gctx := errgroup.WithContext(ctx)

	answerer := generation.NewAnswerer(s.modelProvider)
	shouldStream := true

	rawChunkChan := make(chan string, 1)
	processedEventChan := make(chan sse.Event, 1)

	g.Go(func() error {
		return answerer.Generate(gctx, query, documents, rawChunkChan, shouldStream)
	})

	if !shouldStream {
		select {
		case text := <-rawChunkChan:
			if err := g.Wait(); err != nil {
				render.Render(w, r, InternalServerError(fmt.Sprintf("error generating answer: %v", err)))
				return
			}
			render.JSON(w, r, SearchResponse{text})
		case <-gctx.Done():
			render.Render(w, r, InternalServerError("context cancelled"))
		}
		return
	}

	chunkProcessor := ChunkProcessor{}
	g.Go(func() error {
		chunkProcessor.ProcessChunks(gctx, rawChunkChan, processedEventChan)
		return nil
	})

	stream := sse.NewStream(w)
	if err := stream.Establish(); err != nil {
		render.Render(w, r, InternalServerError(fmt.Sprintf("error establishing stream: %v", err)))
		return
	}

	documentsReference := sse.Event{EventType: "documentsreference", Data: documents}
	if err := stream.Write(documentsReference); err != nil {
		slog.Error("error occurred writing documents reference to stream", "err", err)
	}

	g.Go(func() error {
		return s.writeEventsToStream(gctx, stream, processedEventChan)
	})

	if err := g.Wait(); err != nil {
		slog.Error("error occurred", "err", err)
		if err := stream.Error("Internal server error occurred."); err != nil {
			slog.Error("error occurred writing error to stream", "err", err)
		}
		return
	}
}

func (s *Server) doRetrieval(ctx context.Context, corpora []string, query string) ([]document.Document, error) {
	retrievers, err := s.determineRetrievers(corpora)
	if err != nil {
		return nil, fmt.Errorf("failed to determine retrievers: %w", err)
	}

	documents, err := retrieveAllDocuments(ctx, query, retrievers)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}
	return documents, nil
}

func (s *Server) determineRetrievers(corpora []string) ([]retrieval.Retriever, error) {
	personalCollectionName := "text_collection"
	var retrieversByCorpus = map[string][]retrieval.Retriever{
		"personal": {
			qdrant.NewRetriever(s.qdrantPointsClient, s.modelProvider.OpenAIClient, personalCollectionName),
		},
		"web": {
			exa.NewRetriever(s.exaAPIClient),
			serp.NewRetriever(s.serpAPIClient),
		},
	}

	retrievers, err := corporaToRetrievers(corpora, retrieversByCorpus)
	if err != nil {
		return nil, err
	}
	return retrievers, nil
}

func corporaToRetrievers(corporaSelection []string, retrieversByCorpus map[string][]retrieval.Retriever) ([]retrieval.Retriever, error) {
	var retrievers []retrieval.Retriever

	for _, corpus := range corporaSelection {
		corpusRetrievers, ok := retrieversByCorpus[corpus]
		if !ok {
			return nil, fmt.Errorf("corpus, %v, is invalid", corpus)
		}
		retrievers = append(retrievers, corpusRetrievers...)
	}

	return retrievers, nil
}

func (s *Server) writeEventsToStream(ctx context.Context, stream sse.Stream, processedEventChan <-chan sse.Event) error {
	defer func() {
		if err := stream.Write(sse.Event{EventType: "done", Data: "DONE"}); err != nil {
			slog.Error("failed to write final done event", "err", err)
		}
	}()

	for {
		select {
		case chunk, ok := <-processedEventChan:
			if !ok {
				return nil
			}
			if err := stream.Write(chunk); err != nil {
				return fmt.Errorf("failed to write event to stream: %w", err)
			}
		case <-ctx.Done():
			// Return reason context is done, or nil
			return ctx.Err()
		}
	}
}

func retrieveAllDocuments(ctx context.Context, q string, retrievers []retrieval.Retriever) ([]document.Document, error) {
	var (
		wg           errgroup.Group
		mu           sync.Mutex
		docsBySource = make(map[string][]document.Document)
	)

	for _, r := range retrievers {
		r := r // capture loop variable
		wg.Go(func() error {
			docs, err := r.Query(ctx, q, 20)
			if err != nil {
				return err
			}

			if len(docs) == 0 {
				return nil
			}

			mu.Lock()
			docsBySource[docs[0].WebReference.APISource] = docs
			mu.Unlock()

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error while retrieving documents: %v", err)
	}

	exaDocs, ok := docsBySource["exa"]
	if !ok {
		return nil, fmt.Errorf("no Exa documents found")
	}

	exaDocsByURL := make(map[string]document.Document, len(exaDocs))
	for _, d := range exaDocs {
		exaDocsByURL[d.WebReference.Link] = d
	}

	serpDocs, ok := docsBySource["serp"]
	if !ok {
		return nil, fmt.Errorf("no SERP documents found")
	}

	seen := make(map[string]struct{})

	// Return 6 documents because based of some YOLO intuition should contain sufficient amount / highly relevant content
	// but not swamp the model with text, also 6 docs looks nicest in the UI
	const documentCountToReturn = 6
	ret := make([]document.Document, 0, documentCountToReturn)

	for _, fromSerp := range serpDocs {
		if len(ret) >= documentCountToReturn {
			break
		}
		fromExa, exists := exaDocsByURL[fromSerp.WebReference.Link]
		if !exists {
			continue
		}

		seen[fromSerp.WebReference.Link] = struct{}{}
		ret = append(ret, fromExa)
	}

	slog.Info("SERP / Exa response stats", "Number SERP results Exa has coverage for", len(ret), "Num SERP retrieved", len(serpDocs), "Num Exa retrieved", len(exaDocs))

	for i := 0; len(ret) < documentCountToReturn && i < len(exaDocs); i++ {
		fromExa := exaDocs[i]
		if _, exists := seen[fromExa.WebReference.Link]; exists {
			continue
		}
		ret = append(ret, fromExa)
	}

	return ret, nil
}
