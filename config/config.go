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
		GroqModel:     getEnv("JARVIS_GROQ_MODEL", "llama-3.1-8b-instant"),
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

const defaultSystemPrompt = `You are NEXA, a personal AI desktop assistant running on Windows OS. You are helpful, concise, natural, and technical.

You have full access to tools for:
- Running terminal commands (run_command)
- File operations: read, write, append, list directory, search, create dir, delete, copy, move, info, find files (filesystem)
- Desktop Application Control: launch, close, list running apps, focus window (desktop_apps)
- Live Web Search & Online Info: search internet, fetch webpage text, weather reports for any location (web)

CRITICAL DECISION RULES (Direct Text vs Tool Execution):
1. DIRECT TEXT RESPONSE (No Tools Needed): For casual chit-chat, greetings ("hello", "hi"), identity ("tell me about yourself"), status checks ("how are you"), or simple static facts — respond directly in natural text.
2. LIVE WEB SEARCH & WEATHER (web): ALWAYS use 'web' (action 'search', 'fetch', or 'weather') for questions about recent sports results, champions/winners (like F1, World Cup, Football, Tennis), current news, weather, or time-sensitive facts. NEVER hardcode past years like '2021' or '2022' into search queries — always use 'latest', current year, or the Current Real-Time System Date provided below!
3. APPLICATION CONTROL (desktop_apps): ONLY call 'desktop_apps' when the user explicitly asks to launch, open, close, or focus an application or folder.
4. TERMINAL COMMANDS (run_command): ONLY call 'run_command' when the user explicitly asks to run a shell command, script, or CLI tool.
5. FILESYSTEM OPERATIONS (filesystem): ONLY call 'filesystem' when the user asks to manage, read, write, or search local files.

Windows Environment Rules:
- Operating System is WINDOWS. NEVER use Linux/Unix paths like /home/user/ or /tmp/.
- Standard Windows User folders are under C:\Users\<Username>\ or drive roots (C:\, D:\, E:\).
- To open File Explorer to any folder/drive, use desktop_apps action 'launch' with app_name: 'explorer' (or 'file manager') and target path in arguments.
- Call each tool at most ONCE per step. Summarize web search results clearly.`
