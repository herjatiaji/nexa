package events

import (
	"sort"
	"sync"
)

// Priority levels for system events (1–10).
const (
	PriorityCritical = 10 // Voice command, emergency alert
	PriorityHigh     = 7  // Build failure, docker error
	PriorityMedium   = 5  // Window changed, app switch
	PriorityLow      = 1  // Mouse movement, minor background tick
)

// PriorityEvent wraps an Event with its priority weight.
type PriorityEvent struct {
	Event    Event
	Priority int
}

// PriorityQueue stores system events ordered by priority to prevent event flooding.
type PriorityQueue struct {
	items []PriorityEvent
	cap   int
	mu    sync.Mutex
}

// NewPriorityQueue creates a PriorityQueue with max capacity.
func NewPriorityQueue(capacity int) *PriorityQueue {
	if capacity <= 0 {
		capacity = 200
	}
	return &PriorityQueue{
		items: make([]PriorityEvent, 0, capacity),
		cap:   capacity,
	}
}

// Push adds an event with a priority. If full, drops lowest priority item.
func (pq *PriorityQueue) Push(event Event, priority int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	item := PriorityEvent{Event: event, Priority: priority}

	if len(pq.items) >= pq.cap {
		// Drop lowest priority if incoming item has higher priority
		sort.Slice(pq.items, func(i, j int) bool {
			return pq.items[i].Priority < pq.items[j].Priority
		})
		if pq.items[0].Priority < priority {
			pq.items[0] = item
		}
		return
	}

	pq.items = append(pq.items, item)
}

// Drain returns all events ordered by priority descending and resets queue.
func (pq *PriorityQueue) Drain() []PriorityEvent {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	result := make([]PriorityEvent, len(pq.items))
	copy(result, pq.items)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority > result[j].Priority
	})

	pq.items = pq.items[:0]
	return result
}

// Len returns current number of queued items.
func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.items)
}
