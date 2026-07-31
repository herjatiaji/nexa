package events

import (
	"sync"
	"time"
)

// EventType identifies the kind of event emitted by system components.
type EventType string

const (
	// Voice events
	EventVoiceWake    EventType = "voice.wake"
	EventVoiceCommand EventType = "voice.command"
	EventVoiceResult  EventType = "voice.result"
	EventVoiceState   EventType = "voice.state"

	// Agent & mascot events
	EventAgentThinking  EventType = "agent.thinking"
	EventAgentExecuting EventType = "agent.executing"
	EventAgentHappy     EventType = "agent.happy"
	EventAgentError    EventType = "agent.error"
	EventEmotionChanged EventType = "emotion.changed"
	EventSpeechPlan     EventType = "speech.plan"

	// Tool events
	EventToolStarted   EventType = "tool.started"
	EventToolCompleted EventType = "tool.completed"

	// Observer & awareness events
	EventWindowChanged  EventType = "window.changed"
	EventVisionAnalyzed EventType = "vision.analyzed"
	EventAwarenessMode  EventType = "awareness.mode"

	// System & status events
	EventSystemStatus EventType = "system.status"
)

// Event represents a strongly typed system event payload.
type Event struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp string      `json:"timestamp"`
	Source    string      `json:"source"`
}

// EventHandler is a callback function executed when a subscribed event fires.
type EventHandler func(event Event)

// EventBus provides thread-safe in-process event publish/subscribe capability.
type EventBus struct {
	subscribers map[EventType][]EventHandler
	middlewares []EventMiddleware
	mu          sync.RWMutex
}

// NewEventBus creates a new EventBus instance.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]EventHandler),
		middlewares: make([]EventMiddleware, 0),
	}
}

// Use attaches a middleware to intercept published events.
func (eb *EventBus) Use(m EventMiddleware) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.middlewares = append(eb.middlewares, m)
}

// Subscribe registers a callback function for a specific event type.
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

// Publish broadcasts an event to all registered subscribers.
func (eb *EventBus) Publish(event Event) {
	if event.Timestamp == "" {
		event.Timestamp = time.Now().Format("15:04:05")
	}

	start := time.Now()

	eb.mu.RLock()
	handlers := make([]EventHandler, len(eb.subscribers[event.Type]))
	copy(handlers, eb.subscribers[event.Type])

	// Also get wildcard subscribers (subscribed to "*")
	wildcardHandlers := make([]EventHandler, len(eb.subscribers["*"]))
	copy(wildcardHandlers, eb.subscribers["*"])

	mws := make([]EventMiddleware, len(eb.middlewares))
	copy(mws, eb.middlewares)
	eb.mu.RUnlock()

	for _, m := range mws {
		m.Before(&event)
	}

	// Execute handlers asynchronously to prevent blocking the publisher
	go func() {
		for _, h := range handlers {
			h(event)
		}
		for _, h := range wildcardHandlers {
			h(event)
		}
		dur := time.Since(start)
		for _, m := range mws {
			m.After(&event, dur)
		}
	}()
}

// EmitConvenience helper creates and publishes an Event quickly.
func (eb *EventBus) Emit(eventType EventType, payload interface{}, source string) {
	eb.Publish(Event{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().Format("15:04:05"),
		Source:    source,
	})
}
