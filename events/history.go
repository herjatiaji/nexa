package events

import (
	"sync"
)

// EventHistory maintains a ring buffer of recent events for brain inspection and analysis.
type EventHistory struct {
	buffer []Event
	cap    int
	head   int
	count  int
	mu     sync.RWMutex
}

// NewEventHistory creates a new ring-buffer event history.
func NewEventHistory(capacity int) *EventHistory {
	if capacity <= 0 {
		capacity = 100
	}
	return &EventHistory{
		buffer: make([]Event, capacity),
		cap:    capacity,
	}
}

// Record appends an event to the ring buffer.
func (h *EventHistory) Record(e Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.buffer[h.head] = e
	h.head = (h.head + 1) % h.cap
	if h.count < h.cap {
		h.count++
	}
}

// Recent returns up to n most recent events in chronological order.
func (h *EventHistory) Recent(n int) []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n <= 0 || h.count == 0 {
		return nil
	}
	if n > h.count {
		n = h.count
	}

	result := make([]Event, n)
	start := (h.head - n + h.cap) % h.cap
	for i := 0; i < n; i++ {
		result[i] = h.buffer[(start+i)%h.cap]
	}
	return result
}

// All returns all recorded events in chronological order.
func (h *EventHistory) All() []Event {
	return h.Recent(h.count)
}

// FilterByType returns recent events matching specific EventType.
func (h *EventHistory) FilterByType(eventType EventType, maxResults int) []Event {
	all := h.All()
	var matched []Event
	for i := len(all) - 1; i >= 0; i-- {
		if all[i].Type == eventType {
			matched = append(matched, all[i])
			if len(matched) >= maxResults {
				break
			}
		}
	}
	// Reverse to keep chronological order
	for i, j := 0, len(matched)-1; i < j; i, j = i+1, j-1 {
		matched[i], matched[j] = matched[j], matched[i]
	}
	return matched
}
