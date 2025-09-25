package storage

// Document represents the canonical data structure for our search items.
// It includes a unique identifier and the text content.
type Document struct {
	DocumentID       string `json:"document_id"`
	ParentDocumentID string `json:"parent_document_id,omitempty"`
	Text             string `json:"text"`
}
