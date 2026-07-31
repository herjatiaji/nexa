package cognitive

import (
	"fmt"
	"time"

	"github.com/heraji/jarvis/events"
)

// CognitiveIntent represents a proposed proactive action from the brain.
type CognitiveIntent struct {
	Type      string      `json:"type"`     // "toast", "speech", "tool_execute", "silent"
	Content   string      `json:"content"`  // Text content or payload description
	Priority  int         `json:"priority"` // Intent priority
	Source    string      `json:"source"`   // Rule or module that proposed this
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// IntentDispatcher broadcasts approved CognitiveIntents over EventBus.
type IntentDispatcher struct {
	bus *events.EventBus
}

func NewIntentDispatcher(bus *events.EventBus) *IntentDispatcher {
	return &IntentDispatcher{bus: bus}
}

// Dispatch emits system events corresponding to the intent type.
func (d *IntentDispatcher) Dispatch(intent *CognitiveIntent) {
	if intent == nil || intent.Type == "silent" || intent.Type == "OBSERVE" {
		return
	}

	intent.Timestamp = time.Now()

	switch intent.Type {
	case "toast", "TOAST", "SUGGEST":
		d.bus.Emit(events.EventSpeechPlan, map[string]string{
			"type":    "toast",
			"content": intent.Content,
			"source":  intent.Source,
		}, "brain.intent_dispatcher")

	case "speech", "SPEAK":
		d.bus.Emit(events.EventSpeechPlan, map[string]string{
			"type":    "speech",
			"content": intent.Content,
			"source":  intent.Source,
		}, "brain.intent_dispatcher")

	case "tool_execute", "EXECUTE":
		d.bus.Emit(events.EventToolStarted, map[string]interface{}{
			"tool":    intent.Content,
			"payload": intent.Payload,
			"source":  intent.Source,
		}, "brain.intent_dispatcher")

	default:
		d.bus.Emit(events.EventAgentThinking, fmt.Sprintf("Intent dispatched: [%s] %s", intent.Type, intent.Content), "brain.intent_dispatcher")
	}
}
