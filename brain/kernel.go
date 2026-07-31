package brain

import (
	"context"
	"sync"
	"time"

	"github.com/heraji/jarvis/autonomy"
	"github.com/heraji/jarvis/cognitive"
	"github.com/heraji/jarvis/events"
	"github.com/heraji/jarvis/goals"
	"github.com/heraji/jarvis/perception"
)

// Brain orchestrates the Cognitive Cycle (Perceive -> Reduce -> Observe -> Think -> Arbitrate -> Dispatch).
type Brain struct {
	state        BrainState
	reducer      *StateReducer
	modules      []CognitiveModule
	perceivers   []perception.Perceiver
	fusion       *perception.FusionEngine
	arbiter      *Arbiter
	scheduler    *CycleScheduler
	bus          *events.EventBus
	pq           *events.PriorityQueue
	history      *events.EventHistory
	autonomyCtrl *autonomy.AutonomyController
	goalMgr      *goals.GoalManager
	dispatcher   *cognitive.IntentDispatcher
	cancel       context.CancelFunc
	mu           sync.RWMutex
}

// NewBrain initializes a new Brain Kernel connected to system EventBus.
func NewBrain(bus *events.EventBus) *Brain {
	autonomyCtrl := autonomy.NewAutonomyController(autonomy.AutonomySuggest)
	goalMgr := goals.NewGoalManager()
	history := events.NewEventHistory(100)
	pq := events.NewPriorityQueue(200)

	b := &Brain{
		state:        NewBrainState(),
		reducer:      NewStateReducer(),
		modules:      make([]CognitiveModule, 0),
		perceivers:   make([]perception.Perceiver, 0),
		fusion:       perception.NewFusionEngine(),
		arbiter:      NewArbiter(),
		scheduler:    NewCycleScheduler(),
		bus:          bus,
		pq:           pq,
		history:      history,
		autonomyCtrl: autonomyCtrl,
		goalMgr:      goalMgr,
		dispatcher:   cognitive.NewIntentDispatcher(bus),
	}

	// Register default perceivers
	b.RegisterPerceiver(perception.NewDesktopPerceiver())
	b.RegisterPerceiver(perception.NewVoicePerceiver())

	// Subscribe to all event bus messages and push to priority queue & history
	bus.Subscribe("*", func(e events.Event) {
		history.Record(e)

		priority := events.PriorityMedium
		if e.Type == events.EventVoiceCommand || e.Type == events.EventVoiceWake {
			priority = events.PriorityCritical
		} else if e.Type == events.EventAgentError {
			priority = events.PriorityHigh
		} else if e.Type == events.EventWindowChanged {
			priority = events.PriorityMedium
		}

		pq.Push(e, priority)
	})

	return b
}

// RegisterModule attaches a new CognitiveModule to the brain.
func (b *Brain) RegisterModule(m CognitiveModule) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.modules = append(b.modules, m)
}

// RegisterPerceiver attaches a new Perceiver to the perception pipeline.
func (b *Brain) RegisterPerceiver(p perception.Perceiver) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.perceivers = append(b.perceivers, p)
}

// Start launches the background Cognitive Cycle loop.
func (b *Brain) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	go b.runLoop(ctx)
}

// Stop terminates the Cognitive Cycle background loop.
func (b *Brain) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
}

// Snapshot returns current state snapshot.
func (b *Brain) Snapshot() BrainSnapshot {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state.Snapshot()
}

func (b *Brain) runLoop(ctx context.Context) {
	for {
		snapshot := b.Snapshot()
		interval := b.scheduler.NextInterval(snapshot)

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			b.tick()
		}
	}
}

func (b *Brain) tick() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. Drain priority queue
	priorityEvents := b.pq.Drain()

	// 2. Perception Pipeline: Convert raw events -> Percepts
	var rawPercepts []perception.Percept
	for _, pe := range priorityEvents {
		for _, p := range b.perceivers {
			if percept := p.Interpret(pe.Event); percept != nil {
				rawPercepts = append(rawPercepts, *percept)
			}
		}
	}

	// 3. Fusion Engine: Fuse multi-source percepts
	fused := b.fusion.Fuse(rawPercepts)

	// 4. Reduce percepts into BrainState
	b.reducer.Reduce(&b.state, []StateAction{
		{Type: "set_percepts", Data: rawPercepts},
		{Type: "set_fused", Data: fused},
	})

	// 5. Create immutable snapshot for modules
	snapshot := b.state.Snapshot()

	// 6. Observe Phase: Modules read state
	for _, m := range b.modules {
		m.Observe(snapshot)
	}

	// 7. Think Phase: Modules propose StateActions
	var proposedActions []StateAction
	for _, m := range b.modules {
		actions := m.Think(snapshot)
		proposedActions = append(proposedActions, actions...)
	}

	// 8. Reduce module actions into state
	b.reducer.Reduce(&b.state, proposedActions)

	// 9. Dispatch pending intent if approved
	if b.state.PendingIntent != nil {
		b.dispatcher.Dispatch(b.state.PendingIntent)
		b.autonomyCtrl.RecordAction(b.state.PendingIntent.Type)
		b.reducer.Reduce(&b.state, []StateAction{{Type: "clear_intent"}})
	}

	b.state.CycleCount++
}
