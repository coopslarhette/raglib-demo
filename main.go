package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"os"
	"raglib/internal/document"
	"raglib/internal/generation"
	"raglib/internal/retrieval"
)

var (
	addr = flag.String("addr", "localhost:6334", "the address to connect to")
)

const (
	AdaV2EmbeddingSize = 1536
)

const CollectionName = "text_collection"

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, err := grpc.DialContext(ctx, *addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	httpClient := http.DefaultClient
	openaiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	collectionsClient := qdrant.NewCollectionsClient(conn)
	if err := maybeRecreateCollection(ctx, collectionsClient, CollectionName); err != nil {
		log.Fatalf("error recreating qdrant collection with name '%v': %v", CollectionName, err)
	}

	serpAPIClient := retrieval.NewSerpApiClient(os.Getenv("SERPAPI_API_KEY"), httpClient)

	retrievers := []retrieval.Retriever{
		retrieval.NewQdrantRetriever(qdrant.NewPointsClient(conn), openaiClient, "text_collection"),
		retrieval.NewSERPRetriever(serpAPIClient),
	}

	documents := make(chan []document.Document, len(retrievers))

	q := "what does emissary do"
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

	if err = wg.Wait(); err != nil {
		log.Fatalf("error while query retrievers: %v", err)
	}
	close(documents)

	var allDocs []document.Document
	for docs := range documents {
		allDocs = append(allDocs, docs...)
	}

	answerer := generation.NewAnswerer(openaiClient)

	text, err := answerer.Generate(ctx, fmt.Sprintf("concisely answer this question: <question>%s</question>", q), allDocs)
	if err != nil {
		log.Fatalf("error while generating text: %v", err)
	}

	println(text)

	os.Exit(0)
}

func maybeRecreateCollection(ctx context.Context, collectionsClient qdrant.CollectionsClient, collectionName string) error {
	containsResponse, err := collectionsClient.CollectionExists(ctx, &qdrant.CollectionExistsRequest{
		CollectionName: collectionName,
	})
	if err != nil {
		return fmt.Errorf("error making request to check if colleciton exists: %v", err)
	}

	if containsResponse != nil && containsResponse.Result != nil && !containsResponse.Result.Exists {
		var defaultSegmentNumber uint64 = 2
		_, err = collectionsClient.Create(ctx, &qdrant.CreateCollection{
			CollectionName: collectionName,
			VectorsConfig: &qdrant.VectorsConfig{Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     AdaV2EmbeddingSize,
					Distance: qdrant.Distance_Cosine,
				},
			}},
			OptimizersConfig: &qdrant.OptimizersConfigDiff{
				DefaultSegmentNumber: &defaultSegmentNumber,
			},
		})
		if err != nil {
			return fmt.Errorf("could not create collection: %v", err)
		}

		log.Println("Collection", collectionName, "created")
	}

	return nil
}
