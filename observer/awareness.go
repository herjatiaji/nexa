package observer

import (
	"sync"
	"time"

	"github.com/heraji/jarvis/events"
)

// AwarenessMode defines how proactively NEXA interacts with the user.
type AwarenessMode string

const (
	// ModeActive: User is actively chatting or using voice. High responsiveness.
	ModeActive AwarenessMode = "ACTIVE"

	// ModePassive: Quietly observe window changes and desktop state. No unsolicited voice.
	ModePassive AwarenessMode = "PASSIVE"

	// ModeSilent: Do not interrupt or play audio. Completely silent background monitoring.
	ModeSilent AwarenessMode = "SILENT"
)

// AwarenessManager maintains current awareness state and handles mode transitions.
type AwarenessManager struct {
	mode          AwarenessMode
	bus           *events.EventBus
	lastActivity  time.Time
	mu            sync.RWMutex
	OnModeChanged func(mode AwarenessMode)
}

// NewAwarenessManager creates an AwarenessManager instance.
func NewAwarenessManager(bus *events.EventBus) *AwarenessManager {
	am := &AwarenessManager{
		mode:         ModePassive,
		bus:          bus,
		lastActivity: time.Now(),
	}

	// Auto-subscribe to voice & user events to set ACTIVE mode on interaction
	if bus != nil {
		bus.Subscribe(events.EventVoiceWake, func(e events.Event) {
			am.SetMode(ModeActive)
		})
		bus.Subscribe(events.EventVoiceCommand, func(e events.Event) {
			am.SetMode(ModeActive)
		})
	}

	return am
}

// GetMode returns the current AwarenessMode.
func (am *AwarenessManager) GetMode() AwarenessMode {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.mode
}

// SetMode updates the awareness mode and emits awareness.mode event.
func (am *AwarenessManager) SetMode(mode AwarenessMode) {
	am.mu.Lock()
	if am.mode == mode {
		am.mu.Unlock()
		return
	}
	am.mode = mode
	am.lastActivity = time.Now()
	am.mu.Unlock()

	if am.bus != nil {
		am.bus.Emit(events.EventAwarenessMode, string(mode), "awareness_manager")
	}
	if am.OnModeChanged != nil {
		am.OnModeChanged(mode)
	}
}

// TouchActivity refreshes last activity timestamp.
func (am *AwarenessManager) TouchActivity() {
	am.mu.Lock()
	am.lastActivity = time.Now()
	am.mu.Unlock()
}

// CheckAutoIdle switches ACTIVE mode to PASSIVE mode after 3 minutes of inactivity.
func (am *AwarenessManager) CheckAutoIdle() {
	am.mu.RLock()
	mode := am.mode
	idleDuration := time.Since(am.lastActivity)
	am.mu.RUnlock()

	if mode == ModeActive && idleDuration > 3*time.Minute {
		am.SetMode(ModePassive)
	}
}
