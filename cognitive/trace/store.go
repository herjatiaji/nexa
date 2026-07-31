package trace

import (
	"sync"
)

// TraceStore maintains a thread-safe ring buffer of recent cognitive traces for real-time inspection.
type TraceStore struct {
	buffer []CognitiveTrace
	cap    int
	head   int
	count  int
	mu     sync.RWMutex
}

// NewTraceStore creates a TraceStore with a specified capacity.
func NewTraceStore(capacity int) *TraceStore {
	if capacity <= 0 {
		capacity = 200
	}
	return &TraceStore{
		buffer: make([]CognitiveTrace, capacity),
		cap:    capacity,
	}
}

// Record appends a CognitiveTrace to the in-memory ring buffer.
func (s *TraceStore) Record(t CognitiveTrace) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buffer[s.head] = t
	s.head = (s.head + 1) % s.cap
	if s.count < s.cap {
		s.count++
	}
}

// Recent returns up to n most recent CognitiveTraces in chronological order.
func (s *TraceStore) Recent(n int) []CognitiveTrace {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n <= 0 || s.count == 0 {
		return nil
	}
	if n > s.count {
		n = s.count
	}

	result := make([]CognitiveTrace, n)
	start := (s.head - n + s.cap) % s.cap
	for i := 0; i < n; i++ {
		result[i] = s.buffer[(start+i)%s.cap]
	}
	return result
}

// ByCycleID returns all traces associated with a specific cycleID.
func (s *TraceStore) ByCycleID(cycleID string) []CognitiveTrace {
	all := s.Recent(s.count)
	var matched []CognitiveTrace
	for _, t := range all {
		if t.CycleID == cycleID {
			matched = append(matched, t)
		}
	}
	return matched
}
