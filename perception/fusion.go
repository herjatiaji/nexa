package perception

import (
	"time"
)

// FusionEngine performs multi-source confidence fusion on incoming percepts.
// When multiple perceivers report (e.g. desktop + voice + vision), FusionEngine
// resolves conflicts and produces a unified high-confidence Percept.
type FusionEngine struct{}

func NewFusionEngine() *FusionEngine {
	return &FusionEngine{}
}

// Fuse resolves a slice of Percepts into a single primary Percept with adjusted confidence.
func (f *FusionEngine) Fuse(percepts []Percept) *Percept {
	if len(percepts) == 0 {
		return nil
	}

	if len(percepts) == 1 {
		p := percepts[0]
		return &p
	}

	// Group by Percept Type
	typeScores := make(map[string]float64)
	typeCounts := make(map[string]int)
	bestDetails := make(map[string]map[string]string)
	bestSource := make(map[string]string)

	for _, p := range percepts {
		typeScores[p.Type] += p.Confidence
		typeCounts[p.Type]++
		if _, exists := bestDetails[p.Type]; !exists || p.Confidence > typeScores[p.Type]/float64(typeCounts[p.Type]) {
			bestDetails[p.Type] = p.Details
			bestSource[p.Type] = p.Source
		}
	}

	// Voice events take absolute precedence if recent
	for _, p := range percepts {
		if p.Source == "voice" && p.Confidence >= 0.90 {
			return &p
		}
	}

	// Find the percept type with highest total aggregated score
	var topType string
	var maxScore float64
	for t, score := range typeScores {
		// Agreement boost: multiple sources agreeing boosts confidence
		count := float64(typeCounts[t])
		fusedScore := (score / count) * (1.0 + (count-1.0)*0.1)
		if fusedScore > 1.0 {
			fusedScore = 0.98
		}

		if fusedScore > maxScore {
			maxScore = fusedScore
			topType = t
		}
	}

	return &Percept{
		Type:       topType,
		Source:     bestSource[topType],
		Confidence: maxScore,
		Details:    bestDetails[topType],
		Timestamp:  time.Now(),
	}
}
