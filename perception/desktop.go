package perception

import (
	"strings"
	"time"

	"github.com/heraji/jarvis/events"
)

// DesktopPerceiver interprets desktop window title and active application changes.
type DesktopPerceiver struct{}

func NewDesktopPerceiver() *DesktopPerceiver {
	return &DesktopPerceiver{}
}

func (p *DesktopPerceiver) Name() string {
	return "desktop_perceiver"
}

func (p *DesktopPerceiver) Interpret(event events.Event) *Percept {
	if event.Type != events.EventWindowChanged {
		return nil
	}

	payloadStr, ok := event.Payload.(string)
	if !ok {
		return nil
	}

	lower := strings.ToLower(payloadStr)

	// Default unknown activity
	actType := "desktop_activity"
	confidence := 0.6
	details := map[string]string{
		"window": payloadStr,
	}

	// Classify activity from window title
	if strings.Contains(lower, "visual studio code") || strings.Contains(lower, "vscode") || strings.Contains(lower, ".go") || strings.Contains(lower, ".ts") || strings.Contains(lower, ".py") {
		actType = "coding_activity"
		confidence = 0.90
		details["ide"] = "VS Code"
	} else if strings.Contains(lower, "cmd") || strings.Contains(lower, "powershell") || strings.Contains(lower, "terminal") || strings.Contains(lower, "bash") {
		actType = "terminal_activity"
		confidence = 0.85
		details["shell"] = "Terminal"
	} else if strings.Contains(lower, "chrome") || strings.Contains(lower, "edge") || strings.Contains(lower, "firefox") || strings.Contains(lower, "browser") {
		actType = "browsing"
		confidence = 0.80
		if strings.Contains(lower, "youtube") || strings.Contains(lower, "netflix") {
			actType = "media_consumption"
			confidence = 0.88
		} else if strings.Contains(lower, "github") || strings.Contains(lower, "stackoverflow") {
			actType = "researching"
			confidence = 0.85
		}
	} else if strings.Contains(lower, "docker") || strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
		actType = "troubleshooting"
		confidence = 0.82
	}

	return &Percept{
		Type:       actType,
		Source:     "desktop",
		Confidence: confidence,
		Details:    details,
		Timestamp:  time.Now(),
	}
}
