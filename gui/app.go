package gui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/heraji/jarvis/brain"
	"github.com/heraji/jarvis/cognitive/explain"
	"github.com/heraji/jarvis/cognitive/trace"
	"github.com/heraji/jarvis/config"
	"github.com/heraji/jarvis/core"
	"github.com/heraji/jarvis/events"
	"github.com/heraji/jarvis/memory"
	"github.com/heraji/jarvis/observer"
	runtimepkg "github.com/heraji/jarvis/runtime"
	"github.com/heraji/jarvis/security"
	"github.com/heraji/jarvis/voice"
	"github.com/heraji/jarvis/voice/speech"
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
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	TTS          bool   `json:"tts"`
	VoiceActive  bool   `json:"voiceActive"`
	Version      string `json:"version"`
}

// App is the Wails application backend struct.
type App struct {
	agent           *core.Agent
	cfg             *config.Config
	memStore        *memory.MemoryStore
	ctx             context.Context
	mu              sync.Mutex
	voiceActive     bool
	voiceCleanup    func()
	ttsEngine       *voice.TTS
	speechPlanner   *speech.SpeechPlanner
	eventBus        *events.EventBus
	desktopObserver *observer.DesktopObserver
	awareness       *observer.AwarenessManager
	brainKernel     *brain.Brain
}

// NewApp creates a new App instance with the NEXA agent, config, and memory store.
func NewApp(agent *core.Agent, cfg *config.Config, memStore *memory.MemoryStore) *App {
	tts := voice.NewTTS(cfg.TTSVoice, cfg.TTSRate, cfg.EnableTTS)
	sp := speech.NewSpeechPlanner(speech.DefaultPersonality())
	eb := events.NewEventBus()
	obs := observer.NewDesktopObserver(eb)
	aw := observer.NewAwarenessManager(eb)
	bk := brain.NewBrain(eb)

	app := &App{
		agent:           agent,
		cfg:             cfg,
		memStore:        memStore,
		ttsEngine:       tts,
		speechPlanner:   sp,
		eventBus:        eb,
		desktopObserver: obs,
		awareness:       aw,
		brainKernel:     bk,
	}

	// Forward all EventBus events to Wails frontend
	eb.Subscribe("*", func(evt events.Event) {
		if app.ctx != nil {
			runtime.EventsEmit(app.ctx, string(evt.Type), evt.Payload)
		}
	})

	return app
}

// Startup is called by Wails when the application starts. Stores the runtime context.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	if a.desktopObserver != nil {
		a.desktopObserver.Start()
	}
	if a.brainKernel != nil {
		a.brainKernel.Start(ctx)
	}
}

// Chat sends a message to the NEXA agent and returns the full response with tool logs.
func (a *App) Chat(message string) ChatResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	runtime.EventsEmit(a.ctx, "nexa:status", "thinking")

	var usedTools []ToolCallLog
	prevOnToolCall := a.agent.OnToolCall
	prevOnToolResult := a.agent.OnToolResult
	prevOnEmotion := a.agent.OnEmotion

	a.agent.OnEmotion = func(mascot core.MascotState) {
		runtime.EventsEmit(a.ctx, "nexa:emotion", mascot)
		if prevOnEmotion != nil {
			prevOnEmotion(mascot)
		}
	}

	a.agent.OnToolCall = func(toolName string, args string) {
		usedTools = append(usedTools, ToolCallLog{Name: toolName, Args: args})
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

	a.agent.OnToolCall = prevOnToolCall
	a.agent.OnToolResult = prevOnToolResult
	a.agent.OnEmotion = prevOnEmotion

	runtime.EventsEmit(a.ctx, "nexa:status", "ready")

	if err != nil {
		respText = fmt.Sprintf("❌ Error: %v", err)
	} else if a.cfg.EnableTTS && a.ttsEngine != nil && a.speechPlanner != nil {
		// Use Speech Intelligence Layer for emotion-aware TTS
		go a.speakWithIntelligence(respText)
	}

	return ChatResult{
		Response:  respText,
		ToolCalls: usedTools,
		Memories:  a.memStore.List(),
	}
}

// speakWithIntelligence runs the full Speech Intelligence pipeline:
// Text → Segmentation → Emotion Analysis → Prosody Mapping → TTS Synthesis → Post-processing → Playback.
// Emits nexa:speech:plan and nexa:emotion events for frontend visualization.
func (a *App) speakWithIntelligence(text string) {
	// Set up emotion detection callback to sync mascot state
	a.speechPlanner.OnEmotionDetected = func(emotion speech.EmotionTag, confidence float64) {
		mascotState := core.FromSpeechEmotion(string(emotion), fmt.Sprintf("Speaking (%s)", emotion))
		runtime.EventsEmit(a.ctx, "nexa:emotion", mascotState)
	}

	// Plan the speech (analyze emotion, map prosody for each chunk)
	plan := a.speechPlanner.Plan(text)

	// Emit speech plan to frontend for visualization
	runtime.EventsEmit(a.ctx, "nexa:speech:plan", map[string]interface{}{
		"chunks":          plan.Chunks,
		"dominantEmotion": plan.DominantEmotion,
		"confidence":      plan.Confidence,
		"personality":     plan.Personality.Name,
	})

	// Emit speaking emotion
	runtime.EventsEmit(a.ctx, "nexa:emotion", core.GetMascotExpression(core.EmotionSpeaking, "Speaking..."))

	// Execute synthesis with prosody-tuned TTS
	wavPath, err := a.speechPlanner.Execute(plan, func(chunkText string, prosody speech.ProsodyParams) (string, error) {
		return a.ttsEngine.SynthesizeToFile(chunkText, prosody)
	})

	if err != nil {
		// Fallback to standard flat TTS if speech pipeline fails
		_ = a.ttsEngine.Speak(text)
		return
	}

	// Play the final processed audio
	_ = voice.PlayWAVFile(wavPath)
	_ = os.Remove(wavPath)

	// Emit idle emotion after speech completes
	runtime.EventsEmit(a.ctx, "nexa:emotion", core.GetMascotExpression(core.EmotionIdle, ""))
}

// StartVoiceEngine launches openWakeWord + Whisper STT listening background goroutine.
func (a *App) StartVoiceEngine() string {
	a.mu.Lock()
	if a.voiceActive {
		a.mu.Unlock()
		return "Voice engine already active"
	}

	resultChan, cleanup, err := voice.StartPersistentListener()
	if err != nil {
		a.mu.Unlock()
		return fmt.Sprintf("Failed to start voice engine: %v", err)
	}

	a.voiceActive = true
	a.voiceCleanup = cleanup
	a.mu.Unlock()

	runtime.EventsEmit(a.ctx, "nexa:voice:state", true)
	if a.cfg.EnableTTS && a.ttsEngine != nil {
		a.ttsEngine.SpeakAsync("Voice activation online. At your service, sir.")
	}

	go a.listenVoiceLoop(resultChan)
	return "OK"
}

// StopVoiceEngine stops the voice listener background process.
func (a *App) StopVoiceEngine() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.voiceActive {
		return "Voice engine is not active"
	}

	if a.voiceCleanup != nil {
		a.voiceCleanup()
		a.voiceCleanup = nil
	}
	a.voiceActive = false

	runtime.EventsEmit(a.ctx, "nexa:voice:state", false)
	return "OK"
}

func (a *App) listenVoiceLoop(resultChan <-chan string) {
	whisperEngine := voice.NewWhisperSTT(a.cfg.GroqAPIKey)

	for result := range resultChan {
		if strings.HasPrefix(result, "WAKE:") {
			parts := strings.Split(result, ":")
			wakeWord := "Hey Nexa"
			if len(parts) >= 2 {
				wakeWord = parts[1]
			}

			runtime.EventsEmit(a.ctx, "nexa:status", "listening")
			runtime.EventsEmit(a.ctx, "nexa:voice:wake", wakeWord)

			_ = voice.SendListenCommand("PAUSE")
			if a.cfg.EnableTTS && a.ttsEngine != nil {
				a.ttsEngine.Speak("Yes, sir?")
			}

		} else if strings.HasPrefix(result, "COMMAND_WAV:") {
			wavFile := strings.TrimPrefix(result, "COMMAND_WAV:")
			userCommand, err := whisperEngine.Transcribe(wavFile)
			_ = os.Remove(wavFile)

			userCommand = strings.TrimSpace(userCommand)
			cleanCmd := strings.Trim(userCommand, ".?!, ")
			cleanLower := strings.ToLower(cleanCmd)

			if err != nil || cleanCmd == "" || cleanLower == "thank you" || cleanLower == "thanks" || cleanLower == "thanks for watching" || cleanLower == "you" {
				_ = voice.SendListenCommand("RESUME")
				runtime.EventsEmit(a.ctx, "nexa:status", "ready")
				continue
			}

			// Broadcast voice command to frontend and run chat
			runtime.EventsEmit(a.ctx, "nexa:voice:command", userCommand)

			chatRes := a.Chat(userCommand)

			runtime.EventsEmit(a.ctx, "nexa:voice:result", chatRes)

			// Resume listening
			_ = voice.SendListenCommand("RESUME")
			runtime.EventsEmit(a.ctx, "nexa:status", "ready")
		}
	}
}

// GetTraces returns all recorded agent reasoning trace steps.
func (a *App) GetTraces() []core.TraceStep {
	return a.agent.GetTraces()
}

// CreatePlan decomposes a multi-step user goal into structured sub-tasks.
func (a *App) CreatePlan(prompt string) (*core.Plan, error) {
	return a.agent.CreatePlan(prompt)
}

// SearchMemories performs semantic concept matching over memories.
func (a *App) SearchMemories(query string) []memory.SemanticMatch {
	return a.memStore.SearchSemantic(query)
}

// GetMemories returns all stored long-term memories.
func (a *App) GetMemories() map[string]string {
	return a.memStore.List()
}

// DeleteMemory removes a specific key from the memory store.
func (a *App) DeleteMemory(key string) error {
	return a.memStore.Delete(key)
}

// GetPermissions returns current OS permission settings.
func (a *App) GetPermissions() map[string]string {
	permMap := make(map[string]string)
	for k, v := range security.NewPermissionManager().ListPermissions() {
		permMap[k] = string(v)
	}
	return permMap
}

// SetPermission updates an OS permission rule.
func (a *App) SetPermission(capability string, level string) error {
	pm := security.NewPermissionManager()
	return pm.SetPermission(capability, security.PermissionLevel(level))
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
		Provider:    a.cfg.LLMProvider,
		Model:       model,
		TTS:         a.cfg.EnableTTS,
		VoiceActive: a.voiceActive,
		Version:     "1.4.0",
	}
}

// ResetConversation clears the agent's conversation history.
func (a *App) ResetConversation() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.agent.Reset()
}

// GetBrainState returns a snapshot of current brain state.
func (a *App) GetBrainState() brain.BrainSnapshot {
	if a.brainKernel == nil {
		return brain.BrainSnapshot{}
	}
	return a.brainKernel.Snapshot()
}

// GetCognitiveTraces returns recent reasoning and perception traces.
func (a *App) GetCognitiveTraces(limit int) []trace.CognitiveTrace {
	if a.brainKernel == nil {
		return nil
	}
	if limit <= 0 {
		limit = 50
	}
	return a.brainKernel.GetTraces(limit)
}

// GetBrainMetrics returns system-wide cognitive performance metrics.
func (a *App) GetBrainMetrics() runtimepkg.BrainMetricsTelemetry {
	if a.brainKernel == nil {
		return runtimepkg.BrainMetricsTelemetry{}
	}
	return a.brainKernel.GetMetrics()
}

// ExplainLastDecision returns human-readable natural language explanation of the latest decision.
func (a *App) ExplainLastDecision() explain.DecisionExplanation {
	if a.brainKernel == nil {
		return explain.DecisionExplanation{}
	}
	return a.brainKernel.ExplainLastDecision()
}
