package llm

import (
	"fmt"

	"github.com/heraji/jarvis/config"
	"github.com/heraji/jarvis/types"
)

// LLM defines the interface for interacting with language models.
type LLM interface {
	// Chat sends a conversation history and available tools to the LLM,
	// returning either a text response or tool call requests.
	Chat(messages []types.Message, tools []types.ToolDefinition) (*types.ChatResponse, error)
}

// New creates a new LLM provider based on the config.
func New(cfg *config.Config) (LLM, error) {
	switch cfg.LLMProvider {
	case "gemini":
		return NewGemini(cfg.GeminiAPIKey, cfg.GeminiModel)
	case "groq":
		return NewGroq(cfg.GroqAPIKey, cfg.GroqModel)
	case "ollama":
		return NewOllama(cfg.OllamaURL, cfg.OllamaModel)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.LLMProvider)
	}
}
