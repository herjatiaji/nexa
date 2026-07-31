package cognitive

import (
	"time"
)

// CognitiveContext captures high-level user activity and environmental status.
type CognitiveContext struct {
	Activity        string        `json:"activity"`         // "coding", "browsing", "idle", "voice_interaction"
	FocusedApp      string        `json:"focused_app"`      // Active window title
	Confidence      float64       `json:"confidence"`       // 0.0 – 1.0
	IdleDuration    time.Duration `json:"idle_duration"`
	ProblemDetected bool          `json:"problem_detected"`
	TimeOfDay       string        `json:"time_of_day"`      // "morning", "afternoon", "evening", "late_night"
	LastChangeAt    time.Time     `json:"last_change_at"`
}

// NewCognitiveContext creates initial default context.
func NewCognitiveContext() CognitiveContext {
	return CognitiveContext{
		Activity:     "idle",
		Confidence:   0.5,
		TimeOfDay:    GetTimeOfDay(time.Now()),
		LastChangeAt: time.Now(),
	}
}

// GetTimeOfDay calculates time bucket based on current hour.
func GetTimeOfDay(t time.Time) string {
	hour := t.Hour()
	if hour >= 5 && hour < 12 {
		return "morning"
	} else if hour >= 12 && hour < 17 {
		return "afternoon"
	} else if hour >= 17 && hour < 23 {
		return "evening"
	}
	return "late_night"
}
