package storage

import "context"

// SearchResult represents a single item in a search result set.
type SearchResult struct {
	Document Document
	Score    float64
}

// VectorStore defines the interface for vector database operations.
// This allows for semantic search capabilities.
type VectorStore interface {
	// Upsert adds or updates a document. It may receive a pre-computed vector.
	// If the vector is nil, the store is expected to generate it internally.
	Upsert(ctx context.Context, doc Document, vector []float32) error

	// Query searches for documents. It may receive a pre-computed query vector.
	// If the queryVector is nil, the store is expected to generate it from the queryText.
	Query(ctx context.Context, queryText string, queryVector []float32, topK int) ([]SearchResult, error)
}

// TextStore defines the interface for text-based search operations.
// This is typically used for keyword matching and full-text search.
type TextStore interface {
	Index(ctx context.Context, doc Document) error
	Search(ctx context.Context, queryText string, topK int) ([]SearchResult, error)
}
