package gui

import (
	"context"
	"fmt"
	"sync"

	"github.com/heraji/jarvis/config"
	"github.com/heraji/jarvis/core"
	"github.com/heraji/jarvis/memory"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ToolCallLog captures a single tool invocation for the frontend.
type ToolCallLog struct {
	Name string `json:"name"`
	Args string `json:"args"`
}

// ChatResult is the response payload returned to the React frontend.
type ChatResult struct {
	Response  string            `json:"response"`
	ToolCalls []ToolCallLog     `json:"toolCalls"`
	Memories  map[string]string `json:"memories"`
}

// StatusInfo holds system status displayed in the top bar.
type StatusInfo struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	TTS      bool   `json:"tts"`
	Version  string `json:"version"`
}

// App is the Wails application backend struct.
// All exported methods are automatically bound and callable from React/TypeScript.
type App struct {
	agent    *core.Agent
	cfg      *config.Config
	memStore *memory.MemoryStore
	ctx      context.Context
	mu       sync.Mutex
}

// NewApp creates a new App instance with the NEXA agent, config, and memory store.
func NewApp(agent *core.Agent, cfg *config.Config, memStore *memory.MemoryStore) *App {
	return &App{
		agent:    agent,
		cfg:      cfg,
		memStore: memStore,
	}
}

// Startup is called by Wails when the application starts. Stores the runtime context.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Chat sends a message to the NEXA agent and returns the full response with tool logs.
// Called from React: const result = await Chat("open spotify");
func (a *App) Chat(message string) ChatResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Emit "thinking" event to frontend
	runtime.EventsEmit(a.ctx, "nexa:status", "thinking")

	// Capture tool calls for this request
	var usedTools []ToolCallLog
	prevOnToolCall := a.agent.OnToolCall
	prevOnToolResult := a.agent.OnToolResult

	a.agent.OnToolCall = func(toolName string, args string) {
		usedTools = append(usedTools, ToolCallLog{Name: toolName, Args: args})
		// Emit live tool event to frontend
		runtime.EventsEmit(a.ctx, "nexa:tool", map[string]string{"name": toolName, "args": args})
		if prevOnToolCall != nil {
			prevOnToolCall(toolName, args)
		}
	}
	a.agent.OnToolResult = func(toolName string, result string) {
		if prevOnToolResult != nil {
			prevOnToolResult(toolName, result)
		}
	}

	respText, err := a.agent.Run(message)

	// Restore original hooks
	a.agent.OnToolCall = prevOnToolCall
	a.agent.OnToolResult = prevOnToolResult

	// Emit "ready" event
	runtime.EventsEmit(a.ctx, "nexa:status", "ready")

	if err != nil {
		respText = fmt.Sprintf("❌ Error: %v", err)
	}

	return ChatResult{
		Response:  respText,
		ToolCalls: usedTools,
		Memories:  a.memStore.List(),
	}
}

// GetMemories returns all stored long-term memories.
func (a *App) GetMemories() map[string]string {
	return a.memStore.List()
}

// DeleteMemory removes a specific key from the memory store.
func (a *App) DeleteMemory(key string) error {
	return a.memStore.Delete(key)
}

// GetStatus returns current LLM provider, model, TTS state, and version.
func (a *App) GetStatus() StatusInfo {
	model := a.cfg.GroqModel
	switch a.cfg.LLMProvider {
	case "gemini":
		model = a.cfg.GeminiModel
	case "ollama":
		model = a.cfg.OllamaModel
	case "openai":
		model = a.cfg.OpenAIModel
	case "openrouter":
		model = a.cfg.OpenRouterModel
	case "deepseek":
		model = a.cfg.DeepSeekModel
	}
	return StatusInfo{
		Provider: a.cfg.LLMProvider,
		Model:    model,
		TTS:      a.cfg.EnableTTS,
		Version:  "1.2.0",
	}
}

// ResetConversation clears the agent's conversation history.
func (a *App) ResetConversation() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.agent.Reset()
}
