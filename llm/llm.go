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
	case "openai":
		return NewOpenAI(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	case "openrouter":
		return NewOpenRouter(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
	case "deepseek":
		return NewDeepSeek(cfg.DeepSeekAPIKey, cfg.DeepSeekModel)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %q (supported: groq, gemini, ollama, openai, openrouter, deepseek)", cfg.LLMProvider)
	}
}
