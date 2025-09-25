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
	"github.com/swaggest/swgui/v5emb"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		// This is a normal and expected condition when running in a containerized environment
		// where environment variables are injected directly.
		log.Println("No .env file found, relying on environment variables.")
	}

	pineconeAPIKey := getEnv("PINECONE_API_KEY", "")
	pineconeIndexName := getEnv("PINECONE_INDEX_NAME", "semantic-search-api")
	elasticAddress := getEnv("ELASTICSEARCH_ADDRESS", "http://localhost:9200")
	elasticIndexName := getEnv("ELASTICSEARCH_INDEX", "go-semantic-search")

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

	// Add Swagger UI endpoint for API documentation
	spec, err := os.ReadFile("api/spec.yaml")
	if err != nil {
		log.Fatalf("Failed to read OpenAPI spec: %v", err)
	}

	swguiHandler := v5emb.New("Hybrid Search API", "/docs/openapi.yaml", "/docs")
	chiRouter.Mount("/docs", swguiHandler)

	// Add endpoint to serve the spec file itself
	chiRouter.Get("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Write(spec)
	})

	log.Println("Server starting on port 8080...")
	log.Println("API documentation available at http://localhost:8080/docs")
	if err := http.ListenAndServe(":8080", chiRouter); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv reads an environment variable or returns a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
