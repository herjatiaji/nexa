package replay

import (
	"fmt"

	"github.com/heraji/jarvis/brain"
	"github.com/heraji/jarvis/cognitive/trace"
)

// ReplayStep represents a single reconstructed cognitive cycle step for replay debugging.
type ReplayStep struct {
	CycleID  string                 `json:"cycle_id"`
	Snapshot brain.PersistedSnapshot`json:"snapshot"`
	Traces   []trace.CognitiveTrace `json:"traces"`
}

// ReplayEngine allows developers to step through historical cognitive cycles.
type ReplayEngine struct {
	steps      []ReplayStep
	currentIdx int
}

func NewReplayEngine() *ReplayEngine {
	return &ReplayEngine{
		steps:      make([]ReplayStep, 0),
		currentIdx: -1,
	}
}

// LoadSession constructs a replay timeline from a sequence of snapshots and traces.
func (r *ReplayEngine) LoadSession(snapshots []brain.PersistedSnapshot, traces []trace.CognitiveTrace) {
	traceMap := make(map[string][]trace.CognitiveTrace)
	for _, tr := range traces {
		traceMap[tr.CycleID] = append(traceMap[tr.CycleID], tr)
	}

	r.steps = make([]ReplayStep, 0, len(snapshots))
	for _, snap := range snapshots {
		r.steps = append(r.steps, ReplayStep{
			CycleID:  snap.CycleID,
			Snapshot: snap,
			Traces:   traceMap[snap.CycleID],
		})
	}

	if len(r.steps) > 0 {
		r.currentIdx = 0
	} else {
		r.currentIdx = -1
	}
}

// CurrentStep returns the current replay step.
func (r *ReplayEngine) CurrentStep() (*ReplayStep, error) {
	if r.currentIdx < 0 || r.currentIdx >= len(r.steps) {
		return nil, fmt.Errorf("no active replay session or out of bounds")
	}
	return &r.steps[r.currentIdx], nil
}

// Next moves forward one cycle step in time.
func (r *ReplayEngine) Next() bool {
	if r.currentIdx+1 < len(r.steps) {
		r.currentIdx++
		return true
	}
	return false
}

// Prev moves backward one cycle step in time.
func (r *ReplayEngine) Prev() bool {
	if r.currentIdx-1 >= 0 {
		r.currentIdx--
		return true
	}
	return false
}
