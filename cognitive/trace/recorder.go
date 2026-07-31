package trace

import (
	"fmt"
	"sync/atomic"
	"time"
)

var globalTraceSeq uint64

// TraceSpan measures execution time and outputs of a single component phase.
type TraceSpan struct {
	CycleID   string
	Stage     TraceStage
	Component string
	Input     interface{}
	startTime time.Time
	store     *TraceStore
}

// StartSpan begins recording a new trace span.
func StartSpan(store *TraceStore, cycleID string, stage TraceStage, component string, input interface{}) *TraceSpan {
	return &TraceSpan{
		CycleID:   cycleID,
		Stage:     stage,
		Component: component,
		Input:     input,
		startTime: time.Now(),
		store:     store,
	}
}

// End finishes the trace span, calculates duration, and records the output to store.
func (s *TraceSpan) End(output interface{}) CognitiveTrace {
	dur := time.Since(s.startTime)
	idNum := atomic.AddUint64(&globalTraceSeq, 1)

	tr := CognitiveTrace{
		ID:        fmt.Sprintf("tr_%d", idNum),
		CycleID:   s.CycleID,
		Stage:     s.Stage,
		Component: s.Component,
		Input:     s.Input,
		Output:    output,
		Duration:  dur,
		Timestamp: s.startTime,
	}

	if s.store != nil {
		s.store.Record(tr)
	}

	return tr
}
