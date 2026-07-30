package gui

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/heraji/jarvis/config"
	"github.com/heraji/jarvis/core"
	"github.com/heraji/jarvis/memory"
)

// Server handles the live HTTP & API backend for the NEXA GUI Dashboard.
type Server struct {
	agent    *core.Agent
	cfg      *config.Config
	memStore *memory.MemoryStore
	port     int
	mu       sync.Mutex
}

// NewServer initializes the GUI HTTP server.
func NewServer(agent *core.Agent, cfg *config.Config, memStore *memory.MemoryStore, port int) *Server {
	return &Server{
		agent:    agent,
		cfg:      cfg,
		memStore: memStore,
		port:     port,
	}
}

// Start launches the local HTTP backend and opens the desktop GUI window.
func (s *Server) Start() error {
	serverAddr := fmt.Sprintf("127.0.0.1:%d", s.port)
	appURL := fmt.Sprintf("http://%s", serverAddr)

	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("failed to bind GUI server to %s: %w", serverAddr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/api/memories", s.handleMemories)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("gui", "index.html"))
	})

	// Start HTTP server synchronously on bound listener
	go func() {
		_ = http.Serve(listener, mux)
	}()

	fmt.Printf("🌐 NEXA GUI Server running live at %s\n", appURL)
	fmt.Println("Press Ctrl+C in terminal to stop the GUI server.")

	// Launch Edge App Window & wait for process
	if runtime.GOOS == "windows" {
		cmd := exec.Command("msedge.exe", fmt.Sprintf("--app=%s", appURL), "--window-size=1280,820")
		if err := cmd.Start(); err == nil {
			return cmd.Wait()
		}
		fallbackCmd := exec.Command("cmd", "/c", "start", appURL)
		_ = fallbackCmd.Start()
	} else {
		_ = exec.Command("open", appURL).Start()
	}

	// Keep process alive if browser process detached
	select {}
}

type chatRequest struct {
	Message string `json:"message"`
}

type toolLog struct {
	Name string `json:"name"`
	Args string `json:"args"`
}

type chatResponse struct {
	Response  string            `json:"response"`
	ToolCalls []toolLog         `json:"tool_calls"`
	Memories  map[string]string `json:"memories"`
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Collect tool calls this request — chain with existing hooks instead of replacing them
	var usedTools []toolLog
	prevOnToolCall := s.agent.OnToolCall
	prevOnToolResult := s.agent.OnToolResult

	s.agent.OnToolCall = func(toolName string, args string) {
		usedTools = append(usedTools, toolLog{Name: toolName, Args: args})
		if prevOnToolCall != nil {
			prevOnToolCall(toolName, args)
		}
	}
	s.agent.OnToolResult = func(toolName string, result string) {
		if prevOnToolResult != nil {
			prevOnToolResult(toolName, result)
		}
	}

	respText, err := s.agent.Run(req.Message)

	// Restore original hooks
	s.agent.OnToolCall = prevOnToolCall
	s.agent.OnToolResult = prevOnToolResult

	if err != nil {
		respText = fmt.Sprintf("❌ Error: %v", err)
	}

	respPayload := chatResponse{
		Response:  respText,
		ToolCalls: usedTools,
		Memories:  s.memStore.List(),
	}

	_ = json.NewEncoder(w).Encode(respPayload)
}

func (s *Server) handleMemories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.memStore.List())
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"provider": s.cfg.LLMProvider,
		"model":    s.cfg.GroqModel,
		"tts":      s.cfg.EnableTTS,
		"version":  "1.0.0",
	}
	_ = json.NewEncoder(w).Encode(status)
}
