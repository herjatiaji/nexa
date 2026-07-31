package brain

import (
	"time"

	"github.com/heraji/jarvis/autonomy"
	"github.com/heraji/jarvis/cognitive"
	"github.com/heraji/jarvis/goals"
	"github.com/heraji/jarvis/perception"
)

// Decision represents a candidate action proposed by a policy rule.
type Decision struct {
	Action     string      `json:"action"`     // "SUGGEST", "TOAST", "SPEAK", "EXECUTE", "SILENT"
	Confidence float64     `json:"confidence"` // 0.0 – 1.0
	Reason     string      `json:"reason"`
	RuleName   string      `json:"rule_name"`
	Payload    interface{} `json:"payload"`
}

// BrainState represents the complete mutable state of the NEXA Brain Kernel.
type BrainState struct {
	Context       cognitive.CognitiveContext
	WorkingMemory cognitive.WorkingMemory
	Goals         []goals.Goal
	Social        cognitive.SocialState
	Autonomy      autonomy.AutonomyLevel
	Percepts      []perception.Percept
	FusedPercept  *perception.Percept
	PendingIntent *cognitive.CognitiveIntent
	CycleCount    uint64
	LastTickAt    time.Time
}

// BrainSnapshot is an immutable, read-only point-in-time copy of BrainState
// passed to CognitiveModules during Observe and Think phases.
type BrainSnapshot struct {
	Context       cognitive.CognitiveContext
	WorkingMemory cognitive.WorkingMemory
	Goals         []goals.Goal
	Social        cognitive.SocialState
	Autonomy      autonomy.AutonomyLevel
	Percepts      []perception.Percept
	FusedPercept  *perception.Percept
	PendingIntent *cognitive.CognitiveIntent
	CycleCount    uint64
	LastTickAt    time.Time
}

// NewBrainState creates initial BrainState with clean defaults.
func NewBrainState() BrainState {
	return BrainState{
		Context:       cognitive.NewCognitiveContext(),
		WorkingMemory: cognitive.NewWorkingMemory(),
		Social:        cognitive.NewSocialState(),
		Autonomy:      autonomy.AutonomySuggest,
		LastTickAt:    time.Now(),
	}
}

// Snapshot creates an immutable read-only snapshot of current state.
func (s *BrainState) Snapshot() BrainSnapshot {
	perceptsCopy := make([]perception.Percept, len(s.Percepts))
	copy(perceptsCopy, s.Percepts)

	goalsCopy := make([]goals.Goal, len(s.Goals))
	copy(goalsCopy, s.Goals)

	var fusedCopy *perception.Percept
	if s.FusedPercept != nil {
		fp := *s.FusedPercept
		fusedCopy = &fp
	}

	var intentCopy *cognitive.CognitiveIntent
	if s.PendingIntent != nil {
		pi := *s.PendingIntent
		intentCopy = &pi
	}

	return BrainSnapshot{
		Context:       s.Context,
		WorkingMemory: s.WorkingMemory,
		Goals:         goalsCopy,
		Social:        s.Social,
		Autonomy:      s.Autonomy,
		Percepts:      perceptsCopy,
		FusedPercept:  fusedCopy,
		PendingIntent: intentCopy,
		CycleCount:    s.CycleCount,
		LastTickAt:    s.LastTickAt,
	}
}
