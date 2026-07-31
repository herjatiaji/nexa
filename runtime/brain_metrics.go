package runtime

import (
	"sync"
	"sync/atomic"
	"time"
)

// BrainMetricsTelemetry captures system-wide cognitive performance metrics for dashboard inspection.
type BrainMetricsTelemetry struct {
	TotalCycles        uint64    `json:"total_cycles"`
	AverageTickLatency float64   `json:"avg_tick_latency_ms"`
	TotalDecisions     uint64    `json:"total_decisions"`
	SuggestionsEmitted uint64    `json:"suggestions_emitted"`
	SuggestionsAccepted uint64   `json:"suggestions_accepted"`
	SilentObservations uint64    `json:"silent_observations"`
	UpTimeSeconds      int64     `json:"uptime_seconds"`
	LastCycleTime      time.Time `json:"last_cycle_time"`
}

// BrainMetricsManager manages atomic counters and telemetry calculation.
type BrainMetricsManager struct {
	totalCycles         uint64
	totalDecisions      uint64
	suggestionsEmitted  uint64
	suggestionsAccepted uint64
	silentObservations  uint64
	startTime           time.Time
	mu                  sync.RWMutex
}

func NewBrainMetricsManager() *BrainMetricsManager {
	return &BrainMetricsManager{
		startTime: time.Now(),
	}
}

// RecordCycle increments cycle counter.
func (m *BrainMetricsManager) RecordCycle() {
	atomic.AddUint64(&m.totalCycles, 1)
}

// RecordDecision records a decision type (silent vs suggestion).
func (m *BrainMetricsManager) RecordDecision(action string) {
	atomic.AddUint64(&m.totalDecisions, 1)
	if action == "TOAST" || action == "SUGGEST" || action == "SPEAK" {
		atomic.AddUint64(&m.suggestionsEmitted, 1)
	} else {
		atomic.AddUint64(&m.silentObservations, 1)
	}
}

// RecordAcceptance records user accepted a suggestion.
func (m *BrainMetricsManager) RecordAcceptance() {
	atomic.AddUint64(&m.suggestionsAccepted, 1)
}

// GetTelemetry returns current snapshot of telemetry counters.
func (m *BrainMetricsManager) GetTelemetry(avgLatency float64) BrainMetricsTelemetry {
	return BrainMetricsTelemetry{
		TotalCycles:         atomic.LoadUint64(&m.totalCycles),
		AverageTickLatency:  avgLatency,
		TotalDecisions:      atomic.LoadUint64(&m.totalDecisions),
		SuggestionsEmitted:  atomic.LoadUint64(&m.suggestionsEmitted),
		SuggestionsAccepted: atomic.LoadUint64(&m.suggestionsAccepted),
		SilentObservations:  atomic.LoadUint64(&m.silentObservations),
		UpTimeSeconds:       int64(time.Since(m.startTime).Seconds()),
		LastCycleTime:       time.Now(),
	}
}
