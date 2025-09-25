package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chr1sbest/hybrid-search/api"
	embedding_mocks "github.com/chr1sbest/hybrid-search/pkg/embeddings/mocks"
	search_mocks "github.com/chr1sbest/hybrid-search/pkg/search/mocks"
	"github.com/chr1sbest/hybrid-search/pkg/storage"
	storage_mocks "github.com/chr1sbest/hybrid-search/pkg/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnv_StoreDocument(t *testing.T) {
	// 1. Arrange
	mockEmbeddingClient := new(embedding_mocks.EmbeddingClient)
	mockVectorStore := new(storage_mocks.VectorStore)
	mockTextStore := new(storage_mocks.TextStore)

	env := &Env{
		EmbeddingClient: mockEmbeddingClient,
		VectorStore:     mockVectorStore,
		TextStore:       mockTextStore,
	}

	// Create a sample request body
	storeReq := api.StoreRequest{Text: "This is a test document."}
	body, _ := json.Marshal(storeReq)
	req := httptest.NewRequest(http.MethodPost, "/store", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// 2. Act: Define mock expectations
	// We expect the TextStore's Index method to be called once for the parent document.
	mockTextStore.On("Index", mock.Anything, mock.AnythingOfType("storage.Document")).Return(nil).Once()

	// We expect the EmbeddingClient to be called for each chunk.
	// Since chunking is internal, we'll just say it can be called any number of times.
	mockEmbeddingClient.On("CreateEmbedding", mock.Anything, mock.AnythingOfType("string")).Return([]float32{0.4, 0.5, 0.6}, nil)

	// We expect the VectorStore's Upsert method to be called for each chunk.
	mockVectorStore.On("Upsert", mock.Anything, mock.AnythingOfType("storage.Document"), mock.AnythingOfType("[]float32")).Return(nil)

	// Execute the handler
	env.StoreDocument(w, req)

	// 3. Assert
	assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP status 201 Created")

	var resp api.SuccessMessage
	_ = json.NewDecoder(w.Body).Decode(&resp)
	assert.NotNil(t, resp.Message)
	assert.Equal(t, "Document chunked and stored successfully", *resp.Message)

	// Verify that the mock expectations were met
	mockTextStore.AssertExpectations(t)
}

func TestEnv_QueryDocuments(t *testing.T) {
	// 1. Arrange
	mockSearchService := new(search_mocks.Service)

	env := &Env{
		SearchService: mockSearchService,
	}

	// Create a sample request
	req := httptest.NewRequest(http.MethodGet, "/query?q=test", nil)
	w := httptest.NewRecorder()

	// Define the mock response from the search service
	mockResults := []storage.Document{
		{DocumentID: "doc-1", Text: "This is the first test document."},
	}

	// Define the API parameters
	params := api.QueryDocumentsParams{Q: "test"}

	// 2. Act: Set up the mock expectation
	mockSearchService.On("Search", mock.Anything, "test", 5).Return(mockResults, nil)

	// Execute the handler
	env.QueryDocuments(w, req, params)

	// 3. Assert
	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP status 200 OK")

	var resp []api.Document
	_ = json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1, "Expected one document in the response")
	assert.Equal(t, "doc-1", *resp[0].DocumentId)
	assert.Equal(t, "This is the first test document.", *resp[0].Text)

	// Verify that the mock expectations were met
	mockSearchService.AssertExpectations(t)
}

