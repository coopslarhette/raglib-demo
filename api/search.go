package api

import (
	"context"
	"errors"
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

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	query := queryParams.Get("q")
	corpora := queryParams["corpus"]
	ctx := r.Context()

	if len(corpora) == 0 {
		render.Render(w, r, ErrorMalformedRequest(errors.New("at least one 'corpus' parameter is required")))
		return
	} else if len(query) == 0 {
		render.Render(w, r, ErrorMalformedRequest(errors.New("query parameter, 'q', is required")))
		return
	}

	personalCollectionName := "text_collection"
	var retrieversByCorpus = map[string]retrieval.Retriever{
		"personal": retrieval.NewQdrantRetriever(s.qdrantPointsClient, s.openAIClient, personalCollectionName),
		"web":      retrieval.NewSERPRetriever(s.serpAPIClient),
	}

	retrievers, errRenderer := corporaToRetrievers(corpora, retrieversByCorpus)
	if errRenderer != nil {
		render.Render(w, r, errRenderer)
		return
	}

	documents, err := retrieveAllDocuments(ctx, query, retrievers)
	if err != nil {
		render.Render(w, r, InternalServerError(errors.New(fmt.Sprintf("Error retrieving documents: %v", err))))
		return
	}

	answerer := generation.NewAnswerer(s.openAIClient)

	prompt := fmt.Sprintf("concisely answer this question: <question>%s</question>", query)
	responseChan := make(chan string, 1)
	shouldStream := true

	go func() {
		if err := answerer.Generate(ctx, prompt, documents, responseChan, shouldStream); err != nil {
			render.Render(w, r, InternalServerError(errors.New(fmt.Sprintf("Error generating answer: %v", err))))
			return
		}
	}()

	if !shouldStream {
		text := <-responseChan
		render.JSON(w, r, SearchResponse{text})
	}

	bufferedChunkChan := make(chan sse.Event, 1)
	cp := ChunkProcessor{}
	go cp.ProcessChunks(responseChan, bufferedChunkChan)

	stream := sse.NewStream(w)
	if err = stream.Establish(); err != nil {
		render.Render(w, r, InternalServerError(errors.New(fmt.Sprintf("Error establishing stream: %v", err))))
	}

	documentsReference := sse.Event{EventType: "documentsreference", Data: documents}
	if err := stream.Write(documentsReference); err != nil {
		stream.Error("Error writing to stream.")
	}

	// Send events to the client
	for chunk := range bufferedChunkChan {
		if err := stream.Write(chunk); err != nil {
			stream.Error("Error writing to stream.")
		}
	}

	if err = stream.Write(sse.Event{EventType: "done", Data: "DONE"}); err != nil {
		stream.Error("Error writing to stream.")
	}
}

func corporaToRetrievers(corporaSelection []string, retrieversByCorpus map[string]retrieval.Retriever) ([]retrieval.Retriever, render.Renderer) {
	retrievers := make([]retrieval.Retriever, len(corporaSelection))
	for i, c := range corporaSelection {
		retriever, ok := retrieversByCorpus[c]
		if !ok {
			return nil, ErrorMalformedRequest(errors.New(fmt.Sprintf("Corpus, %v, is invalid", c)))
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
