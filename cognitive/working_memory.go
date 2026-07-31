package cognitive

import (
	"time"
)

// Action represents a recent action performed by user or NEXA.
type Action struct {
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Timestamp   time.Time `json:"timestamp"`
}

// Problem represents an unresolved issue or error detected.
type Problem struct {
	Description string    `json:"description"`
	Severity    float64   `json:"severity"` // 0.0 – 1.0
	Resolved    bool      `json:"resolved"`
	Timestamp   time.Time `json:"timestamp"`
}

// WorkingMemory maintains short-term transient state with auto-expiration.
type WorkingMemory struct {
	TaskContext    string        `json:"task_context"`     // e.g. "Fixing Docker deployment"
	AttentionFocus string        `json:"attention_focus"`   // e.g. "nginx.conf line 42"
	RecentActions  []Action      `json:"recent_actions"`    // Last 10 actions
	OpenProblems   []Problem     `json:"open_problems"`    // Unresolved issues
	TTL            time.Duration `json:"ttl"`              // Auto-expires after idle
	LastUpdate     time.Time     `json:"last_update"`
}

// NewWorkingMemory creates a fresh WorkingMemory with default 30s TTL.
func NewWorkingMemory() WorkingMemory {
	return WorkingMemory{
		RecentActions: make([]Action, 0, 10),
		OpenProblems:  make([]Problem, 0, 5),
		TTL:           30 * time.Second,
		LastUpdate:    time.Now(),
	}
}

// AddAction records a new action, keeping only the last 10.
func (wm *WorkingMemory) AddAction(desc, source string) {
	action := Action{Description: desc, Source: source, Timestamp: time.Now()}
	wm.RecentActions = append(wm.RecentActions, action)
	if len(wm.RecentActions) > 10 {
		wm.RecentActions = wm.RecentActions[1:]
	}
	wm.LastUpdate = time.Now()
}

// AddProblem records an open problem.
func (wm *WorkingMemory) AddProblem(desc string, severity float64) {
	problem := Problem{Description: desc, Severity: severity, Resolved: false, Timestamp: time.Now()}
	wm.OpenProblems = append(wm.OpenProblems, problem)
	wm.LastUpdate = time.Now()
}

// IsExpired checks if working memory has timed out due to user inactivity.
func (wm *WorkingMemory) IsExpired() bool {
	if wm.TTL <= 0 {
		return false
	}
	return time.Since(wm.LastUpdate) > wm.TTL
}

// Clear resets transient fields upon expiration.
func (wm *WorkingMemory) Clear() {
	wm.TaskContext = ""
	wm.AttentionFocus = ""
	wm.RecentActions = nil
	wm.OpenProblems = nil
	wm.LastUpdate = time.Now()
}
