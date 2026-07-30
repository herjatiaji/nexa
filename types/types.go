package types

// Message represents a single message in the conversation history.
type Message struct {
	Role       string     `json:"role"`                   // "system", "user", "assistant", "tool"
	Content    string     `json:"content"`                // Text content
	ToolCallID string     `json:"tool_call_id,omitempty"` // ID of the tool call this message responds to
	Name       string     `json:"name,omitempty"`          // Tool name (for tool responses)
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`    // Tool calls made by the assistant
}

// ToolDefinition describes a tool's schema for the LLM.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// ToolCall represents a request from the LLM to invoke a tool.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string of arguments
}

// ChatResponse holds the LLM's response, which may be text and/or tool calls.
type ChatResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}
