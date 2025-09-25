package chunker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunker(t *testing.T) {
	parentDocID := "parent-123"

	t.Run("LibraryChunkingWithOverlap", func(t *testing.T) {
		text := "This is the first sentence. This is the second sentence. This is the third sentence."
		chunker := NewChunker(35, 15) // Use values that are easy to verify with the library's behavior.
		chunks := chunker.Chunk(text, parentDocID)

		assert.Len(t, chunks, 4)
		assert.Equal(t, "This is the first sentence. This is", chunks[0].Text)
		assert.Equal(t, "This is the second sentence. This", chunks[1].Text)
		assert.Equal(t, "sentence. This is the third", chunks[2].Text)
		assert.Equal(t, "is the third sentence.", chunks[3].Text)
		for _, chunk := range chunks {
			assert.Equal(t, parentDocID, chunk.ParentDocumentID)
		}
	})

	t.Run("ShortTextReturnsOneChunk", func(t *testing.T) {
		text := "This text is shorter than the chunk size."
		chunker := NewChunker(100, 10)
		chunks := chunker.Chunk(text, parentDocID)
		assert.Len(t, chunks, 1)
		assert.Equal(t, text, chunks[0].Text)
	})
}
