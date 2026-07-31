package brain

// CognitiveModule defines a pluggable module that reads BrainSnapshot and proposes StateActions.
type CognitiveModule interface {
	Name() string
	Observe(snapshot BrainSnapshot)
	Think(snapshot BrainSnapshot) []StateAction
}

// StateAction represents a proposed atomic modification to BrainState.
// State changes must be routed through StateReducer.
type StateAction struct {
	Module string      `json:"module"`
	Type   string      `json:"type"` // e.g. "set_context", "update_working_memory", "propose_intent"
	Data   interface{} `json:"data"`
}
