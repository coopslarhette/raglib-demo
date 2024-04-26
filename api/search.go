package api

import (
	"context"
	"fmt"
	"github.com/go-chi/render"
	"golang.org/x/sync/errgroup"
	"net/http"
	"raglib/internal/document"
	"raglib/internal/generation"
	"raglib/internal/retrieval"
)

type SearchResponse struct {
	Answer string `json:"answer"`
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	query := queryParams.Get("q")
	corpora := queryParams["corpus"]
	ctx := r.Context()

	personalSetToRetrieveFrom := "text_collection"

	var retrieversByCorpus = map[string]retrieval.Retriever{
		"personal": retrieval.NewQdrantRetriever(s.qdrantPointsClient, s.openAIClient, personalSetToRetrieveFrom),
		"web":      retrieval.NewSERPRetriever(s.serpAPIClient),
	}

	retrievers := make([]retrieval.Retriever, len(corpora))
	for i, c := range corpora {
		retriever, ok := retrieversByCorpus[c]
		if !ok {
			http.Error(w, fmt.Sprintf("Invalid corpus: %s", c), http.StatusBadRequest)
			return
		}
		retrievers[i] = retriever
	}

	documents, err := retrieveAllDocuments(ctx, query, retrievers)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving documents: %v", err), http.StatusInternalServerError)
		return
	}

	answerer := generation.NewAnswerer(s.openAIClient)

	prompt := fmt.Sprintf("concisely answer this question: <question>%s</question>", query)
	text, err := answerer.Generate(ctx, prompt, documents)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating answer: %v", err), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, SearchResponse{text})
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
