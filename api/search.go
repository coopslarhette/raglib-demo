package api

import (
	"context"
	"fmt"
	"github.com/go-chi/render"
	"golang.org/x/sync/errgroup"
	"net/http"
	"raglib/api/sse"
	"raglib/lib/document"
	"raglib/lib/generation"
	"raglib/lib/retrieval"
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

func (s *Server) validateAndExtractParams(r *http.Request) (query string, corpora []string, err error) {
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
		webRetriever = retrieval.NewExaRetriever(s.exaAPIClient)
	} else {
		webRetriever = retrieval.NewSERPRetriever(s.serpAPIClient)
	}

	personalCollectionName := "text_collection"
	var retrieversByCorpus = map[string]retrieval.Retriever{
		"personal": retrieval.NewQdrantRetriever(s.qdrantPointsClient, s.openAIClient, personalCollectionName),
		"web":      webRetriever,
	}

	retrievers, err := corporaToRetrievers(corpora, retrieversByCorpus)
	if err != nil {
		return nil, err
	}
	return retrievers, nil
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query, corpora, err := s.validateAndExtractParams(r)
	if err != nil {
		render.Render(w, r, MalformedRequest(err.Error()))
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	retrievers, err := s.determineRetrievers(corpora)
	if err != nil {
		render.Render(w, r, MalformedRequest(err.Error()))
		return
	}

	documents, err := retrieveAllDocuments(ctx, query, retrievers)
	if err != nil {
		render.Render(w, r, InternalServerError(fmt.Sprintf("error retrieving documents: %v", err)))
		return
	}

	answerer := generation.NewAnswerer(s.openAIClient)
	shouldStream := true

	rawChunkChan := make(chan string, 1)
	defer close(rawChunkChan)

	prompt := fmt.Sprintf("<question>%s</question>", query)
	go func() {
		if err := answerer.Generate(ctx, prompt, documents, rawChunkChan, shouldStream); err != nil {
			cancel()
			render.Render(w, r, InternalServerError(fmt.Sprintf("error generating answer: %v", err)))
			return
		}
	}()

	if !shouldStream {
		select {
		case text := <-rawChunkChan:
			render.JSON(w, r, SearchResponse{text})
		case <-ctx.Done():
			render.Render(w, r, InternalServerError("request cancelled"))
		}
		return
	}

	processedChunkChan := make(chan sse.Event, 1)

	chunkProcessor := ChunkProcessor{}
	go chunkProcessor.ProcessChunks(ctx, rawChunkChan, processedChunkChan)

	if err := s.setupAndRunSSEStream(w, r, ctx, documents, processedChunkChan); err != nil {
		render.Render(w, r, InternalServerError(fmt.Sprintf("error with SSE stream: %v", err)))
	}
}

func (s *Server) setupAndRunSSEStream(w http.ResponseWriter, r *http.Request, ctx context.Context, documents []document.Document, processedChunkChan <-chan sse.Event) error {
	stream := sse.NewStream(w)
	if err := stream.Establish(); err != nil {
		return fmt.Errorf("error establishing stream: %v", err)
	}

	documentsReference := sse.Event{EventType: "documentsreference", Data: documents}
	if err := stream.Write(documentsReference); err != nil {
		return fmt.Errorf("error writing documents reference: %v", err)
	}

	for {
		select {
		case chunk, ok := <-processedChunkChan:
			if !ok {
				return stream.Write(sse.Event{EventType: "done", Data: "DONE"})
			}
			if err := stream.Write(chunk); err != nil {
				return fmt.Errorf("error writing chunk: %v", err)
			}
		case <-ctx.Done():
			return stream.Write(sse.Event{EventType: "cancelled", Data: "Request cancelled"})
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
