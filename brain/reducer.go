package brain

import (
	"sync"
	"time"

	"github.com/heraji/jarvis/autonomy"
	"github.com/heraji/jarvis/cognitive"
	"github.com/heraji/jarvis/goals"
	"github.com/heraji/jarvis/perception"
)

// StateReducer applies proposed StateActions atomically to BrainState (Redux pattern).
// Modules cannot directly mutate BrainState.
type StateReducer struct {
	mu sync.Mutex
}

func NewStateReducer() *StateReducer {
	return &StateReducer{}
}

// Reduce processes a slice of StateActions and updates BrainState in place.
func (r *StateReducer) Reduce(state *BrainState, actions []StateAction) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, action := range actions {
		switch action.Type {
		case "set_context":
			if ctx, ok := action.Data.(cognitive.CognitiveContext); ok {
				state.Context = ctx
			}

		case "set_percepts":
			if percepts, ok := action.Data.([]perception.Percept); ok {
				state.Percepts = percepts
			}

		case "set_fused":
			if fused, ok := action.Data.(*perception.Percept); ok {
				state.FusedPercept = fused
			}

		case "update_working_memory":
			if wm, ok := action.Data.(cognitive.WorkingMemory); ok {
				state.WorkingMemory = wm
			}

		case "set_social":
			if social, ok := action.Data.(cognitive.SocialState); ok {
				state.Social = social
			}

		case "set_autonomy":
			if level, ok := action.Data.(autonomy.AutonomyLevel); ok {
				state.Autonomy = level
			}

		case "set_goals":
			if goalsList, ok := action.Data.([]goals.Goal); ok {
				state.Goals = goalsList
			}

		case "propose_intent":
			if intent, ok := action.Data.(*cognitive.CognitiveIntent); ok {
				state.PendingIntent = intent
			}

		case "clear_intent":
			state.PendingIntent = nil

		case "add_action":
			if actStr, ok := action.Data.(string); ok {
				state.WorkingMemory.AddAction(actStr, action.Module)
			}

		case "add_problem":
			if probStr, ok := action.Data.(string); ok {
				state.WorkingMemory.AddProblem(probStr, 0.8)
			}
		}
	}

	state.LastTickAt = time.Now()
}
