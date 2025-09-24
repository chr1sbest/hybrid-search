package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chr1sbest/hybrid-search/pkg/search"
	"github.com/chr1sbest/hybrid-search/pkg/storage"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Env holds application-wide dependencies.
type Env struct {
	VectorStore   storage.VectorStore
	TextStore     storage.TextStore
	SearchService *search.SearchService
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

	doc := storage.Document{
		DocumentID: uuid.New().String(),
		Text:       reqBody.Text,
	}

	// Upsert to both stores concurrently
	ctx := r.Context()
	var g errgroup.Group

	g.Go(func() error {
		return env.VectorStore.Upsert(ctx, doc)
	})

	g.Go(func() error {
		return env.TextStore.Index(ctx, doc)
	})

	if err := g.Wait(); err != nil {
		http.Error(w, "Failed to store document", http.StatusInternalServerError)
		log.Printf("Failed to store document: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Record stored successfully"})
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
