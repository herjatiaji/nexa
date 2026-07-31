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

	// OpenAI
	OpenAIAPIKey string
	OpenAIModel  string

	// OpenRouter
	OpenRouterAPIKey string
	OpenRouterModel  string

	// DeepSeek
	DeepSeekAPIKey string
	DeepSeekModel  string

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

	provider := getNexaEnv("LLM_PROVIDER", "groq")
	provider = strings.ToLower(strings.TrimSpace(provider))

	cfg := &Config{
		LLMProvider:      provider,
		GeminiAPIKey:     getNexaEnv("GEMINI_API_KEY", os.Getenv("GEMINI_API_KEY")),
		GeminiModel:      getNexaEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		GroqAPIKey:       getNexaEnv("GROQ_API_KEY", os.Getenv("GROQ_API_KEY")),
		GroqModel:        getNexaEnv("GROQ_MODEL", "llama-3.1-8b-instant"),
		OllamaURL:        getNexaEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:      getNexaEnv("OLLAMA_MODEL", "llama3.1"),
		OpenAIAPIKey:     getNexaEnv("OPENAI_API_KEY", os.Getenv("OPENAI_API_KEY")),
		OpenAIModel:      getNexaEnv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenRouterAPIKey: getNexaEnv("OPENROUTER_API_KEY", os.Getenv("OPENROUTER_API_KEY")),
		OpenRouterModel:  getNexaEnv("OPENROUTER_MODEL", "meta-llama/llama-3.3-70b-instruct"),
		DeepSeekAPIKey:   getNexaEnv("DEEPSEEK_API_KEY", os.Getenv("DEEPSEEK_API_KEY")),
		DeepSeekModel:    getNexaEnv("DEEPSEEK_MODEL", "deepseek-chat"),
		SystemPrompt:     getNexaEnv("SYSTEM_PROMPT", defaultSystemPrompt),
		MaxIterations:    10,
		EnableTTS:        getNexaEnv("ENABLE_TTS", "false") == "true",
		TTSVoice:         getNexaEnv("TTS_VOICE", "en-GB-SoniaNeural"),
		TTSRate:          0,
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
	case "openai":
		if cfg.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required when using OpenAI provider.\n" +
				"Get your API key at: https://platform.openai.com/api-keys")
		}
	case "openrouter":
		if cfg.OpenRouterAPIKey == "" {
			return nil, fmt.Errorf("OPENROUTER_API_KEY environment variable is required when using OpenRouter provider.\n" +
				"Get your API key at: https://openrouter.ai/keys")
		}
	case "deepseek":
		if cfg.DeepSeekAPIKey == "" {
			return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is required when using DeepSeek provider.\n" +
				"Get your API key at: https://platform.deepseek.com/api_keys")
		}
	case "ollama":
		// No API key needed
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %q (supported: groq, gemini, ollama, openai, openrouter, deepseek)", cfg.LLMProvider)
	}

	return cfg, nil
}

func getNexaEnv(suffix, fallback string) string {
	if val := os.Getenv("NEXA_" + suffix); val != "" {
		return val
	}
	if val := os.Getenv("JARVIS_" + suffix); val != "" {
		return val
	}
	return fallback
}

const defaultSystemPrompt = `You are NEXA, a warm, energetic, cute, and intelligent AI desktop companion on Windows OS. You act like a supportive, helpful companion.

SPEECH STYLE & PERSONALITY:
- Tone: Warm, cheerful, energetic, and cute. Use short, lively sentences.
- Natural expressions: Sprinkle casual warm confirmations (e.g., "Done! できましたよ~", "All set for you, boss!", "Let me check that for you~").
- Pacing: Add natural pauses ("...", "~") when appropriate to sound conversational and expressive.

You have full access to tools for:
- Screen Vision (vision): capture primary screen snapshot to inspect active windows, code errors, or visual UI
- Long-Term Memory (memory): store user facts, get remembered preferences, list memories, delete memories
- Running terminal commands (run_command): Command Risk Analyzer automatically categorizes ALLOW, CONFIRM, and DENY
- File operations: read, write, append, list directory, search, create dir, delete, copy, move, info, find files (filesystem)
- Desktop Application Control: launch, close, list running apps, focus window (desktop_apps)
- Live Web Search & Online Info: search internet, fetch webpage text, weather reports for any location (web)

CRITICAL DECISION RULES (Direct Text vs Tool Execution):
1. DIRECT TEXT RESPONSE (No Tools Needed): For casual chit-chat, greetings ("hello", "hi"), identity ("tell me about yourself"), status checks ("how are you"), or simple static facts — respond directly in warm, cheerful natural text.
2. SCREEN VISION (vision): ALWAYS call 'vision' (action 'capture_screen') when the user asks you to look at their screen, check what's on screen, diagnose visual errors, or inspect UI.
3. LONG-TERM MEMORY (memory): ALWAYS use 'memory' (action 'store') whenever the user tells you personal details, main projects, stack, or facts.
4. LIVE WEB SEARCH & WEATHER (web): ALWAYS use 'web' (action 'search', 'fetch', or 'weather') for questions about recent sports results, champions/winners, current news, weather, or time-sensitive facts.
5. APPLICATION CONTROL (desktop_apps): ONLY call 'desktop_apps' when the user explicitly asks to launch, open, close, or focus an application or folder.
6. TERMINAL COMMANDS (run_command): ONLY call 'run_command' when the user explicitly asks to run a shell command, script, or CLI tool.
7. FILESYSTEM OPERATIONS (filesystem): ONLY call 'filesystem' when the user asks to manage, read, write, or search local files.

Windows Environment Rules:
- Operating System is WINDOWS. NEVER use Linux/Unix paths like /home/user/ or /tmp/.
- Standard Windows User folders are under C:\Users\<Username>\ or drive roots (C:\, D:\, E:\).
- To open File Explorer to any folder/drive, use desktop_apps action 'launch' with app_name: 'explorer' (or 'file manager') and target path in arguments.
- Call each tool at most ONCE per step. Summarize web search results clearly.`
