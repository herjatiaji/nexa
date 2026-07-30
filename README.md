# 🤖 NEXA — Personal AI Desktop & Voice Assistant

[![Release](https://img.shields.io/badge/Version-v1.1.0-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/Python-3.10%2B-3776AB?style=flat-square&logo=python)](https://python.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v1.1.0** is an ultra-fast, hands-free personal AI desktop assistant built in Go & Python. It features multi-LLM provider support (Groq, OpenAI, DeepSeek, OpenRouter, Gemini, Ollama), a live futuristic desktop GUI dashboard, real-time neural wake word detection, OpenAI Whisper speech-to-text, Rhasspy Piper neural voice synthesis, live web searching, persistent long-term memory, a 3-tier command risk analyzer, screen vision, and MCP protocol support — all in a seamless conversational agent.

---

## ✨ Key Features

- 🖥️ **Live Desktop GUI Dashboard**  
  A futuristic, dark-glassmorphism HUD window (Edge App Window or default browser) with real-time conversation stream, voice reactor orb, persistent memory inspector, and active tool indicators. Powered by an embedded Go HTTP server on `http://127.0.0.1:18420`.

- 🏷️ **Version 1.1.0**  
  Full semantic versioning, CLI flags for runtime provider/model switching (`-p` / `--provider`, `-m` / `--model`), and version diagnostics (`nexa version`).

- 🤖 **Multi-LLM Provider Engine**  
  Effortlessly switch between 6 major LLM providers:
  - ⚡ **Groq** (`llama-3.1-8b-instant`, `llama-3.3-70b-versatile`, `gemma2-9b-it`)
  - 🧠 **OpenAI** (`gpt-4o`, `gpt-4o-mini`, `o3-mini`)
  - 🌐 **OpenRouter** (`meta-llama/llama-3.3-70b-instruct`, `anthropic/claude-3.5-sonnet`, `deepseek/deepseek-r1`)
  - 🔬 **DeepSeek** (`deepseek-chat`, `deepseek-reasoner`)
  - ♊ **Google Gemini** (`gemini-2.0-flash`, `gemini-1.5-pro`)
  - 🦙 **Ollama (Local)** (`llama3.1`, `qwen2.5`, `mistral`)

- 🎙️ **Hands-Free Neural Wake Word ("Hey Nexa")**  
  Powered by `openWakeWord` ONNX neural stream and dynamic VAD audio recording on Windows.

- 🗣️ **Neural Voice Synthesis (Piper TTS)**  
  High-fidelity offline British female voice output using `en_GB-alba-medium.onnx` via Rhasspy Piper.

- 💬 **Continuous Multi-Turn Follow-Up Mode**  
  After responding, NEXA automatically keeps listening for follow-up commands so you can have fluid, multi-step conversations without repeating the wake word.

- 🧠 **Persistent Long-Term Memory Store (`nexa_memory.json`)**  
  Remembers user preferences, project tech stacks (e.g., *"Main project: Kazeer, Backend: Go"*), and context across restarts using the `memory` tool (`store`, `get`, `list`, `delete`). Automatically injects remembered facts into the system prompt and live GUI memory inspector!

- 🛡️ **3-Tier Command Risk Analyzer**  
  Categorizes shell commands into:
  - **`ALLOW`** (Safe read-only: `git status`, `dir`, `ls`, `ipconfig`) → Instant execution.
  - **`CONFIRM`** (Medium/High risk: `rm`, `del`, `git push --force`) → Asks user confirmation.
  - **`DENY`** (System wipe: `rm -rf /`, `format c:`, `diskpart`) → Blocked automatically.

- 👁️ **Screen Vision Tool**  
  Captures real-time Windows primary screen snapshots (Base64 PNG) and feeds them to the LLM for screen-aware assistance.

- 🔌 **MCP Protocol Client**  
  JSON-RPC 2.0 stdio transport client supporting Model Context Protocol for connecting to external MCP tool servers.

- 🌐 **Live Web Search & Weather Engine**  
  Real-time DuckDuckGo live web search and instant weather reporting via `wttr.in` for any location.

- 📁 **Multi-Partition File System Access**  
  Full access across `C:\`, `D:\`, `E:\`, and Windows user directories (`Documents`, `Downloads`, `Desktop`, `Pictures`, `Videos`) with safe permission handling.

- ⚙️ **Multi-Format ReAct Tool Parser**  
  Rock-solid tool parsing supporting JSON objects (`{"tool": ...}`), Markdown code blocks (` ```json ... ``` `), XML tags (`<desktop_apps>...</desktop_apps>`), and function signatures (`desktop_apps(...)`).

---

## 🏗️ System Architecture

```
                       ┌────────────────────────┐
                       │   Microphone Input     │
                       └───────────┬────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │  Python openWakeWord Engine │
                    │     Trigger: "Hey Nexa"     │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │  OpenAI Whisper STT (Groq)  │
                    │   Converts speech to text   │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │   ReAct AI Agent Loop (Go)  │
                    │ Multi-LLM Provider Engine   │
                    └──────┬───────────────┬──────┘
                           │               │
            ┌──────────────▼──────┐ ┌──────▼──────────────┐
            │   Tool Executions   │ │  Piper Neural TTS   │
            │ • desktop_apps      │ │  Voice Synthesis    │
            │ • filesystem        │ └─────────────────────┘
            │ • memory            │
            │ • web               │       ┌──────────────────────┐
            │ • run_command       │◄──────│  GUI Dashboard HUD   │
            │ • vision            │       │  http://127.0.0.1:   │
            │ • mcp               │       │       18420          │
            └─────────────────────┘       └──────────────────────┘
```

---

## 🛠️ Prerequisites

1. **Go**: Version `1.20` or higher.
2. **Python**: Version `3.10` or higher with required packages:
   ```bash
   pip install sounddevice numpy openwakeword onnxruntime
   ```
3. **API Key**: Groq (free), OpenAI, OpenRouter, DeepSeek, or Gemini API key.
4. **Microsoft Edge** (recommended) or any default browser for the GUI Dashboard.

---

## 🚀 Installation & Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-username/nexa.git
cd nexa
```

### 2. Configure Environment (`.env`)

Choose your preferred LLM provider in `.env`:

```env
# Supported Providers: groq, openai, openrouter, deepseek, gemini, ollama
NEXA_LLM_PROVIDER=groq

# Groq API (Free: https://console.groq.com/keys)
GROQ_API_KEY=gsk_your_groq_key_here
NEXA_GROQ_MODEL=llama-3.1-8b-instant

# OpenAI API (https://platform.openai.com/api-keys)
OPENAI_API_KEY=sk-your_openai_key_here
NEXA_OPENAI_MODEL=gpt-4o-mini

# OpenRouter API (https://openrouter.ai/keys)
OPENROUTER_API_KEY=sk-or-v1-your_openrouter_key_here
NEXA_OPENROUTER_MODEL=meta-llama/llama-3.3-70b-instruct

# DeepSeek API (https://platform.deepseek.com/api_keys)
DEEPSEEK_API_KEY=sk-your_deepseek_key_here
NEXA_DEEPSEEK_MODEL=deepseek-chat

# Google Gemini (Free: https://aistudio.google.com/apikey)
GEMINI_API_KEY=your_gemini_key_here
NEXA_GEMINI_MODEL=gemini-2.0-flash

# Ollama Local (http://localhost:11434)
NEXA_OLLAMA_URL=http://localhost:11434
NEXA_OLLAMA_MODEL=llama3.1

# Voice Output
NEXA_ENABLE_TTS=true
```

### 3. Build Executable

```bash
go build -o nexa.exe ./cmd/nexa/
```

---

## 🎮 Usage Guide

### 🖥️ 1. Desktop GUI Dashboard (New!)

Launch NEXA's futuristic live HUD dashboard window:

```powershell
.\nexa.exe gui
```

**What you get:**
- 💬 **Live Conversation Stream** — Type commands and see NEXA's full responses with real-time tool execution badges (e.g., `🔧 memory · store`).
- 🎯 **Voice Reactor Orb** — Animates to a *purple glow* while NEXA is thinking and executing tools.
- 🧠 **Persistent Memory Inspector** — Right panel shows all stored long-term memories, auto-refreshing after every interaction.
- 📡 **Provider Status Bar** — Top bar shows the active LLM provider and model name live (e.g., `GROQ (llama-3.1-8b-instant)`).
- 🛑 **Anti-double-submit protection** — Input is disabled while NEXA is processing.

The GUI runs a local REST API server at `http://127.0.0.1:18420`. Close the browser window to stop.

---

### 🎤 2. Hands-Free Voice Activation Mode

Start NEXA's persistent voice listening engine:

```powershell
.\nexa.exe listen
```

- Say **"Hey Nexa"** or **"Nexa"** to activate.
- Ask your command (e.g., *"Open Spotify"*, *"Check weather in London"*, *"Remember my project is Kazeer"*).
- NEXA responds and opens a continuous follow-up window for fluid multi-step conversations.

---

### 💬 3. Interactive Terminal Chat Mode

```powershell
.\nexa.exe chat -t
```

Runs a colored REPL interface in PowerShell with optional TTS output (`-t`).

---

### ❓ 4. Single Question (CLI Mode)

```powershell
.\nexa.exe ask "Open drive E in File Manager" -t
```

---

### ⚡ 5. Switch Providers & Models on the Fly

Use `-p` (`--provider`) and `-m` (`--model`) flags to instantly switch models without editing `.env`:

```powershell
# Run with Groq 70B
.\nexa.exe ask "Who won the F1 championship?" -p groq -m llama-3.3-70b-versatile

# Run GUI Dashboard with GPT-4o
.\nexa.exe gui -p openai -m gpt-4o

# Run with DeepSeek Chat
.\nexa.exe ask "Explain quantum computing" -p deepseek -m deepseek-chat

# Start Hands-Free Mode with OpenRouter Claude
.\nexa.exe listen -p openrouter -m anthropic/claude-3.5-sonnet
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

## 🌐 GUI Dashboard REST API

The `nexa gui` command starts an embedded HTTP server at `http://127.0.0.1:18420` exposing:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `GET /` | GET | Serves the NEXA Dashboard HTML/CSS/JS frontend. |
| `POST /api/chat` | POST | Accepts `{"message": "..."}`, runs agent, returns `{"response": "...", "tool_calls": [...], "memories": {...}}`. |
| `GET /api/memories` | GET | Returns all long-term memories as `{"key": "value", ...}`. |
| `GET /api/status` | GET | Returns `{"provider": "groq", "model": "llama-3.1-8b-instant", "tts": true, "version": "1.1.0"}`. |

---

## 📁 Project Structure

```
.
├── cmd/
│   └── nexa/             # Main CLI application entry point (cobra commands: chat, ask, listen, gui, voice, version)
├── config/               # Environment & system prompt configurations
├── core/                 # ReAct agent execution loop & hybrid tool parsers
├── gui/                  # Desktop GUI Dashboard
│   ├── server.go         # Embedded Go HTTP server + REST API (chat, memories, status)
│   ├── gui.go            # GUI launcher (Edge app window / browser fallback)
│   └── index.html        # Futuristic dark-glassmorphism HTML/CSS/JS dashboard
├── llm/                  # Multi-LLM provider clients (Groq, OpenAI, OpenRouter, DeepSeek, Gemini, Ollama)
├── mcp/                  # Model Context Protocol JSON-RPC 2.0 stdio client
├── memory/               # Persistent JSON memory store engine (`nexa_memory.json`)
├── tools/                # Extensible tool registry
│   ├── apps/             # 5-Layer Windows application launcher & Explorer control
│   ├── filesystem/       # Multi-drive file system manager
│   ├── mcp/              # MCP bridge tool wrapper
│   ├── memory/           # Memory management tool (store, get, list, delete)
│   ├── terminal/         # 3-Tier Command Risk Analyzer & PowerShell executor
│   ├── vision/           # Windows primary screen capture tool
│   └── web/              # Live web search & weather tool
├── voice/                # Audio subsystem
│   ├── openwakeword_listener.py  # Real-time ONNX wake word & VAD recorder
│   ├── wake.go           # Go-Python IPC process bridge
│   ├── whisper.go        # OpenAI Whisper Large v3 STT client
│   └── tts.go            # Rhasspy Piper neural voice synthesis wrapper
├── .env                  # Local configuration settings
├── go.mod                # Go module definition
└── README.md             # Project documentation
```

---

## 📄 License

This project is licensed under the **MIT License**.
