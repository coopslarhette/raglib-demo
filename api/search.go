package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"golang.org/x/sync/errgroup"
	"net/http"
	"raglib/api/sse"
	"raglib/internal/document"
	"raglib/internal/generation"
	"raglib/internal/retrieval"
	"strings"
)

type SearchResponse struct {
	Answer string `json:"answer"`
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

	bufferedChunkChan := make(chan string, 1)
	go processAndBufferChunks(responseChan, bufferedChunkChan)

	stream := sse.NewStream(w)
	if err = stream.Establish(); err != nil {
		render.Render(w, r, InternalServerError(errors.New(fmt.Sprintf("Error establishing stream: %v", err))))
	}

	documentsReference := sse.Event{EventType: "documentsreference", Data: documents}
	if err := stream.Write(documentsReference); err != nil {
		stream.Error("Error writing to stream.")
	}

	// Send events to the client
	for completionText := range bufferedChunkChan {
		if err := stream.Write(sse.Event{EventType: "completion", Data: completionText}); err != nil {
			stream.Error("Error writing to stream.")
		}
	}

	if err = stream.Write(sse.Event{EventType: "done", Data: "DONE"}); err != nil {
		stream.Error("Error writing to stream.")
	}
}

func processAndBufferChunks(responseChan <-chan string, bufferedChunkChan chan<- string) {
	var buffer strings.Builder
	var isCitation bool

	for chunk := range responseChan {
		for _, char := range chunk {
			if char == '<' {
				if isCitation {
					buffer.WriteRune(char)
				} else {
					if buffer.Len() > 0 {
						bufferedChunkChan <- buffer.String()
						buffer.Reset()
					}
					buffer.WriteRune(char)
					isCitation = true
				}
			} else if char == '>' {
				buffer.WriteRune(char)
				if isCitation && strings.HasSuffix(buffer.String(), "</cited>") {
					bufferedChunkChan <- buffer.String()
					buffer.Reset()
					isCitation = false
				}
			} else {
				buffer.WriteRune(char)
			}
		}
	}

	if buffer.Len() > 0 {
		bufferedChunkChan <- buffer.String()
	}

	close(bufferedChunkChan)
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
