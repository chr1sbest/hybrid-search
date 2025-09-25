package search

import (
	"context"
	"testing"

	"github.com/chr1sbest/hybrid-search/pkg/storage"
	embedding_mocks "github.com/chr1sbest/hybrid-search/pkg/embeddings/mocks"
	storage_mocks "github.com/chr1sbest/hybrid-search/pkg/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchService_Search(t *testing.T) {
	// 1. Arrange: Create instances of our mocks
	mockEmbeddingClient := new(embedding_mocks.EmbeddingClient)
	mockVectorStore := new(storage_mocks.VectorStore)
	mockTextStore := new(storage_mocks.TextStore)

	// Create the service with the mocked dependencies
	service := NewSearchService(mockEmbeddingClient, mockVectorStore, mockTextStore)

	// Define test inputs
	ctx := context.Background()
	query := "test query"
	topK := 2
	queryVector := []float32{0.1, 0.2, 0.3}

	// Define the mock data that our stores will return
	vectorResults := []storage.SearchResult{
		{Document: storage.Document{DocumentID: "doc-vec-1"}, Score: 0.9},
		{Document: storage.Document{DocumentID: "doc-shared-1"}, Score: 0.8},
	}
	textResults := []storage.SearchResult{
		{Document: storage.Document{DocumentID: "doc-shared-1"}, Score: 0.95},
		{Document: storage.Document{DocumentID: "doc-text-1"}, Score: 0.7},
	}

	// 2. Act: Set up the expected calls and return values for our mocks
	// We use mock.Anything for the context because the errgroup creates a derived context.
	mockEmbeddingClient.On("CreateEmbedding", mock.Anything, query).Return(queryVector, nil)
	mockVectorStore.On("Query", mock.Anything, query, queryVector, topK).Return(vectorResults, nil)
	mockTextStore.On("Search", mock.Anything, query, topK).Return(textResults, nil)

	// Execute the method we're testing
	results, err := service.Search(ctx, query, topK)

	// 3. Assert: Check that the results are what we expect
	assert.NoError(t, err)
	assert.NotNil(t, results)

	// Based on Reciprocal Rank Fusion, the shared document should be ranked highest.
	// RRF(doc-shared-1) = 1/(60+2) + 1/(60+1) = ~0.032
	// RRF(doc-vec-1) = 1/(60+1) = ~0.016
	// RRF(doc-text-1) = 1/(60+2) = ~0.016
	// Therefore, the expected order is doc-shared-1, doc-vec-1, doc-text-1 (or doc-text-1, doc-vec-1)
	assert.Equal(t, 3, len(results), "Should combine results from both stores")
	assert.Equal(t, "doc-shared-1", results[0].DocumentID, "The highest-ranked document should be first")

	// Verify that all the expected mock calls were made
	mockEmbeddingClient.AssertExpectations(t)
	mockVectorStore.AssertExpectations(t)
	mockTextStore.AssertExpectations(t)
}
