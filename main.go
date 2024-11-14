package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	qdrant "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"raglib/api"
)

var (
	qdrantAddress = flag.String("addr", "localhost:6334", "The address of the Qdrant instance to connect to")
)

const (
	AdaV2EmbeddingSize = 1536
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("errror loading .env file: %v", err)
	}

	conn, err := grpc.DialContext(ctx, *qdrantAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	server := api.NewServer(conn)

	server.Start(ctx)
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
