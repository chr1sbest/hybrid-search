package chunker

import (
	"log"

	"github.com/chr1sbest/hybrid-search/pkg/storage"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/textsplitter"
)

// Chunker is a wrapper around the langchaingo text splitter.
type Chunker struct {
	splitter textsplitter.RecursiveCharacter
}

// NewChunker creates a new Chunker.
func NewChunker(chunkSize, chunkOverlap int) *Chunker {
	return &Chunker{
		splitter: textsplitter.NewRecursiveCharacter(
			textsplitter.WithChunkSize(chunkSize),
			textsplitter.WithChunkOverlap(chunkOverlap),
		),
	}
}

// Chunk splits the input text into a slice of Document chunks.
func (c *Chunker) Chunk(text, parentDocID string) []storage.Document {
	// Use the library to split the text into strings.
	chunksText, err := c.splitter.SplitText(text)
	if err != nil {
		// In a real application, you might want to handle this error more gracefully.
		log.Printf("Error chunking text: %v", err)
		return nil
	}

	// Convert the string chunks into our Document model.
	var chunks []storage.Document
	for _, chunkText := range chunksText {
		chunks = append(chunks, storage.Document{
			DocumentID:       uuid.New().String(),
			ParentDocumentID: parentDocID,
			Text:             chunkText,
		})
	}

	return chunks
}
