package mcptool

import (
	"encoding/json"
	"fmt"

	"github.com/heraji/jarvis/mcp"
)

// MCPBridgeTool wraps an external MCP tool as a native NEXA tool interface.
type MCPBridgeTool struct {
	client      *mcp.Client
	def         mcp.MCPToolDefinition
	name        string
	description string
}

// NewBridgeTool converts an MCP tool definition into a NEXA Tool interface.
func NewBridgeTool(client *mcp.Client, def mcp.MCPToolDefinition) *MCPBridgeTool {
	return &MCPBridgeTool{
		client:      client,
		def:         def,
		name:        "mcp_" + def.Name,
		description: fmt.Sprintf("[MCP Protocol Tool] %s", def.Description),
	}
}

func (m *MCPBridgeTool) Name() string {
	return m.name
}

func (m *MCPBridgeTool) Description() string {
	return m.description
}

func (m *MCPBridgeTool) Parameters() map[string]interface{} {
	if m.def.InputSchema != nil {
		return m.def.InputSchema
	}
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (m *MCPBridgeTool) Execute(input string) (string, error) {
	var args map[string]interface{}
	if input != "" {
		_ = json.Unmarshal([]byte(input), &args)
	}
	if args == nil {
		args = make(map[string]interface{})
	}

	result, err := m.client.CallTool(m.def.Name, args)
	if err != nil {
		return "", fmt.Errorf("MCP execution error: %w", err)
	}

	return result, nil
}
