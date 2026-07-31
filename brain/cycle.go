package brain

import (
	"time"
)

// CycleScheduler calculates the dynamic tick interval for the Cognitive Cycle.
// Adjusts frequency based on user activity state to minimize CPU usage while staying responsive.
type CycleScheduler struct{}

func NewCycleScheduler() *CycleScheduler {
	return &CycleScheduler{}
}

// NextInterval returns adaptive tick duration:
// - ACTIVE / Voice interaction: 200ms – 500ms
// - NORMAL activity: 2s
// - IDLE: 10s
// - SILENT / Deep Sleep: 30s
func (c *CycleScheduler) NextInterval(snapshot BrainSnapshot) time.Duration {
	if snapshot.PendingIntent != nil {
		return 500 * time.Millisecond // Fast tick while executing intent
	}

	if snapshot.Context.Activity == "voice_interaction" || snapshot.Context.Activity == "wake_word_detected" {
		return 200 * time.Millisecond // Near realtime during voice dialogue
	}

	if snapshot.Context.Activity == "idle" {
		if snapshot.Context.IdleDuration > 5*time.Minute {
			return 30 * time.Second // Deep idle: conserve CPU
		}
		return 10 * time.Second // Short idle
	}

	if snapshot.Context.Activity == "coding_activity" || snapshot.Context.Activity == "terminal_activity" {
		return 2 * time.Second // Normal active work
	}

	return 2 * time.Second // Default
}
