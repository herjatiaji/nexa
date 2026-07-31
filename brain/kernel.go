package brain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/heraji/jarvis/autonomy"
	"github.com/heraji/jarvis/cognitive"
	"github.com/heraji/jarvis/cognitive/explain"
	"github.com/heraji/jarvis/cognitive/trace"
	"github.com/heraji/jarvis/events"
	"github.com/heraji/jarvis/goals"
	"github.com/heraji/jarvis/perception"
	"github.com/heraji/jarvis/runtime"
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
	traceStore   *trace.TraceStore
	profiler     *TickProfiler
	snapStore    *SnapshotStore
	explainer    *explain.DecisionExplainer
	metrics      *runtime.BrainMetricsManager
	lastExpl     explain.DecisionExplanation
	policyEngine *cognitive.PolicyEngine
	cancel       context.CancelFunc
	mu           sync.RWMutex
}

// NewBrain initializes a new Brain Kernel connected to system EventBus.
func NewBrain(bus *events.EventBus) *Brain {
	autonomyCtrl := autonomy.NewAutonomyController(autonomy.AutonomySuggest)
	goalMgr := goals.NewGoalManager()
	history := events.NewEventHistory(100)
	pq := events.NewPriorityQueue(200)
	trStore := trace.NewTraceStore(200)
	profiler := NewTickProfiler(100)
	snapStore := NewSnapshotStore(50)
	explainer := explain.NewDecisionExplainer()
	metrics := runtime.NewBrainMetricsManager()
	policyEng := cognitive.NewPolicyEngine()

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
		traceStore:   trStore,
		profiler:     profiler,
		snapStore:    snapStore,
		explainer:    explainer,
		metrics:      metrics,
		policyEngine: policyEng,
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

// GetTraces returns recent cognitive traces for dashboard visualization.
func (b *Brain) GetTraces(limit int) []trace.CognitiveTrace {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.traceStore.Recent(limit)
}

// GetMetrics returns real-time cognitive performance telemetry.
func (b *Brain) GetMetrics() runtime.BrainMetricsTelemetry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	avgLatency := b.profiler.AverageLatency()
	return b.metrics.GetTelemetry(avgLatency)
}

// ExplainLastDecision returns human-readable natural language explanation of latest decision.
func (b *Brain) ExplainLastDecision() explain.DecisionExplanation {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastExpl
}

// GetPersistedSnapshot retrieves a saved snapshot by cycle ID.
func (b *Brain) GetPersistedSnapshot(cycleID string) (PersistedSnapshot, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.snapStore.Get(cycleID)
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

	tickStart := time.Now()
	cycleID := fmt.Sprintf("cycle_%d", b.state.CycleCount+1)

	// 1. Drain priority queue
	priorityEvents := b.pq.Drain()

	// 2. Perception Pipeline: Convert raw events -> Percepts
	perceiveSpan := trace.StartSpan(b.traceStore, cycleID, trace.StagePerceive, "perception_pipeline", map[string]int{"events": len(priorityEvents)})
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
	perceiveSpan.End(fused)
	perceiveDuration := time.Since(tickStart)

	// 4. Reduce percepts into BrainState
	b.reducer.Reduce(&b.state, []StateAction{
		{Type: "set_percepts", Data: rawPercepts},
		{Type: "set_fused", Data: fused},
	})

	// 5. Create immutable snapshot for modules
	snapshot := b.state.Snapshot()

	// 6. Observe Phase: Modules read state
	observeStart := time.Now()
	for _, m := range b.modules {
		m.Observe(snapshot)
	}

	// 7. Think Phase: Modules propose StateActions
	thinkSpan := trace.StartSpan(b.traceStore, cycleID, trace.StageThink, "cognitive_modules", map[string]int{"modules": len(b.modules)})
	var proposedActions []StateAction
	for _, m := range b.modules {
		actions := m.Think(snapshot)
		proposedActions = append(proposedActions, actions...)
	}

	// Evaluate policy rules
	ruleResults := b.policyEngine.EvaluateAll(b.state.Context, b.state.WorkingMemory)
	thinkSpan.End(ruleResults)
	thinkDuration := time.Since(observeStart)

	// 8. Reduce module actions into state
	b.reducer.Reduce(&b.state, proposedActions)

	// 9. Arbitrate & Decide
	decideStart := time.Now()
	var candidates []Decision
	for _, rr := range ruleResults {
		candidates = append(candidates, Decision{
			Action:     rr.Action,
			Confidence: rr.Confidence,
			Reason:     rr.Reason,
			RuleName:   rr.RuleName,
			Payload:    rr.Payload,
		})
	}
	winningDecision := b.arbiter.Select(candidates, b.state.Social, b.autonomyCtrl)
	decideDuration := time.Since(decideStart)

	winningAction := "SILENT"
	winningConf := 0.0
	if winningDecision != nil {
		winningAction = winningDecision.Action
		winningConf = winningDecision.Confidence
		b.reducer.Reduce(&b.state, []StateAction{
			{Type: "propose_intent", Data: &cognitive.CognitiveIntent{
				Type:     winningDecision.Action,
				Content:  winningDecision.Reason,
				Priority: 5,
				Source:   winningDecision.RuleName,
				Payload:  winningDecision.Payload,
			}},
		})
	}

	// Generate decision explanation
	b.lastExpl = b.explainer.Explain(
		b.state.Context,
		b.state.Social,
		b.state.Autonomy,
		ruleResults,
		winningAction,
		winningConf,
	)

	// 10. Dispatch pending intent if approved
	if b.state.PendingIntent != nil {
		execSpan := trace.StartSpan(b.traceStore, cycleID, trace.StageExecute, "intent_dispatcher", b.state.PendingIntent)
		b.dispatcher.Dispatch(b.state.PendingIntent)
		b.autonomyCtrl.RecordAction(b.state.PendingIntent.Type)
		b.reducer.Reduce(&b.state, []StateAction{{Type: "clear_intent"}})
		execSpan.End("dispatched")
	}

	// Record cycle metrics & snapshot
	totalDuration := time.Since(tickStart)
	b.profiler.Record(CycleMetrics{
		CycleID:            cycleID,
		TotalDuration:      totalDuration,
		PerceptionDuration: perceiveDuration,
		ThinkingDuration:   thinkDuration,
		DecisionDuration:   decideDuration,
		EventsProcessed:    len(priorityEvents),
		ModulesExecuted:    len(b.modules),
		Timestamp:          tickStart,
	})

	b.snapStore.Save(cycleID, b.state.Snapshot())
	b.metrics.RecordCycle()
	b.metrics.RecordDecision(winningAction)

	b.state.CycleCount++
}
