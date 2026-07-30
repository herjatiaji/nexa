package tools

import "github.com/heraji/jarvis/types"

// Tool defines the interface that all JARVIS tools must implement.
type Tool interface {
	// Name returns the tool's unique identifier.
	Name() string

	// Description returns a human-readable description of what the tool does.
	// This is sent to the LLM to help it decide when to use the tool.
	Description() string

	// Parameters returns a JSON Schema describing the tool's input parameters.
	Parameters() map[string]interface{}

	// Execute runs the tool with the given JSON input and returns the result.
	Execute(input string) (string, error)
}

// ToDefinition converts a Tool to a ToolDefinition for the LLM.
func ToDefinition(t Tool) types.ToolDefinition {
	return types.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters:  t.Parameters(),
	}
}
