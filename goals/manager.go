package goals

import (
	"sync"
	"time"
)

// Task represents a single sub-task within a Goal.
type Task struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// Goal represents a high-level user goal (Goal → Task → Action hierarchy).
type Goal struct {
	ID          string    `json:"id"`
	Description string    `json:"description"` // e.g. "Deploy Kazeer successfully"
	Priority    int       `json:"priority"`    // 1 (low) to 5 (critical)
	Progress    float64   `json:"progress"`    // 0.0 to 1.0
	Status      string    `json:"status"`      // "active", "blocked", "completed"
	Tasks       []Task    `json:"tasks"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GoalManager tracks active goals and subtask completion.
type GoalManager struct {
	goals map[string]*Goal
	mu    sync.RWMutex
}

func NewGoalManager() *GoalManager {
	return &GoalManager{
		goals: make(map[string]*Goal),
	}
}

// Add creates or registers a new goal.
func (gm *GoalManager) Add(g Goal) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	g.CreatedAt = time.Now()
	g.UpdatedAt = time.Now()
	if g.Status == "" {
		g.Status = "active"
	}
	gm.goals[g.ID] = &g
}

// Get retrieves a goal by ID.
func (gm *GoalManager) Get(id string) (*Goal, bool) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	g, ok := gm.goals[id]
	return g, ok
}

// Active returns all currently active goals.
func (gm *GoalManager) Active() []Goal {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	var active []Goal
	for _, g := range gm.goals {
		if g.Status == "active" || g.Status == "blocked" {
			active = append(active, *g)
		}
	}
	return active
}

// UpdateProgress sets progress percentage for a goal.
func (gm *GoalManager) UpdateProgress(id string, progress float64) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if g, ok := gm.goals[id]; ok {
		g.Progress = progress
		g.UpdatedAt = time.Now()
		if progress >= 1.0 {
			g.Status = "completed"
		}
	}
}
