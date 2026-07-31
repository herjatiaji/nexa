# 🤖 NEXA — Personal AI Desktop & Voice Assistant

[![Release](https://img.shields.io/badge/Version-v1.2.0-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react)](https://reactjs.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v1.2.0** is an ultra-fast, hands-free personal AI desktop assistant built in Go, Wails v2, React, and Python. It features a **Native Windows Desktop Application** (no browser/localhost needed), interactive Voice Activation (`🎤 START VOICE`), multi-LLM provider support (Groq, OpenAI, DeepSeek, OpenRouter, Gemini, Ollama), System Tray integration, `Ctrl+Space` global hotkey, real-time neural wake word detection, OpenAI Whisper speech-to-text, Rhasspy Piper neural voice synthesis, live web searching, persistent long-term memory, a 3-tier command risk analyzer, screen vision, and Model Context Protocol (MCP) support.

---

## ✨ Key Features in v1.2.0

- 🖥️ **Native Windows Desktop App (Wails v2 + React 18 + TS)**  
  Standalone native desktop application using Wails v2 with direct Go↔JS bindings (no HTTP API, no localhost browser required). Features a dark glassmorphism 3-column layout.

- 🎙️ **Desktop GUI Voice Activation Mode (`🎤 START VOICE`)**  
  Directly toggle hands-free voice command listening inside the GUI desktop app. Say **"Hey Nexa"** or **"Nexa"** to activate, speak your command, and watch user speech & NEXA responses render live in the conversation stream!

- 🔔 **Windows System Tray Integration**  
  Runs cleanly in the background with a native Windows system tray icon for status monitoring.

- ⌨️ **Global Hotkey (`Ctrl+Space`)**  
  Instantly toggle/bring NEXA to the foreground from anywhere in Windows with `Ctrl+Space`.

- 🤖 **Multi-LLM Provider Engine**  
  Effortlessly switch between 6 major LLM providers:
  - ⚡ **Groq** (`llama-3.1-8b-instant`, `llama-3.3-70b-versatile`, `gemma2-9b-it`)
  - 🧠 **OpenAI** (`gpt-4o`, `gpt-4o-mini`, `o3-mini`)
  - 🌐 **OpenRouter** (`meta-llama/llama-3.3-70b-instruct`, `anthropic/claude-3.5-sonnet`, `deepseek/deepseek-r1`)
  - 🔬 **DeepSeek** (`deepseek-chat`, `deepseek-reasoner`)
  - ♊ **Google Gemini** (`gemini-2.0-flash`, `gemini-1.5-pro`)
  - 🦙 **Ollama (Local)** (`llama3.1`, `qwen2.5`, `mistral`)

- 🗣️ **Neural Voice Synthesis (Piper TTS)**  
  High-fidelity offline British female voice output using `en_GB-alba-medium.onnx` via Rhasspy Piper.

- 🧠 **Persistent Long-Term Memory Store (`nexa_memory.json`)**  
  Remembers user preferences, project tech stacks, and context across restarts using the `memory` tool (`store`, `get`, `list`, `delete`). Features live deletion (`✕`) and memory inspection directly in the desktop app!

- 🛡️ **3-Tier Command Risk Analyzer**  
  Categorizes shell commands into `ALLOW` (safe read-only), `CONFIRM` (user prompt), and `DENY` (destructive wipe blocked).

- 👁️ **Screen Vision Tool**  
  Captures real-time Windows primary screen snapshots (Base64 PNG) for screen-aware assistance.

- 🔌 **Model Context Protocol (MCP) Client**  
  JSON-RPC 2.0 stdio transport client supporting Model Context Protocol for connecting to external MCP tool servers.

---

## 🏗️ System Architecture

```
                 +───────────────────────────────────+
                 │   NEXA Native Windows App        │
                 │   (Wails v2 + React 18 + TS)      │
                 +─────────────────┬─────────────────+
                                   │
                         Wails Go↔JS Bindings
                         (Direct In-Process IPC)
                                   │
            ┌──────────────────────┴──────────────────────┐
            │                                             │
     React Frontend                                 Go Agent Runtime
     • ChatStream (with tool badges)                • core/agent.go
     • VoiceReactor (animated orb + 🎤 Voice ON)     • llm/ (6 Providers)
     • MemoryPanel (with live delete)               • memory/
     • StatusBar (provider & model info)            • tools/ (7 Tools)
     • System Tray & Hotkey (Ctrl+Space)            • voice/ (STT/TTS/Wake)
                                                    • mcp/
```

---

## 🛠️ Prerequisites

1. **Go**: Version `1.20` or higher.
2. **Node.js**: Version `18` or higher (for building the React frontend).
3. **Wails v2 CLI**: Installed via `go install github.com/wailsapp/wails/v2/cmd/wails@latest`.
4. **Python**: Version `3.10` or higher with packages: `sounddevice`, `numpy`, `openwakeword`, `onnxruntime`.

---

## 🚀 Installation & Setup

### 1. Clone & Setup

```bash
git clone https://github.com/your-username/nexa.git
cd nexa
```

### 2. Configure Environment (`.env`)

```env
# Supported Providers: groq, openai, openrouter, deepseek, gemini, ollama
NEXA_LLM_PROVIDER=groq

# Groq API (Free: https://console.groq.com/keys)
GROQ_API_KEY=gsk_your_groq_key_here
NEXA_GROQ_MODEL=llama-3.1-8b-instant

# Voice Output
NEXA_ENABLE_TTS=true
```

### 3. Build Executable

```powershell
# Build React frontend
cd gui/frontend
npm install
npx tsc; npx vite build
cd ../..

# Build NEXA desktop binary
go build -tags desktop,production -o nexa.exe ./cmd/nexa/
```

---

## 🎮 Usage Guide

### 🖥️ 1. Native Desktop GUI App

Launch NEXA's native desktop application:

```powershell
.\nexa.exe gui
```

- **Voice Command**: Click the **`🎤 START VOICE`** button in the Voice Reactor panel to turn on hands-free wake word detection ("Hey Nexa").
- **Global Hotkey**: Press **`Ctrl+Space`** at any time to show or hide NEXA.
- **System Tray**: NEXA displays a tray icon in Windows notification area.
- **In-Process Performance**: Zero network latency for GUI-to-agent communication; uses direct Go function invocation.

---

### 🎤 2. Hands-Free Voice Activation (CLI Mode)

```powershell
.\nexa.exe listen
```

- Say **"Hey Nexa"** or **"Nexa"** to activate.
- Speak your command (*"Open Spotify"*, *"What is the weather in Tokyo?"*).

---

### 💬 3. Terminal Chat & Ask Modes

```powershell
# Multi-turn CLI Chat
.\nexa.exe chat -t

# Single Question
.\nexa.exe ask "Summarize project status" -t
```

---

## 🔧 Registered Tools & Capabilities

| Tool | Actions | Description |
|------|---------|-------------|
| **`memory`** | `store`, `get`, `list`, `delete` | Long-term persistent memory store (`nexa_memory.json`). Remembers user preferences, project tech stacks, and context across sessions. Injects facts into system prompt automatically. |
| **`run_command`** | `execute` | Safely execute PowerShell and CMD commands. Features a 3-tier Command Risk Analyzer (`ALLOW` for safe read-only commands, `CONFIRM` for modifications, `DENY` for system wipes). |
| **`desktop_apps`** | `launch`, `close`, `list`, `focus` | Launch desktop apps, UWP store apps, File Explorer drives (`C:\`, `D:\`, `E:\`), close processes, list running windows. |
| **`filesystem`** | `list_dir`, `read_file`, `write_file`, `append_file`, `copy`, `move`, `delete`, `search`, `find_files`, `get_info` | Complete file manager operations with automatic drive letter normalization and safe permission handling. |
| **`web`** | `search`, `fetch`, `weather` | Live web search via DuckDuckGo POST scraping, URL content extraction, and real-time weather reports via `wttr.in`. |
| **`vision`** | `capture` | Captures a real-time screenshot of the primary Windows screen as a Base64 PNG payload for screen-aware AI responses. |
| **`mcp`** | `call` | JSON-RPC 2.0 stdio transport bridge to external MCP tool servers via the Model Context Protocol. |

---

## 📁 Project Structure

```
.
├── cmd/
│   └── nexa/             # Main CLI application entry point (cobra commands: chat, ask, listen, gui, voice, version)
├── config/               # Environment & system prompt configurations
├── core/                 # ReAct agent execution loop & hybrid tool parsers
├── gui/                  # Native Desktop Application (Wails v2 + React 18)
│   ├── app.go            # Wails App struct (Chat, StartVoiceEngine, GetMemories, GetStatus bindings)
│   ├── wails.go          # Native Wails window launcher & asset embed
│   ├── tray_windows.go   # Win32 System Tray icon (Shell_NotifyIconW) & Ctrl+Space global hotkey (RegisterHotKey)
│   └── frontend/         # React + TypeScript + Vite frontend dashboard
│       ├── src/
│       │   ├── components/  # ChatStream, VoiceReactor, MemoryPanel, StatusBar, ToolsList
│       │   ├── App.tsx      # 3-column dark glassmorphism layout
│       │   └── App.css      # Outfit + JetBrains Mono theme
├── llm/                  # Multi-LLM provider clients (Groq, OpenAI, OpenRouter, DeepSeek, Gemini, Ollama)
├── mcp/                  # Model Context Protocol JSON-RPC 2.0 stdio client
├── memory/               # Persistent JSON memory store engine (`nexa_memory.json`)
├── tools/                # Extensible tool registry (apps, filesystem, memory, terminal, web, vision, mcp)
├── voice/                # Audio subsystem (openWakeWord, Whisper STT, Piper TTS)
├── .env                  # Local configuration settings
├── wails.json            # Wails v2 project configuration
├── go.mod                # Go module definition
└── README.md             # Project documentation
```

---

## 📄 License

This project is licensed under the **MIT License**.
