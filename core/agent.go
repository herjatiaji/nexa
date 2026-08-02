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

	// Track executed tool calls in this turn to prevent infinite loops
	executedToolCounts := make(map[string]int)

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
			if extractedCalls := parseTextToolCalls(resp.Content); len(extractedCalls) > 0 {
				resp.ToolCalls = append(resp.ToolCalls, extractedCalls...)
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
			toolKey := tc.Name + ":" + tc.Arguments
			executedToolCounts[toolKey]++

			var result string
			if executedToolCounts[toolKey] > 2 {
				// Prevent infinite tool execution loop by forcing the LLM to summarize
				result = fmt.Sprintf("System Notice: Tool '%s' with arguments '%s' was already executed %d times. Do NOT call this tool again. Please answer the user directly with the information already gathered.", tc.Name, tc.Arguments, executedToolCounts[toolKey]-1)
			} else {
				if a.OnEmotion != nil {
					a.OnEmotion(GetMascotExpression(EmotionExecuting, fmt.Sprintf("Using %s...", tc.Name)))
				}
				if a.OnToolCall != nil {
					a.OnToolCall(tc.Name, tc.Arguments)
				}

				res, err := a.tools.Execute(tc.Name, tc.Arguments)
				if err != nil {
					result = fmt.Sprintf("Error executing tool %s: %v", tc.Name, err)
					if a.OnEmotion != nil {
						a.OnEmotion(GetMascotExpression(EmotionConfused, fmt.Sprintf("%s error", tc.Name)))
					}
				} else {
					result = res
					if a.OnEmotion != nil {
						a.OnEmotion(GetMascotExpression(EmotionHappy, fmt.Sprintf("%s completed!", tc.Name)))
					}
				}

				if a.OnToolResult != nil {
					a.OnToolResult(tc.Name, result)
				}
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
	// 1. XML opening tag attribute format (flexible closing tag/bracket): <filesystem {"action": "list_dir", "path": "..."}}
	regexp.MustCompile(`<([a-zA-Z0-9_]+)\s+(\{[\s\S]*?\})(?:</[a-zA-Z0-9_]+>|/>|>|\s|$)`),

	// 2. XML tag wrapper format: <desktop_apps>{"action": "launch"}</desktop_apps>
	regexp.MustCompile(`<([a-zA-Z0-9_]+)>\s*(\{[\s\S]*?\})\s*</[a-zA-Z0-9_]+>`),

	// 3. Markdown code block or JSON tool_call: ```json {"tool": "desktop_apps", "arguments": {...}} ``` or {"name": "...", "parameters": {...}}
	regexp.MustCompile("(?:```(?:json)?\\s*)?\\{\\s*\"(?:tool|name)\"\\s*:\\s*\"([a-zA-Z0-9_]+)\"\\s*,\\s*\"(?:arguments|parameters|args)\"\\s*:\\s*(\\{[\\s\\S]*?\\})\\s*\\}"),

	// 4. Function tag format: <function=desktop_apps>{"action": "launch"}</function>
	regexp.MustCompile(`<function=([a-zA-Z0-9_]+)>\s*(\{[\s\S]*?\})(?:</function>)?`),

	// 5. Function call syntax: desktop_apps({"action": "launch"})
	regexp.MustCompile(`([a-zA-Z0-9_]+)\s*\(\s*(\{[\s\S]*?\})\s*\)`),
}

func parseTextToolCalls(content string) []types.ToolCall {
	var toolCalls []types.ToolCall
	seen := make(map[string]bool)

	for _, re := range textToolCallRegexes {
		matches := re.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) >= 3 {
				toolName := strings.TrimSpace(m[1])
				rawArgs := strings.TrimSpace(m[2])

				// Clean up trailing markdown backticks
				rawArgs = strings.TrimSuffix(rawArgs, "```")
				rawArgs = strings.TrimSpace(rawArgs)

				// Iteratively fix extra trailing braces (e.g. }} produced by some LLMs)
				currArgs := rawArgs
				var validArgs string
				var validJSON map[string]interface{}

				for len(currArgs) > 0 {
					var js map[string]interface{}
					if json.Unmarshal([]byte(currArgs), &js) == nil {
						validArgs = currArgs
						validJSON = js
						break
					}
					// If json unmarshal failed and it ends with '}', try trimming one trailing '}'
					if strings.HasSuffix(currArgs, "}") {
						currArgs = strings.TrimSuffix(currArgs, "}")
						currArgs = strings.TrimSpace(currArgs)
					} else {
						break
					}
				}

				if validArgs != "" && validJSON != nil {
					key := toolName + ":" + validArgs
					if !seen[key] {
						seen[key] = true
						toolCalls = append(toolCalls, types.ToolCall{
							ID:        fmt.Sprintf("call_text_%d_%d", time.Now().UnixNano(), len(toolCalls)),
							Name:      toolName,
							Arguments: validArgs,
						})
					}
				}
			}
		}
	}

	// Also parse XML attribute syntax: <filesystem action="list_dir" path="..."></filesystem>
	xmlCalls := parseXMLAttributeToolCalls(content)
	for _, xc := range xmlCalls {
		key := xc.Name + ":" + xc.Arguments
		if !seen[key] {
			seen[key] = true
			toolCalls = append(toolCalls, xc)
		}
	}

	return toolCalls
}

var xmlTagRegex = regexp.MustCompile(`<([a-zA-Z0-9_]+)\s+([^>]+)>(?:</[a-zA-Z0-9_]+>|/>)?`)
var xmlAttrRegex = regexp.MustCompile(`([a-zA-Z0-9_]+)=(?:"([^"]*)"|'([^']*)')`)

func parseXMLAttributeToolCalls(content string) []types.ToolCall {
	var toolCalls []types.ToolCall

	// Normalize double closing brackets like ">>" -> ">"
	cleanedContent := strings.ReplaceAll(content, ">>", ">")

	matches := xmlTagRegex.FindAllStringSubmatch(cleanedContent, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			toolName := strings.TrimSpace(m[1])
			attrStr := strings.TrimSpace(m[2])

			// Skip if attrStr starts with '{' (handled by JSON parser)
			if strings.HasPrefix(attrStr, "{") {
				continue
			}

			attrMatches := xmlAttrRegex.FindAllStringSubmatch(attrStr, -1)
			if len(attrMatches) == 0 {
				continue
			}

			argsMap := make(map[string]interface{})
			for _, am := range attrMatches {
				if len(am) >= 2 {
					k := am[1]
					v := ""
					if len(am) >= 3 {
						v = am[2]
					}
					if v == "" && len(am) >= 4 {
						v = am[3]
					}
					argsMap[k] = v
				}
			}

			if len(argsMap) > 0 {
				jsonBytes, err := json.Marshal(argsMap)
				if err == nil {
					toolCalls = append(toolCalls, types.ToolCall{
						ID:        fmt.Sprintf("call_xml_%d_%d", time.Now().UnixNano(), len(toolCalls)),
						Name:      toolName,
						Arguments: string(jsonBytes),
					})
				}
			}
		}
	}

	return toolCalls
}
