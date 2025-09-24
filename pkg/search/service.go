package search

import (
	"context"

	"github.com/chr1sbest/hybrid-search/pkg/ranking"
	"github.com/chr1sbest/hybrid-search/pkg/storage"
	"golang.org/x/sync/errgroup"
)

// SearchService orchestrates hybrid search operations.
type SearchService struct {
	vectorStore storage.VectorStore
	textStore   storage.TextStore
}

// NewSearchService creates a new SearchService.
func NewSearchService(vectorStore storage.VectorStore, textStore storage.TextStore) *SearchService {
	return &SearchService{
		vectorStore: vectorStore,
		textStore:   textStore,
	}
}

// Search performs a hybrid search across the vector and text stores and re-ranks the results.
func (s *SearchService) Search(ctx context.Context, query string, topK int) ([]storage.Document, error) {
	var vectorResults []storage.SearchResult
	var textResults []storage.SearchResult

	g, ctx := errgroup.WithContext(ctx)

	// Concurrently search the vector store
	g.Go(func() error {
		var err error
		vectorResults, err = s.vectorStore.Query(ctx, query, topK)
		return err
	})

	// Concurrently search the text store
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
