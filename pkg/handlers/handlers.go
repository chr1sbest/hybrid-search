package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chr1sbest/hybrid-search/pkg/chunker"
	"github.com/chr1sbest/hybrid-search/pkg/embeddings"
	"github.com/chr1sbest/hybrid-search/pkg/search"
	"github.com/chr1sbest/hybrid-search/pkg/storage"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Env holds application-wide dependencies.
type Env struct {
	EmbeddingClient embeddings.EmbeddingClient
	VectorStore     storage.VectorStore
	TextStore       storage.TextStore
	SearchService   *search.SearchService
}

func (env *Env) StoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if reqBody.Text == "" {
		http.Error(w, "'text' field cannot be empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	parentDocID := uuid.New().String()

	// 1. Create chunks from the document text.
	// Using 512 for chunk size and 50 for overlap as a starting point.
	chunkr := chunker.NewChunker(512, 50)
	chunks := chunkr.Chunk(reqBody.Text, parentDocID)

	// 2. Index the full document in the text store and the chunks in the vector store concurrently.
	var g errgroup.Group

	// Index the original, full-text document in Elasticsearch.
	g.Go(func() error {
		parentDoc := storage.Document{
			DocumentID: parentDocID,
			Text:       reqBody.Text,
		}
		return env.TextStore.Index(ctx, parentDoc)
	})

	// Process and upsert all chunks to the vector store.
	g.Go(func() error {
		for _, chunk := range chunks {
			// Create the embedding for the chunk.
			vector, err := env.EmbeddingClient.CreateEmbedding(ctx, chunk.Text)
			if err != nil {
				// In a real app, you might want more robust error handling, like a retry.
				log.Printf("Failed to create embedding for chunk %s: %v", chunk.DocumentID, err)
				continue // Continue to the next chunk
			}

			// Upsert the chunk to the vector store.
			if err := env.VectorStore.Upsert(ctx, chunk, vector); err != nil {
				log.Printf("Failed to upsert chunk %s: %v", chunk.DocumentID, err)
				// Decide if one failed chunk should fail the whole request.
				// For now, we'll log and continue.
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		http.Error(w, "Failed to store document and chunks", http.StatusInternalServerError)
		log.Printf("Error during storage: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Document chunked and stored successfully"})
}

func (env *Env) QueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "'q' query parameter is required", http.StatusBadRequest)
		return
	}

	results, err := env.SearchService.Search(r.Context(), query, 5)
	if err != nil {
		http.Error(w, "Failed to search records", http.StatusInternalServerError)
		log.Printf("Failed to search records: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
