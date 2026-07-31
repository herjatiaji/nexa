package events

import (
	"fmt"
	"time"
)

// EventMiddleware intercepts events published to EventBus.
type EventMiddleware interface {
	Before(event *Event)
	After(event *Event, duration time.Duration)
}

// EventLogger logs all published system events.
type EventLogger struct{}

func NewEventLogger() *EventLogger {
	return &EventLogger{}
}

func (l *EventLogger) Before(event *Event) {
	// Pre-event hook
}

func (l *EventLogger) After(event *Event, duration time.Duration) {
	// Log events taking > 10ms or critical priority events
	if duration > 10*time.Millisecond || event.Type == EventVoiceCommand || event.Type == EventAgentError {
		fmt.Printf("⚡ [EVENT %s] source=%s dur=%v\n", event.Type, event.Source, duration)
	}
}

// EventProfiler tracks processing duration per event type.
type EventProfiler struct {
	counts map[EventType]uint64
}

func NewEventProfiler() *EventProfiler {
	return &EventProfiler{
		counts: make(map[EventType]uint64),
	}
}

func (p *EventProfiler) Before(event *Event) {
	p.counts[event.Type]++
}

func (p *EventProfiler) After(event *Event, duration time.Duration) {}
