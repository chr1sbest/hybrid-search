package ranking

import (
	"sort"

	"github.com/chr1sbest/hybrid-search/pkg/storage"
)

// ReciprocalRankFusion combines multiple sets of search results using the RRF algorithm.
// It returns a single, re-ranked list of documents.
func ReciprocalRankFusion(resultsSets ...[]storage.SearchResult) []storage.Document {
	const k = 60.0 // RRF constant

	// scores maps document IDs to their RRF scores.
	scores := make(map[string]float64)
	// docs maps document IDs to the actual Document object to avoid duplicates.
	docs := make(map[string]storage.Document)

	for _, results := range resultsSets {
		for i, result := range results {
			rank := i + 1
			score := 1.0 / (k + float64(rank))
			docID := result.Document.DocumentID

			scores[docID] += score
			// If we haven't seen this document, or if the stored version has no text
			// and this one does, store it.
			existingDoc, ok := docs[docID]
			if !ok || (existingDoc.Text == "" && result.Document.Text != "") {
				docs[docID] = result.Document
			}
		}
	}

	// Convert the map of documents to a slice for sorting.
	var rankedDocs []storage.Document
	for _, doc := range docs {
		rankedDocs = append(rankedDocs, doc)
	}

	// Sort the documents by their RRF score in descending order.
	sort.Slice(rankedDocs, func(i, j int) bool {
		return scores[rankedDocs[i].DocumentID] > scores[rankedDocs[j].DocumentID]
	})

	return rankedDocs
}
