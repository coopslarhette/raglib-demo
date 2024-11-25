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

func (s *Server) determineRetrievers(corpora []string) ([]retrieval.Retriever, error) {
	var webRetriever retrieval.Retriever
	// TODO: use google ranking and document content from Exa
	if true {
		webRetriever = exa.NewRetriever(s.exaAPIClient)
	} else {
		webRetriever = serp.NewRetriever(s.serpAPIClient)
	}

	personalCollectionName := "text_collection"
	var retrieversByCorpus = map[string]retrieval.Retriever{
		"personal": qdrant.NewRetriever(s.qdrantPointsClient, s.modelProvider.OpenAIClient, personalCollectionName),
		"web":      webRetriever,
	}

	retrievers, err := corporaToRetrievers(corpora, retrieversByCorpus)
	if err != nil {
		return nil, err
	}
	return retrievers, nil
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query, corpora, err := validateAndExtractParams(r)
	if err != nil {
		render.Render(w, r, MalformedRequest(err.Error()))
		return
	}

	retrievers, err := s.determineRetrievers(corpora)
	if err != nil {
		render.Render(w, r, MalformedRequest(err.Error()))
		return
	}

	documents, err := retrieveAllDocuments(r.Context(), query, retrievers)
	if err != nil {
		render.Render(w, r, InternalServerError(fmt.Sprintf("error retrieving documents: %v", err)))
		return
	}

	ctx := r.Context()
	g, gctx := errgroup.WithContext(ctx)

	answerer := generation.NewAnswerer(s.modelProvider)
	shouldStream := true

	rawChunkChan := make(chan string, 1)
	processedEventChan := make(chan sse.Event, 1)

	prompt := fmt.Sprintf("<question>%s</question>", query)
	g.Go(func() error {
		return answerer.Generate(gctx, prompt, documents, rawChunkChan, shouldStream)
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

func corporaToRetrievers(corporaSelection []string, retrieversByCorpus map[string]retrieval.Retriever) ([]retrieval.Retriever, error) {
	retrievers := make([]retrieval.Retriever, len(corporaSelection))
	for i, c := range corporaSelection {
		retriever, ok := retrieversByCorpus[c]
		if !ok {
			return nil, fmt.Errorf("corpus, %v, is invalid", c)
		}
		retrievers[i] = retriever
	}
	return retrievers, nil
}

func retrieveAllDocuments(ctx context.Context, q string, retrievers []retrieval.Retriever) ([]document.Document, error) {
	documents := make(chan []document.Document, len(retrievers))

	var wg errgroup.Group
	for _, r := range retrievers {
		wg.Go(func() error {
			docs, err := r.Query(ctx, q, 5)
			if err != nil {
				return err
			}

			documents <- docs
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error while retrieving documents: %v", err)
	}
	close(documents)

	var allDocs []document.Document
	for docs := range documents {
		allDocs = append(allDocs, docs...)
	}

	return allDocs, nil
}
