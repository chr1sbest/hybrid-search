package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/chr1sbest/hybrid-search/api"
	"github.com/chr1sbest/hybrid-search/pkg/embeddings"
	"github.com/chr1sbest/hybrid-search/pkg/handlers"
	"github.com/chr1sbest/hybrid-search/pkg/search"
	"github.com/chr1sbest/hybrid-search/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	embeddingClient := embeddings.NewPassthroughEmbeddingService()

	searchService := search.NewSearchService(embeddingClient, vectorStore, textStore)

	env := &handlers.Env{
		EmbeddingClient: embeddingClient,
		VectorStore:     vectorStore,
		TextStore:       textStore,
		SearchService:   searchService,
	}

	// Create the router from the generated OpenAPI spec.
	// Our Env struct implements the api.ServerInterface.
	router := api.Handler(env)

	// Add some middleware for logging and recovery.
	chiRouter := chi.NewRouter()
	chiRouter.Use(middleware.Logger)
	chiRouter.Use(middleware.Recoverer)
	chiRouter.Mount("/", router)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", chiRouter); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
