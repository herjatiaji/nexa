package core

import (
	"encoding/json"
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

// Agent orchestrates the ReAct (Reason + Act) loop with Tracing and Task Planning.
type Agent struct {
	llm           LLM
	tools         *tools.Registry
	history       []types.Message
	systemPrompt  string
	maxIterations int
	tracer        *TraceRecorder
	planner       *Planner

	// OnToolCall is called when the agent invokes a tool (for UI feedback).
	OnToolCall func(toolName string, args string)

	// OnToolResult is called when a tool returns a result (for UI feedback).
	OnToolResult func(toolName string, result string)

	// OnTrace is called when a new trace step is recorded.
	OnTrace func(step TraceStep)

	// OnEmotion is called when mascot emotion state changes.
	OnEmotion func(mascot MascotState)
}

// NewAgent creates a new Agent with the given LLM and tool registry.
func NewAgent(llm LLM, registry *tools.Registry, systemPrompt string, maxIterations int) *Agent {
	cwd, _ := os.Getwd()
	currentDate := time.Now().Format("Monday, January 2, 2006")
	fullPrompt := systemPrompt + fmt.Sprintf("\n\nCurrent Real-Time System Date: %s\nCurrent working directory: %s", currentDate, cwd)

	agent := &Agent{
		llm:           llm,
		tools:         registry,
		systemPrompt:  fullPrompt,
		maxIterations: maxIterations,
		tracer:        NewTraceRecorder(),
		planner:       NewPlanner(llm),
	}

	// Initialize history with system prompt
	agent.history = []types.Message{
		{Role: "system", Content: fullPrompt},
	}

	return agent
}

// GetTraces returns all recorded trace steps.
func (a *Agent) GetTraces() []TraceStep {
	return a.tracer.GetSteps()
}

// CreatePlan decomposes a multi-step user prompt into a task plan.
func (a *Agent) CreatePlan(prompt string) (*Plan, error) {
	return a.planner.CreatePlan(prompt)
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
		if a.OnEmotion != nil {
			a.OnEmotion(GetMascotExpression(EmotionThinking, "Thinking..."))
		}

		// Send to LLM
		resp, err := a.llm.Chat(a.history, toolDefs)
		if err != nil {
			if a.OnEmotion != nil {
				a.OnEmotion(GetMascotExpression(EmotionConfused, "Error thinking"))
			}
			return "", fmt.Errorf("LLM error: %w", err)
		}

		// If no native tool calls, check for text-embedded tool calls
		if len(resp.ToolCalls) == 0 {
			if extractedCall, ok := parseTextToolCall(resp.Content); ok {
				resp.ToolCalls = append(resp.ToolCalls, extractedCall)
			} else {
				a.history = append(a.history, types.Message{
					Role:    "assistant",
					Content: resp.Content,
				})

				step := TraceStep{
					Thought:  resp.Content,
					Response: resp.Content,
				}
				a.tracer.AddStep(step)
				if a.OnTrace != nil {
					a.OnTrace(step)
				}
				if a.OnEmotion != nil {
					a.OnEmotion(GetMascotExpression(EmotionHappy, "Done! 😎"))
				}

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
			if a.OnEmotion != nil {
				a.OnEmotion(GetMascotExpression(EmotionExecuting, fmt.Sprintf("Using %s...", tc.Name)))
			}
			if a.OnToolCall != nil {
				a.OnToolCall(tc.Name, tc.Arguments)
			}

			result, err := a.tools.Execute(tc.Name, tc.Arguments)
			if err != nil {
				result = fmt.Sprintf("Error executing tool %s: %v", tc.Name, err)
				if a.OnEmotion != nil {
					a.OnEmotion(GetMascotExpression(EmotionConfused, fmt.Sprintf("%s error", tc.Name)))
				}
			} else {
				if a.OnEmotion != nil {
					a.OnEmotion(GetMascotExpression(EmotionHappy, fmt.Sprintf("%s completed!", tc.Name)))
				}
			}

			if a.OnToolResult != nil {
				a.OnToolResult(tc.Name, result)
			}

			// Record tool trace step
			traceStep := TraceStep{
				Thought:   resp.Content,
				Tool:      tc.Name,
				Arguments: tc.Arguments,
				Result:    result,
			}
			a.tracer.AddStep(traceStep)
			if a.OnTrace != nil {
				a.OnTrace(traceStep)
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

var textToolCallRegexes = []*regexp.Regexp{
	// 1. Markdown code block or JSON tool_call: ```json {"tool": "desktop_apps", "arguments": {...}} ```
	regexp.MustCompile("(?:```(?:json)?\\s*)?\\{\\s*\"tool\"\\s*:\\s*\"([a-zA-Z0-9_]+)\"\\s*,\\s*\"arguments\"\\s*:\\s*(\\{[\\s\\S]*?\\})\\s*\\}"),

	// 2. XML tag format: <desktop_apps>{"action": "launch"}</desktop_apps>
	regexp.MustCompile(`<([a-zA-Z0-9_]+)>\s*(\{[\s\S]*?\})\s*</[a-zA-Z0-9_]+>`),

	// 3. Function tag format: <function=desktop_apps>{"action": "launch"}</function>
	regexp.MustCompile(`<function=([a-zA-Z0-9_]+)>\s*(\{[\s\S]*?\})(?:</function>)?`),

	// 4. Function call syntax: desktop_apps({"action": "launch"})
	regexp.MustCompile(`([a-zA-Z0-9_]+)\s*\(\s*(\{[\s\S]*?\})\s*\)`),
}

func parseTextToolCall(content string) (types.ToolCall, bool) {
	for _, re := range textToolCallRegexes {
		matches := re.FindStringSubmatch(content)
		if len(matches) >= 3 {
			toolName := strings.TrimSpace(matches[1])
			toolArgs := strings.TrimSpace(matches[2])

			// Clean up trailing brackets or markdown backticks if matched
			toolArgs = strings.TrimSuffix(toolArgs, "```")
			toolArgs = strings.TrimSpace(toolArgs)

			if strings.HasSuffix(toolArgs, "}}") && !strings.HasSuffix(toolArgs, "}}}") {
				toolArgs = strings.TrimSuffix(toolArgs, "}")
			}

			// Validate JSON
			var js map[string]interface{}
			if json.Unmarshal([]byte(toolArgs), &js) == nil {
				return types.ToolCall{
					ID:        fmt.Sprintf("call_text_%d", time.Now().UnixNano()),
					Name:      toolName,
					Arguments: toolArgs,
				}, true
			}
		}
	}
	return types.ToolCall{}, false
}
