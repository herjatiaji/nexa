package brain

import (
	"sync"
	"time"
)

// CycleMetrics stores execution time breakdown for a single tick of the Cognitive Cycle.
type CycleMetrics struct {
	CycleID            string        `json:"cycle_id"`
	TotalDuration      time.Duration `json:"total_duration"`
	PerceptionDuration time.Duration `json:"perception_duration"`
	ThinkingDuration   time.Duration `json:"thinking_duration"`
	DecisionDuration   time.Duration `json:"decision_duration"`
	EventsProcessed    int           `json:"events_processed"`
	ModulesExecuted    int           `json:"modules_executed"`
	Timestamp          time.Time     `json:"timestamp"`
}

// TickProfiler tracks latency performance across cognitive ticks.
type TickProfiler struct {
	recentMetrics []CycleMetrics
	maxMetrics    int
	mu            sync.RWMutex
}

func NewTickProfiler(maxMetrics int) *TickProfiler {
	if maxMetrics <= 0 {
		maxMetrics = 100
	}
	return &TickProfiler{
		recentMetrics: make([]CycleMetrics, 0, maxMetrics),
		maxMetrics:    maxMetrics,
	}
}

// Record saves a completed tick's performance metrics.
func (p *TickProfiler) Record(m CycleMetrics) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.recentMetrics = append(p.recentMetrics, m)
	if len(p.recentMetrics) > p.maxMetrics {
		p.recentMetrics = p.recentMetrics[1:]
	}
}

// Recent returns recent metrics.
func (p *TickProfiler) Recent(n int) []CycleMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if n <= 0 || len(p.recentMetrics) == 0 {
		return nil
	}
	if n > len(p.recentMetrics) {
		n = len(p.recentMetrics)
	}

	res := make([]CycleMetrics, n)
	copy(res, p.recentMetrics[len(p.recentMetrics)-n:])
	return res
}

// AverageLatency returns average total cycle latency in milliseconds over recorded ticks.
func (p *TickProfiler) AverageLatency() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.recentMetrics) == 0 {
		return 0.0
	}

	var sum int64
	for _, m := range p.recentMetrics {
		sum += m.TotalDuration.Milliseconds()
	}
	return float64(sum) / float64(len(p.recentMetrics))
}
