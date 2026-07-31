package perception

import (
	"fmt"
	"time"

	"github.com/heraji/jarvis/events"
)

// VoicePerceiver interprets voice events (wake words, voice commands).
type VoicePerceiver struct{}

func NewVoicePerceiver() *VoicePerceiver {
	return &VoicePerceiver{}
}

func (p *VoicePerceiver) Name() string {
	return "voice_perceiver"
}

func (p *VoicePerceiver) Interpret(event events.Event) *Percept {
	switch event.Type {
	case events.EventVoiceWake:
		return &Percept{
			Type:       "wake_word_detected",
			Source:     "voice",
			Confidence: 0.98,
			Details:    map[string]string{"action": "wake"},
			Timestamp:  time.Now(),
		}
	case events.EventVoiceCommand:
		cmdStr := fmt.Sprintf("%v", event.Payload)
		return &Percept{
			Type:       "voice_command",
			Source:     "voice",
			Confidence: 0.95,
			Details:    map[string]string{"command": cmdStr},
			Timestamp:  time.Now(),
		}
	case events.EventVoiceResult:
		resStr := fmt.Sprintf("%v", event.Payload)
		return &Percept{
			Type:       "voice_response",
			Source:     "voice",
			Confidence: 0.90,
			Details:    map[string]string{"result": resStr},
			Timestamp:  time.Now(),
		}
	default:
		return nil
	}
}
