package search

import (
	"context"
	"fmt"

	"github.com/chr1sbest/hybrid-search/pkg/embeddings"
	"github.com/chr1sbest/hybrid-search/pkg/ranking"
	"github.com/chr1sbest/hybrid-search/pkg/storage"
	"golang.org/x/sync/errgroup"
)

// Service defines the interface for search operations.
type Service interface {
	Search(ctx context.Context, query string, topK int) ([]storage.Document, error)
}

// SearchService orchestrates hybrid search operations.
// It implements the Service interface.
type SearchService struct {
	embeddingClient embeddings.EmbeddingClient
	vectorStore     storage.VectorStore
	textStore       storage.TextStore
}

// NewSearchService creates a new SearchService.
func NewSearchService(embeddingClient embeddings.EmbeddingClient, vectorStore storage.VectorStore, textStore storage.TextStore) *SearchService {
	return &SearchService{
		embeddingClient: embeddingClient,
		vectorStore:     vectorStore,
		textStore:       textStore,
	}
}

// Search performs a hybrid search across the vector and text stores and re-ranks the results.
func (s *SearchService) Search(ctx context.Context, query string, topK int) ([]storage.Document, error) {
	// 1. Create the vector embedding for the query.
	queryVector, err := s.embeddingClient.CreateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to create query embedding: %w", err)
	}

	// 2. Concurrently search the vector and text stores.
	var vectorResults []storage.SearchResult
	var textResults []storage.SearchResult

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		vectorResults, err = s.vectorStore.Query(ctx, query, queryVector, topK)
		return err
	})

	g.Go(func() error {
		var err error
		textResults, err = s.textStore.Search(ctx, query, topK)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Combine and re-rank the results using RRF
	rankedDocs := ranking.ReciprocalRankFusion(vectorResults, textResults)

	return rankedDocs, nil
}
