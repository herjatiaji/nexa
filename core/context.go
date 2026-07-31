package core

import (
	"strings"
	"sync"
)

// SessionEntity stores dynamic context references during a active conversation.
type SessionEntity struct {
	LastApp    string `json:"last_app"`
	LastPath   string `json:"last_path"`
	LastSearch string `json:"last_search"`
}

// ContextManager maintains active short-term conversation state across turns.
type ContextManager struct {
	entity SessionEntity
	mu     sync.RWMutex
}

// NewContextManager creates a new ContextManager.
func NewContextManager() *ContextManager {
	return &ContextManager{}
}

// UpdateApp sets the last opened/focused desktop application.
func (cm *ContextManager) UpdateApp(appName string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.entity.LastApp = appName
}

// UpdatePath sets the last accessed file or directory path.
func (cm *ContextManager) UpdatePath(path string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.entity.LastPath = path
}

// GetEntity returns current session entity context.
func (cm *ContextManager) GetEntity() SessionEntity {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.entity
}

// ResolveAmbiguity resolves pronouns like "it", "the project", or "the app" using session state.
func (cm *ContextManager) ResolveAmbiguity(userPrompt string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	lower := strings.ToLower(userPrompt)

	// Resolve "close it" / "open it" to last_app
	if (strings.Contains(lower, "close it") || strings.Contains(lower, "focus it") || strings.Contains(lower, "the app")) && cm.entity.LastApp != "" {
		return strings.ReplaceAll(userPrompt, "it", cm.entity.LastApp)
	}

	// Resolve "the project" to last_path
	if strings.Contains(lower, "the project") && cm.entity.LastPath != "" {
		return strings.ReplaceAll(userPrompt, "the project", cm.entity.LastPath)
	}

	return userPrompt
}
