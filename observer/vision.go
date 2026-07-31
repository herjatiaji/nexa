package observer

import (
	"sync"
	"time"

	"github.com/heraji/jarvis/events"
)

// IntelligentVisionLoop manages screen vision observation with adaptive sampling:
// - SILENT / IDLE mode: 1 screenshot / 60 seconds
// - PASSIVE mode: 1 screenshot / 30 seconds
// - ACTIVE mode: 1 screenshot / 5 seconds (or on window change)
type IntelligentVisionLoop struct {
	bus       *events.EventBus
	awareness *AwarenessManager
	mu        sync.Mutex
	stopChan  chan struct{}
	running   bool
}

// NewIntelligentVisionLoop creates a new VisionLoop.
func NewIntelligentVisionLoop(bus *events.EventBus, awareness *AwarenessManager) *IntelligentVisionLoop {
	vl := &IntelligentVisionLoop{
		bus:       bus,
		awareness: awareness,
		stopChan:  make(chan struct{}),
	}

	// Listen for window changes to trigger an immediate vision check if ACTIVE
	if bus != nil {
		bus.Subscribe(events.EventWindowChanged, func(e events.Event) {
			if vl.awareness.GetMode() == ModeActive {
				go vl.sampleOnce("window_changed")
			}
		})
	}

	return vl
}

// Start begins adaptive vision observation loop.
func (vl *IntelligentVisionLoop) Start() {
	vl.mu.Lock()
	if vl.running {
		vl.mu.Unlock()
		return
	}
	vl.running = true
	vl.stopChan = make(chan struct{})
	vl.mu.Unlock()

	go vl.loop()
}

// Stop terminates the vision loop.
func (vl *IntelligentVisionLoop) Stop() {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	if !vl.running {
		return
	}
	vl.running = false
	close(vl.stopChan)
}

func (vl *IntelligentVisionLoop) loop() {
	for {
		interval := vl.getSamplingInterval()
		timer := time.NewTimer(interval)

		select {
		case <-vl.stopChan:
			timer.Stop()
			return
		case <-timer.C:
			vl.sampleOnce("timer")
		}
	}
}

func (vl *IntelligentVisionLoop) getSamplingInterval() time.Duration {
	mode := ModePassive
	if vl.awareness != nil {
		mode = vl.awareness.GetMode()
	}

	switch mode {
	case ModeActive:
		return 5 * time.Second
	case ModePassive:
		return 30 * time.Second
	case ModeSilent:
		return 60 * time.Second
	default:
		return 30 * time.Second
	}
}

func (vl *IntelligentVisionLoop) sampleOnce(triggerReason string) {
	// Emit vision observation tick event to event bus
	if vl.bus != nil {
		vl.bus.Emit(events.EventVisionAnalyzed, map[string]interface{}{
			"trigger":   triggerReason,
			"timestamp": time.Now().Format("15:04:05"),
		}, "vision_observer")
	}
}
