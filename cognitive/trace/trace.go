package trace

import (
	"time"
)

// TraceStage indicates which phase of the cognitive cycle produced the trace.
type TraceStage string

const (
	StagePerceive TraceStage = "perceive"
	StageUpdate   TraceStage = "update"
	StageThink    TraceStage = "think"
	StageDecide   TraceStage = "decide"
	StageExecute  TraceStage = "execute"
)

// CognitiveTrace captures a single reasoning or perception step within a cognitive cycle.
type CognitiveTrace struct {
	ID        string        `json:"id"`
	CycleID   string        `json:"cycle_id"`
	Stage     TraceStage    `json:"stage"`
	Component string        `json:"component"`
	Input     interface{}   `json:"input"`
	Output    interface{}   `json:"output"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}
