package tools

import (
	"fmt"

	"github.com/heraji/jarvis/types"
)

// Registry manages available tools and dispatches execution requests.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// ListDefinitions returns all tool definitions for sending to the LLM.
func (r *Registry) ListDefinitions() []types.ToolDefinition {
	var defs []types.ToolDefinition
	for _, t := range r.tools {
		defs = append(defs, ToDefinition(t))
	}
	return defs
}

// Execute runs a tool by name with the given input.
func (r *Registry) Execute(name string, input string) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}
	return t.Execute(input)
}
