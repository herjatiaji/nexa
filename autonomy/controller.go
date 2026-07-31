package autonomy

import (
	"sync"
	"time"
)

// AutonomyLevel defines user-configured permission for proactive AI actions.
type AutonomyLevel int

const (
	AutonomyPassive AutonomyLevel = 0 // Wait for explicit user commands only
	AutonomySuggest AutonomyLevel = 1 // May display toast suggestions (DEFAULT)
	AutonomyAssist  AutonomyLevel = 2 // May take minor non-destructive actions
	AutonomyAgent   AutonomyLevel = 3 // May execute approved multi-step task plans
)

// AutonomyBudget tracks rate limits to prevent NEXA from spamming user.
type AutonomyBudget struct {
	MaxToastsPerHour    int
	MaxActionsPerHour   int
	ToastsThisHour      int
	ActionsThisHour     int
	LastReset           time.Time
	RequireConfirmation bool
}

// AutonomyController manages autonomy level and rate-limits proactive actions.
type AutonomyController struct {
	level  AutonomyLevel
	budget AutonomyBudget
	mu     sync.Mutex
}

func NewAutonomyController(level AutonomyLevel) *AutonomyController {
	return &AutonomyController{
		level: level,
		budget: AutonomyBudget{
			MaxToastsPerHour:    5,
			MaxActionsPerHour:   3,
			LastReset:           time.Now(),
			RequireConfirmation: true,
		},
	}
}

// Level returns current autonomy level.
func (ac *AutonomyController) Level() AutonomyLevel {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	return ac.level
}

// SetLevel updates autonomy level.
func (ac *AutonomyController) SetLevel(level AutonomyLevel) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.level = level
}

// PermissionFor returns a permission score (0.0 to 1.0) based on autonomy level.
func (ac *AutonomyController) PermissionFor(action string) float64 {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.resetBudgetIfExpired()

	switch ac.level {
	case AutonomyPassive:
		if action == "SILENT" || action == "OBSERVE" {
			return 1.0
		}
		return 0.0

	case AutonomySuggest:
		if action == "TOAST" || action == "SUGGEST" {
			if ac.budget.ToastsThisHour >= ac.budget.MaxToastsPerHour {
				return 0.0 // Rate-limited
			}
			return 0.9
		}
		if action == "SPEAK" {
			return 0.3
		}
		return 0.0 // EXECUTE denied

	case AutonomyAssist:
		if action == "TOAST" || action == "SUGGEST" {
			return 1.0
		}
		if action == "SPEAK" {
			return 0.8
		}
		if action == "EXECUTE" {
			if ac.budget.ActionsThisHour >= ac.budget.MaxActionsPerHour {
				return 0.0
			}
			return 0.6
		}
		return 1.0

	case AutonomyAgent:
		return 1.0

	default:
		return 0.5
	}
}

// RecordAction increments action counters.
func (ac *AutonomyController) RecordAction(action string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if action == "TOAST" || action == "SUGGEST" {
		ac.budget.ToastsThisHour++
	} else if action == "EXECUTE" {
		ac.budget.ActionsThisHour++
	}
}

func (ac *AutonomyController) resetBudgetIfExpired() {
	if time.Since(ac.budget.LastReset) > time.Hour {
		ac.budget.ToastsThisHour = 0
		ac.budget.ActionsThisHour = 0
		ac.budget.LastReset = time.Now()
	}
}
