# 🤖 NEXA — Personal AI Desktop & Voice Assistant

[![Release](https://img.shields.io/badge/Version-v1.4.0-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react)](https://reactjs.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io/)
[![SQLite](https://img.shields.io/badge/Database-SQLite-003B57?style=flat-square&logo=sqlite)](https://sqlite.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v1.4.0** is an ultra-fast, hands-free personal AI desktop assistant built in Go, Wails v2, React, SQLite, and Python. It features a **Plugin SDK Platform**, **OS-Style Permission Manager** (`🛡️ PERMISSIONS`), **Typed Parameter Schema Validation**, **Short-Term Conversation Context**, **SQLite Database Storage** (`nexa_data.db`), **Agent Tracing System** (`⚡ TRACES`), **Task Planning Engine**, **Semantic Vector Memory**, interactive Voice Activation (`🎤 START VOICE`), multi-LLM provider support (Groq, OpenAI, DeepSeek, OpenRouter, Gemini, Ollama), System Tray integration, `Ctrl+Space` global hotkey, real-time neural wake word detection, OpenAI Whisper speech-to-text, Rhasspy Piper neural voice synthesis, live web searching, screen vision, and Model Context Protocol (MCP) support.

---

## ✨ Key Features in v1.4.0

- 🔌 **Plugin SDK Platform (`plugins/plugin.go`)**  
  Modular Go plugin interface (`Plugin`) allowing developers to extend NEXA with custom plugins (e.g. `nexa-plugin-docker`, `nexa-plugin-github`). Comes pre-bundled with an example Docker plugin for container management (`docker ps`, `logs`, `start`, `stop`).

- 🛡️ **OS-Style Agent Permission Manager (`security/permissions.go`)**  
  Granular security access control model (`nexa_permissions.json`). Set capability rules (`ALLOW`, `CONFIRM`, `DENY`) for applications, file deletion, and shell execution via the interactive **`🛡️ PERMISSIONS`** modal in the React GUI Desktop App.

- 📐 **Typed Tool Schema & Parameter Validation (`tools/schema.go`)**  
  Strict parameter JSON Schema validation with `enum`, `required`, and `type` constraints. Catch invalid LLM tool parameters before execution.

- 🧠 **Short-Term Conversation Context Manager (`core/context.go`)**  
  Tracks session context entities (`last_app`, `last_path`) across turns. Resolves ambiguous follow-up references like *"Open the project"* or *"Close it"*.

- 🗄️ **SQLite Database Storage (`nexa_data.db`)**  
  High-performance SQLite database storage (`modernc.org/sqlite`). Auto-migrates legacy `nexa_memory.json` data on startup.

- ⚡ **Agent Tracing System (`⚡ TRACES`)**  
  Inspect real-time reasoning thoughts, tool invocations, arguments, and raw output via the **`⚡ TRACES`** drawer in the React GUI Desktop App.

- 🎯 **Task Planning Engine**  
  Decomposes complex multi-step user prompts into structured sub-tasks (`Plan: [Task 1, Task 2, Task 3]`).

- 🖥️ **Native Windows Desktop App (Wails v2 + React 18 + TS)**  
  Standalone native desktop application using Wails v2 with direct Go↔JS bindings.

- 🎙️ **Desktop GUI Voice Activation Mode (`🎤 START VOICE`)**  
  Hands-free wake word detection ("Hey Nexa") directly inside the GUI desktop app.

- 🔔 **Windows System Tray & `Ctrl+Space` Global Hotkey**  
  Runs in the notification tray. Press **`Ctrl+Space`** at any time to toggle NEXA to the foreground.

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
     • ChatStream (with tool badges)                • core/agent.go (ReAct Loop)
     • VoiceReactor (animated orb + 🎤 Voice ON)     • core/trace.go (Agent Tracer)
     • MemoryPanel (with live delete)               • core/planner.go (Task Planner)
     • TraceInspector (⚡ TRACES Drawer)             • core/context.go (Short-Term State)
     • PermissionsModal (🛡️ PERMISSIONS)            • security/permissions.go (OS Rules)
     • StatusBar (provider & model info)            • tools/schema.go (Parameter Validation)
     • System Tray & Hotkey (Ctrl+Space)            • plugins/ (Plugin SDK Platform)
                                                    • memory/db.go (SQLite Database)
                                                    • llm/ (6 Providers)
                                                    • voice/ (STT/TTS/Wake)
```

---

## 🛠️ Prerequisites

1. **Go**: Version `1.20` or higher.
2. **Node.js**: Version `18` or higher.
3. **Wails v2 CLI**: Installed via `go install github.com/wailsapp/wails/v2/cmd/wails@latest`.
4. **Python**: Version `3.10` or higher with packages: `sounddevice`, `numpy`, `openwakeword`, `onnxruntime`.

---

## 🚀 Installation & Setup

### 1. Build Executable

```powershell
# Build React frontend
cd gui/frontend
npm install
npx tsc; npx vite build
cd ../..

# Build NEXA desktop binary
go build -tags desktop,production -o nexa.exe ./cmd/nexa/
```

### 2. Run NEXA Desktop App

```powershell
.\nexa.exe gui
```

---

## 🔧 Registered Tools & Plugins

| Tool / Plugin | Type | Description |
|---------------|------|-------------|
| **`docker`** | Plugin (`nexa-plugin-docker`) | Docker container management plugin (`docker ps`, `logs`, `start`, `stop`, `inspect`). |
| **`memory`** | Built-in Tool | Persistent SQLite memory store (`nexa_data.db`). Supports semantic vector search and dynamic system prompt injection. |
| **`run_command`** | Built-in Tool | Safely execute PowerShell and CMD commands with OS Permission rules. |
| **`desktop_apps`** | Built-in Tool | Launch desktop apps, UWP store apps, File Explorer drives, close processes. |
| **`filesystem`** | Built-in Tool | Complete file manager operations with drive letter normalization. |
| **`web`** | Built-in Tool | DuckDuckGo live web search, URL content extraction, and weather reporting via `wttr.in`. |
| **`vision`** | Built-in Tool | Captures primary Windows screen snapshot as Base64 PNG payload. |
| **`mcp`** | Built-in Tool | JSON-RPC 2.0 stdio transport bridge to external MCP tool servers. |

---

## 📁 Project Structure

```
.
├── cmd/
│   └── nexa/             # Main CLI application entry point
├── config/               # Environment & system prompt configurations
├── core/                 # ReAct agent loop, TraceRecorder, Planner, ContextManager
├── gui/                  # Native Desktop Application (Wails v2 + React 18)
│   ├── app.go            # Wails App struct (Chat, GetPermissions, SetPermission, GetTraces)
│   ├── wails.go          # Native Wails window launcher & asset embed
│   ├── tray_windows.go   # Win32 System Tray icon & Ctrl+Space global hotkey
│   └── frontend/         # React + TypeScript + Vite frontend dashboard
│       └── src/components/ # ChatStream, VoiceReactor, MemoryPanel, TraceInspector, PermissionsModal
├── llm/                  # Multi-LLM provider clients (Groq, OpenAI, OpenRouter, DeepSeek, Gemini, Ollama)
├── mcp/                  # Model Context Protocol JSON-RPC 2.0 stdio client
├── memory/               # SQLite database store (`nexa_data.db`) & semantic vector search
├── plugins/              # Plugin SDK Platform
│   ├── plugin.go         # Plugin interface & Registry
│   └── docker/           # Built-in Docker management plugin
├── security/             # OS-Style Permission Manager (`nexa_permissions.json`)
├── tools/                # Extensible tool registry & typed schema validator (`schema.go`)
├── voice/                # Audio subsystem (openWakeWord, Whisper STT, Piper TTS)
├── .env                  # Local configuration settings
├── wails.json            # Wails v2 project configuration
└── README.md             # Project documentation
```

---

## 📄 License

This project is licensed under the **MIT License**.
