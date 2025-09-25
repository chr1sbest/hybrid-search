package storage

import (
	"context"
	"fmt"

	"github.com/pinecone-io/go-pinecone/v4/pinecone"
)

// PineconeClient wraps the Pinecone index connection and implements the VectorStore interface.
// This implementation is for an index with an INTEGRATED embedding model.
type PineconeClient struct {
	idxConn *pinecone.IndexConnection
}

// NewPineconeClient creates and initializes a new client for interacting with a Pinecone index.
func NewPineconeClient(ctx context.Context, apiKey, indexName string) (*PineconeClient, error) {
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pinecone client: %w", err)
	}

	idxModel, err := pc.DescribeIndex(ctx, indexName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe Pinecone index: %w", err)
	}

	idxConn, err := pc.Index(pinecone.NewIndexConnParams{Host: idxModel.Host, Namespace: "ns1"})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pinecone index connection: %w", err)
	}

	return &PineconeClient{idxConn: idxConn}, nil
}

// Upsert uses the integrated embedding model to add or update a document.
// It IGNORES the pre-computed vector argument to satisfy the VectorStore interface.
func (c *PineconeClient) Upsert(ctx context.Context, doc Document, vector []float32) error {
	records := []*pinecone.IntegratedRecord{
		{"_id": doc.DocumentID, "chunk_text": doc.Text},
	}

	if err := c.idxConn.UpsertRecords(ctx, records); err != nil {
		return fmt.Errorf("failed to upsert record to Pinecone: %w", err)
	}
	return nil
}

// Query uses the integrated embedding model to perform a semantic search.
// It IGNORES the pre-computed queryVector argument to satisfy the VectorStore interface.
func (c *PineconeClient) Query(ctx context.Context, queryText string, queryVector []float32, topK int) ([]SearchResult, error) {
	res, err := c.idxConn.SearchRecords(ctx, &pinecone.SearchRecordsRequest{
		Query: pinecone.SearchRecordsQuery{
			TopK: int32(topK),
			Inputs: &map[string]interface{}{
				"text": queryText,
			},
		},
		Fields: &[]string{"chunk_text"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query Pinecone: %w", err)
	}

	var results []SearchResult
	if res != nil {
		for _, hit := range res.Result.Hits {
			var text string
			if chunkText, ok := hit.Fields["chunk_text"]; ok {
				text = chunkText.(string)
			}
			results = append(results, SearchResult{
				Document: Document{
					DocumentID: hit.Id,
					Text:       text,
				},
				Score:    float64(hit.Score),
			})
		}
	}

	return results, nil
}
