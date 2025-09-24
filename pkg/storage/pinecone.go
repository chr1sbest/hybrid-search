package storage

import (
	"context"
	"fmt"

	"github.com/pinecone-io/go-pinecone/v4/pinecone"
)

// PineconeClient wraps the Pinecone index connection and implements the VectorStore interface.
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

// Upsert adds or updates a document in the Pinecone index.
func (c *PineconeClient) Upsert(ctx context.Context, doc Document) error {
	records := []*pinecone.IntegratedRecord{
		{"_id": doc.DocumentID, "chunk_text": doc.Text},
	}

	if err := c.idxConn.UpsertRecords(ctx, records); err != nil {
		return fmt.Errorf("failed to upsert record to Pinecone: %w", err)
	}
	return nil
}

// Query performs a semantic search on the Pinecone index.
func (c *PineconeClient) Query(ctx context.Context, queryText string, topK int) ([]SearchResult, error) {

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

	return results, nil
}
