package core

import (
	"sync"
	"time"
)

// TraceStep records a single step of agent reasoning and tool execution.
type TraceStep struct {
	Index     int    `json:"index"`
	Thought   string `json:"thought"`
	Tool      string `json:"tool"`
	Arguments string `json:"arguments"`
	Result    string `json:"result"`
	Response  string `json:"response"`
	Timestamp string `json:"timestamp"`
}

// TraceRecorder stores and manages execution traces for NEXA.
type TraceRecorder struct {
	steps []TraceStep
	mu    sync.RWMutex
}

// NewTraceRecorder creates a new TraceRecorder instance.
func NewTraceRecorder() *TraceRecorder {
	return &TraceRecorder{
		steps: make([]TraceStep, 0),
	}
}

// AddStep appends a new trace step to the recorder.
func (tr *TraceRecorder) AddStep(step TraceStep) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	step.Index = len(tr.steps) + 1
	if step.Timestamp == "" {
		step.Timestamp = time.Now().Format("15:04:05")
	}
	tr.steps = append(tr.steps, step)
}

// GetSteps returns a copy of all recorded trace steps.
func (tr *TraceRecorder) GetSteps() []TraceStep {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]TraceStep, len(tr.steps))
	copy(result, tr.steps)
	return result
}

// Clear resets the trace history.
func (tr *TraceRecorder) Clear() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.steps = make([]TraceStep, 0)
}
