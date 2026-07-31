# 🤖 NEXA — Personal AI Desktop & Voice Assistant

[![Release](https://img.shields.io/badge/Version-v1.3.0-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react)](https://reactjs.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io/)
[![SQLite](https://img.shields.io/badge/Database-SQLite-003B57?style=flat-square&logo=sqlite)](https://sqlite.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v1.3.0** is an ultra-fast, hands-free personal AI desktop assistant built in Go, Wails v2, React, SQLite, and Python. It features a **Native Windows Desktop Application**, **SQLite Database Storage** (`nexa_data.db`), **Agent Tracing System**, **Task Planning Engine**, **Semantic Vector Memory**, interactive Voice Activation (`🎤 START VOICE`), multi-LLM provider support (Groq, OpenAI, DeepSeek, OpenRouter, Gemini, Ollama), System Tray integration, `Ctrl+Space` global hotkey, real-time neural wake word detection, OpenAI Whisper speech-to-text, Rhasspy Piper neural voice synthesis, live web searching, a 3-tier command risk analyzer, screen vision, and Model Context Protocol (MCP) support.

---

## ✨ Key Features in v1.3.0

- 🗄️ **SQLite Database Storage (`nexa_data.db`)**  
  Upgraded memory storage engine from single-file JSON to a high-performance CGO-free SQLite database (`modernc.org/sqlite`). Automatically migrates legacy `nexa_memory.json` data on startup with zero data loss! Includes tables for `memories` and `conversations`.

- ⚡ **Agent Tracing System (`⚡ TRACES`)**  
  Full developer mode trace logging. Inspect real-time reasoning steps, thoughts, tool invocations, arguments, and raw tool output via the interactive **`⚡ TRACES`** drawer in the React GUI Desktop App.

- 🎯 **Task Planning Engine**  
  Decomposes complex multi-step user prompts into structured sub-tasks (`Plan: [Task 1, Task 2, Task 3]`) before execution.

- 🧠 **Semantic Vector Memory**  
  Conceptual similarity search over stored memories using cosine similarity over token frequency vectors. Querying *"remind me about my food idea"* automatically matches related concepts like *"family recipe pasta"*.

- 🖥️ **Native Windows Desktop App (Wails v2 + React 18 + TS)**  
  Standalone native desktop application using Wails v2 with direct Go↔JS bindings (no HTTP API, no localhost browser required).

- 🎙️ **Desktop GUI Voice Activation Mode (`🎤 START VOICE`)**  
  Toggle hands-free voice command listening inside the GUI desktop app. Say **"Hey Nexa"** or **"Nexa"** to activate and speak commands naturally.

- 🔔 **Windows System Tray & `Ctrl+Space` Global Hotkey**  
  Runs in the notification tray. Press **`Ctrl+Space`** at any time to toggle NEXA to the foreground.

- 🤖 **Multi-LLM Provider Engine**  
  Supports Groq, OpenAI, OpenRouter, DeepSeek, Google Gemini, and Ollama Local.

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
     • TraceInspector (⚡ TRACES Drawer)             • memory/db.go (SQLite Database)
     • StatusBar (provider & model info)            • memory/semantic.go (Vector Match)
     • System Tray & Hotkey (Ctrl+Space)            • llm/ (6 Providers)
                                                    • tools/ (7 Tools)
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

## 🔧 Registered Tools & Capabilities

| Tool | Actions | Description |
|------|---------|-------------|
| **`memory`** | `store`, `get`, `list`, `delete` | Long-term persistent SQLite memory store (`nexa_data.db`). Supports semantic vector search and dynamic system prompt injection. |
| **`run_command`** | `execute` | Safely execute PowerShell and CMD commands with 3-tier Risk Analyzer (`ALLOW`, `CONFIRM`, `DENY`). |
| **`desktop_apps`** | `launch`, `close`, `list`, `focus` | Launch desktop apps, UWP store apps, File Explorer drives, close processes. |
| **`filesystem`** | `list_dir`, `read_file`, `write_file`, `append_file`, `copy`, `move`, `delete`, `search`, `find_files`, `get_info` | Complete file manager operations with drive letter normalization. |
| **`web`** | `search`, `fetch`, `weather` | DuckDuckGo live web search, URL content extraction, and weather reporting via `wttr.in`. |
| **`vision`** | `capture` | Captures primary Windows screen snapshot as Base64 PNG payload. |
| **`mcp`** | `call` | JSON-RPC 2.0 stdio transport bridge to external MCP tool servers. |

---

## 📄 License

This project is licensed under the **MIT License**.
