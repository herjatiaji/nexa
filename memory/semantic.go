package memory

import (
	"math"
	"strings"
)

// SemanticMatch represents a memory match with a similarity score.
type SemanticMatch struct {
	Fact  Fact    `json:"fact"`
	Score float64 `json:"score"`
}

// Tokenize converts a text string into a frequency map of normalized word tokens.
func Tokenize(text string) map[string]float64 {
	words := strings.Fields(strings.ToLower(text))
	freq := make(map[string]float64)

	// Stop words to ignore
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "is": true, "are": true, "was": true,
		"were": true, "in": true, "on": true, "at": true, "to": true, "for": true,
		"of": true, "and": true, "or": true, "my": true, "me": true, "about": true,
		"remind": true, "tell": true, "what": true, "how": true,
	}

	for _, w := range words {
		w = strings.Trim(w, ".,!?:;\"'()[]{}")
		if len(w) > 1 && !stopWords[w] {
			freq[w] += 1.0
		}
	}
	return freq
}

// CosineSimilarity computes cosine similarity score [0.0 - 1.0] between two token vectors.
func CosineSimilarity(vecA, vecB map[string]float64) float64 {
	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for word, countA := range vecA {
		normA += countA * countA
		if countB, ok := vecB[word]; ok {
			dotProduct += countA * countB
		}
	}

	for _, countB := range vecB {
		normB += countB * countB
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// SearchSemantic performs conceptual/semantic vector search over a list of facts.
func SearchSemantic(facts []Fact, query string, threshold float64) []SemanticMatch {
	queryVec := Tokenize(query)
	var matches []SemanticMatch

	for _, f := range facts {
		// Vector combining key + value
		docText := f.Key + " " + f.Value
		docVec := Tokenize(docText)

		score := CosineSimilarity(queryVec, docVec)

		// Also check partial substring match bonus
		queryLower := strings.ToLower(query)
		if strings.Contains(strings.ToLower(f.Key), queryLower) || strings.Contains(strings.ToLower(f.Value), queryLower) {
			score += 0.3
			if score > 1.0 {
				score = 1.0
			}
		}

		if score >= threshold {
			matches = append(matches, SemanticMatch{Fact: f, Score: score})
		}
	}

	return matches
}
