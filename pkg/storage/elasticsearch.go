package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticsearchClient wraps the Elasticsearch client and implements the TextStore interface.
type ElasticsearchClient struct {
	client    *elasticsearch.Client
	indexName string
}

// NewElasticsearchClient creates a new client for Elasticsearch and ensures the index exists.
func NewElasticsearchClient(address, indexName string) (*ElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{address},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating the Elasticsearch client: %w", err)
	}

	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("error getting Elasticsearch response: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	client := &ElasticsearchClient{client: es, indexName: indexName}
	if err := client.createIndexIfNotExists(); err != nil {
		return nil, err
	}

	log.Println("Elasticsearch client initialized successfully")
	return client, nil
}

// createIndexIfNotExists checks if the index exists and creates it if it doesn't.
func (c *ElasticsearchClient) createIndexIfNotExists() error {
	res, err := c.client.Indices.Exists([]string{c.indexName})
	if err != nil {
		return fmt.Errorf("error checking if index exists: %w", err)
	}

	if res.StatusCode == 404 {
		log.Printf("Index '%s' not found, creating...", c.indexName)
		mapping := `{
			"mappings": {
				"properties": {
					"document_id": {"type": "keyword"},
					"text": {"type": "text"}
				}
			}
		}`

		res, err = c.client.Indices.Create(
			c.indexName,
			c.client.Indices.Create.WithBody(bytes.NewReader([]byte(mapping))),
		)

		if err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}
		if res.IsError() {
			return fmt.Errorf("error creating index: %s", res.String())
		}
		log.Printf("Index '%s' created.", c.indexName)
	} else {
		log.Printf("Index '%s' already exists.", c.indexName)
	}
	return nil
}

// Index adds a document to the Elasticsearch index.
func (c *ElasticsearchClient) Index(ctx context.Context, doc Document) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("error marshalling document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      c.indexName,
		DocumentID: doc.DocumentID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document ID=%s: %s", doc.DocumentID, res.String())
	}
	return nil
}

// Search performs a full-text search on the Elasticsearch index.
func (c *ElasticsearchClient) Search(ctx context.Context, queryText string, topK int) ([]SearchResult, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"text": queryText,
			},
		},
		"size": topK,
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(c.indexName),
		c.client.Search.WithBody(&buf),
		c.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch search error: %s", res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %w", err)
	}

	var results []SearchResult
	hits, found := r["hits"].(map[string]interface{})["hits"].([]interface{})
	if !found {
		return results, nil
	}

	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		doc := Document{
			DocumentID: source["document_id"].(string),
			Text:       source["text"].(string),
		}
		results = append(results, SearchResult{
			Document: doc,
			Score:    hitMap["_score"].(float64),
		})
	}

	return results, nil
}
