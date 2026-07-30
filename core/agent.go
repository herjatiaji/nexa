package core

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/heraji/jarvis/tools"
	"github.com/heraji/jarvis/types"
)

// LLM defines the interface the agent uses to communicate with language models.
type LLM interface {
	Chat(messages []types.Message, toolDefs []types.ToolDefinition) (*types.ChatResponse, error)
}

// Agent orchestrates the ReAct (Reason + Act) loop.
type Agent struct {
	llm           LLM
	tools         *tools.Registry
	history       []types.Message
	systemPrompt  string
	maxIterations int

	// OnToolCall is called when the agent invokes a tool (for UI feedback).
	OnToolCall func(toolName string, args string)

	// OnToolResult is called when a tool returns a result (for UI feedback).
	OnToolResult func(toolName string, result string)
}

// NewAgent creates a new Agent with the given LLM and tool registry.
func NewAgent(llm LLM, registry *tools.Registry, systemPrompt string, maxIterations int) *Agent {
	// Append current working directory to system prompt
	cwd, _ := os.Getwd()
	fullPrompt := systemPrompt + fmt.Sprintf("\n\nCurrent working directory: %s", cwd)

	agent := &Agent{
		llm:           llm,
		tools:         registry,
		systemPrompt:  fullPrompt,
		maxIterations: maxIterations,
	}

	// Initialize history with system prompt
	agent.history = []types.Message{
		{Role: "system", Content: fullPrompt},
	}

	return agent
}

// Run processes a user input through the ReAct loop and returns the final response.
func (a *Agent) Run(userInput string) (string, error) {
	// Add user message to history
	a.history = append(a.history, types.Message{
		Role:    "user",
		Content: userInput,
	})

	// Get tool definitions for the LLM
	toolDefs := a.tools.ListDefinitions()

	// ReAct loop
	for i := 0; i < a.maxIterations; i++ {
		// Send to LLM
		resp, err := a.llm.Chat(a.history, toolDefs)
		if err != nil {
			return "", fmt.Errorf("LLM error: %w", err)
		}

		// If no native tool calls, check for text-embedded tool calls (e.g. <desktop_apps>{...}</desktop_apps>)
		if len(resp.ToolCalls) == 0 {
			if extractedCall, ok := parseTextToolCall(resp.Content); ok {
				resp.ToolCalls = append(resp.ToolCalls, extractedCall)
			} else {
				a.history = append(a.history, types.Message{
					Role:    "assistant",
					Content: resp.Content,
				})
				return resp.Content, nil
			}
		}

		// Add assistant message with tool calls to history
		a.history = append(a.history, types.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		// Execute each tool call
		for _, tc := range resp.ToolCalls {
			if a.OnToolCall != nil {
				a.OnToolCall(tc.Name, tc.Arguments)
			}

			result, err := a.tools.Execute(tc.Name, tc.Arguments)
			if err != nil {
				result = fmt.Sprintf("Error executing tool %s: %v", tc.Name, err)
			}

			if a.OnToolResult != nil {
				a.OnToolResult(tc.Name, result)
			}

			a.history = append(a.history, types.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
				Name:       tc.Name,
			})
		}
	}

	return "I've reached the maximum number of steps. Here's what I've done so far — please try a simpler request.", nil
}

// Reset clears the conversation history (keeping the system prompt).
func (a *Agent) Reset() {
	a.history = []types.Message{
		{Role: "system", Content: a.systemPrompt},
	}
}

var textToolCallRegex = regexp.MustCompile(`<([a-zA-Z0-9_]+)>\s*(\{[\s\S]*?\})\s*</[a-zA-Z0-9_]+>`)

func parseTextToolCall(content string) (types.ToolCall, bool) {
	matches := textToolCallRegex.FindStringSubmatch(content)
	if len(matches) >= 3 {
		toolName := strings.TrimSpace(matches[1])
		toolArgs := strings.TrimSpace(matches[2])
		return types.ToolCall{
			ID:        fmt.Sprintf("call_text_%d", time.Now().UnixNano()),
			Name:      toolName,
			Arguments: toolArgs,
		}, true
	}
	return types.ToolCall{}, false
}
