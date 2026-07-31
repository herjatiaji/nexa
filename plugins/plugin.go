package plugins

import (
	"fmt"
	"sync"

	"github.com/heraji/jarvis/config"
	"github.com/heraji/jarvis/tools"
)

// Plugin is the standard interface for external/community NEXA plugins.
type Plugin interface {
	Name() string
	Description() string
	Version() string
	Tools() []tools.Tool
	Init(cfg *config.Config) error
}

// Registry manages registered plugins for NEXA.
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewRegistry creates a new Plugin Registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	r.plugins[name] = p
	return nil
}

// List returns all registered plugins.
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []Plugin
	for _, p := range r.plugins {
		list = append(list, p)
	}
	return list
}
