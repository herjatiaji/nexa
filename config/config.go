package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the JARVIS application.
type Config struct {
	// LLM Provider: "gemini", "groq", "ollama"
	LLMProvider string

	// Google Gemini
	GeminiAPIKey string
	GeminiModel  string

	// Groq
	GroqAPIKey string
	GroqModel  string

	// Ollama
	OllamaURL   string
	OllamaModel string

	// Agent
	SystemPrompt  string
	MaxIterations int

	// Voice (TTS)
	EnableTTS bool
	TTSVoice  string
	TTSRate   int
}

// LoadConfig reads configuration from .env file and environment variables.
func LoadConfig() (*Config, error) {
	// Try loading .env from current working directory
	_ = godotenv.Load()

	// Also try loading .env from the executable's directory (for double-click in Explorer)
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		_ = godotenv.Load(filepath.Join(exeDir, ".env"))
	}

	provider := getEnv("JARVIS_LLM_PROVIDER", "gemini")
	provider = strings.ToLower(strings.TrimSpace(provider))

	cfg := &Config{
		LLMProvider:   provider,
		GeminiAPIKey:  os.Getenv("GEMINI_API_KEY"),
		GeminiModel:   getEnv("JARVIS_GEMINI_MODEL", "gemini-2.0-flash"),
		GroqAPIKey:    os.Getenv("GROQ_API_KEY"),
		GroqModel:     getEnv("JARVIS_GROQ_MODEL", "llama-3.3-70b-versatile"),
		OllamaURL:     getEnv("JARVIS_OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:   getEnv("JARVIS_OLLAMA_MODEL", "llama3.1"),
		SystemPrompt:  getEnv("JARVIS_SYSTEM_PROMPT", defaultSystemPrompt),
		MaxIterations: 10,
		EnableTTS:     getEnv("JARVIS_ENABLE_TTS", "false") == "true",
		TTSVoice:      getEnv("JARVIS_TTS_VOICE", "en-GB-SoniaNeural"),
		TTSRate:       0,
	}

	// Validate provider-specific requirements
	switch cfg.LLMProvider {
	case "gemini":
		if cfg.GeminiAPIKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required when using Gemini provider.\n" +
				"Get your free API key at: https://aistudio.google.com/apikey")
		}
	case "groq":
		if cfg.GroqAPIKey == "" {
			return nil, fmt.Errorf("GROQ_API_KEY environment variable is required when using Groq provider.\n" +
				"Get your free API key at: https://console.groq.com/keys")
		}
	case "ollama":
		// No API key needed, just check URL is set
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %q (supported: gemini, groq, ollama)", cfg.LLMProvider)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

const defaultSystemPrompt = `You are JARVIS, a personal AI desktop assistant. You are helpful, concise, and technical.

You have access to tools for:
- Running terminal commands (run_command)
- File operations: read, write, append, list directory, search, create dir, delete, copy, move, info, find files (filesystem)
- Desktop Application Control: launch, close, list running apps, focus window (desktop_apps)

Rules:
1. Use tools when the user needs system interaction. Give a text answer when no tool is needed.
2. Be concise. Don't repeat tool results verbatim — summarize them.
3. Always communicate in English.
4. ALWAYS use ABSOLUTE paths for filesystem operations.
5. When user says "this project" or "current directory", use the working directory below.
6. Call each tool only ONCE per step. After receiving a tool result, respond with your final answer as text.
7. Do NOT call the same tool again with the same arguments.
8. If something could be dangerous, warn the user first.`
