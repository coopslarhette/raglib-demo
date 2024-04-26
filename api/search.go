package api

import (
	"context"
	"errors"
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

	retrievers, errRenderer := toRetrievers(corpora, retrieversByCorpus)
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
	text, err := answerer.Generate(ctx, prompt, documents)
	if err != nil {
		render.Render(w, r, InternalServerError(errors.New(fmt.Sprintf("Error generating answer: %v", err))))
		return
	}

	render.JSON(w, r, SearchResponse{text})
}

func toRetrievers(corpora []string, retrieversByCorpus map[string]retrieval.Retriever) ([]retrieval.Retriever, render.Renderer) {
	retrievers := make([]retrieval.Retriever, len(corpora))
	for i, c := range corpora {
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
