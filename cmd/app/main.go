package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/chr1sbest/hybrid-search/pkg/handlers"
	"github.com/chr1sbest/hybrid-search/pkg/search"
	"github.com/chr1sbest/hybrid-search/pkg/storage"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	pineconeAPIKey := os.Getenv("PINECONE_API_KEY")
	pineconeIndexName := "semantic-search-api"
	elasticAddress := "http://localhost:9200"
	elasticIndexName := "go-semantic-search"

	ctx := context.Background()

	// Initialize Pinecone client
	vectorStore, err := storage.NewPineconeClient(ctx, pineconeAPIKey, pineconeIndexName)
	if err != nil {
		log.Fatalf("Failed to create Pinecone client: %v", err)
	}

	// Initialize Elasticsearch client
	textStore, err := storage.NewElasticsearchClient(elasticAddress, elasticIndexName)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	searchService := search.NewSearchService(vectorStore, textStore)

	env := &handlers.Env{
		VectorStore:   vectorStore,
		TextStore:     textStore,
		SearchService: searchService,
	}

	http.HandleFunc("/store", env.StoreHandler)
	http.HandleFunc("/query", env.QueryHandler)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
