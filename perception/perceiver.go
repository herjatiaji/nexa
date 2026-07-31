package perception

import (
	"time"

	"github.com/heraji/jarvis/events"
)

// Percept represents a semantic interpretation of raw system events.
type Percept struct {
	Type       string            `json:"type"`       // e.g., "coding_activity", "browsing", "voice_command"
	Source     string            `json:"source"`     // "desktop", "voice", "vision"
	Confidence float64           `json:"confidence"` // 0.0 – 1.0
	Details    map[string]string `json:"details"`
	Timestamp  time.Time         `json:"timestamp"`
}

// Perceiver converts raw low-level events into high-level semantic Percepts.
type Perceiver interface {
	Name() string
	Interpret(event events.Event) *Percept // Returns nil if event is irrelevant
}
