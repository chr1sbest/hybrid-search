package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chr1sbest/hybrid-search/api"
	"github.com/chr1sbest/hybrid-search/pkg/chunker"
	"github.com/chr1sbest/hybrid-search/pkg/embeddings"
	"github.com/chr1sbest/hybrid-search/pkg/search"
	"github.com/chr1sbest/hybrid-search/pkg/storage"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Env holds application-wide dependencies and implements the api.ServerInterface.
type Env struct {
	EmbeddingClient embeddings.EmbeddingClient
	VectorStore     storage.VectorStore
	TextStore       storage.TextStore
	SearchService   *search.SearchService
}

// StoreDocument handles the POST /store endpoint.
func (env *Env) StoreDocument(w http.ResponseWriter, r *http.Request) {
	var req api.StoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		msg := "Invalid request body"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.Error{Message: &msg})
		return
	}

	if req.Text == "" {
		msg := "'text' field cannot be empty"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.Error{Message: &msg})
		return
	}

	ctx := r.Context()
	parentDocID := uuid.New().String()

	chunkr := chunker.NewChunker(512, 50)
	chunks := chunkr.Chunk(req.Text, parentDocID)

	var g errgroup.Group

	g.Go(func() error {
		parentDoc := storage.Document{
			DocumentID: parentDocID,
			Text:       req.Text,
		}
		return env.TextStore.Index(ctx, parentDoc)
	})

	g.Go(func() error {
		for _, chunk := range chunks {
			vector, err := env.EmbeddingClient.CreateEmbedding(ctx, chunk.Text)
			if err != nil {
				log.Printf("Failed to create embedding for chunk %s: %v", chunk.DocumentID, err)
				continue
			}

			if err := env.VectorStore.Upsert(ctx, chunk, vector); err != nil {
				log.Printf("Failed to upsert chunk %s: %v", chunk.DocumentID, err)
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		msg := "Failed to store document and chunks"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(api.Error{Message: &msg})
		log.Printf("Error during storage: %v", err)
		return
	}

	msg := "Document chunked and stored successfully"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(api.SuccessMessage{Message: &msg})
}

// QueryDocuments handles the GET /query endpoint.
func (env *Env) QueryDocuments(w http.ResponseWriter, r *http.Request, params api.QueryDocumentsParams) {
	results, err := env.SearchService.Search(r.Context(), params.Q, 5)
	if err != nil {
		msg := "Failed to search records"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(api.Error{Message: &msg})
		log.Printf("Failed to search records: %v", err)
		return
	}

	// Convert storage.Document to api.Document
	apiResults := make([]api.Document, len(results))
	for i, res := range results {
		// Create copies of the strings to take their address
		docID := res.DocumentID
		parentDocID := res.ParentDocumentID
		text := res.Text

		apiResults[i] = api.Document{
			DocumentId:       &docID,
			ParentDocumentId: &parentDocID,
			Text:             &text,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiResults)
}
