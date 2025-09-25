package embeddings

import (
	"context"
)

// EmbeddingClient is the interface for any service that can convert text into a vector embedding.
// This allows for a pluggable embedding model.
type EmbeddingClient interface {
	// CreateEmbedding takes a string of text and returns its vector representation.
	CreateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// PassthroughEmbeddingService is a no-op implementation of EmbeddingClient.
// It is used for VectorStore implementations (like the integrated Pinecone client)
// that handle their own embedding generation internally.
// It returns a nil vector, signaling to the VectorStore that it must generate the embedding itself.
type PassthroughEmbeddingService struct{}

// NewPassthroughEmbeddingService creates a new PassthroughEmbeddingService.
func NewPassthroughEmbeddingService() *PassthroughEmbeddingService {
	return &PassthroughEmbeddingService{}
}

// CreateEmbedding does nothing and returns a nil vector and no error.
func (s *PassthroughEmbeddingService) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}
