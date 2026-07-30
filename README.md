# 🤖 NEXA — Personal AI Desktop & Voice Assistant

[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/Python-3.10%2B-3776AB?style=flat-square&logo=python)](https://python.org/)
[![Groq API](https://img.shields.io/badge/LLM-Groq%20Llama%203.1--8B-f05032?style=flat-square)](https://groq.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA** is an ultra-fast, hands-free personal AI desktop assistant built in Go & Python. It combines real-time neural wake word detection, OpenAI Whisper speech-to-text, Rhasspy Piper neural voice synthesis, live web searching, and deep Windows OS desktop automation into a seamless conversational agent.

---

## ✨ Key Features

- 🎙️ **Hands-Free Neural Wake Word ("Hey Nexa")**  
  Powered by `openWakeWord` ONNX neural stream and dynamic VAD audio recording on Windows.
- 🗣️ **Neural Voice Synthesis (Piper TTS)**  
  High-fidelity offline British female voice output using `en_GB-alba-medium.onnx` via Rhasspy Piper.
- ⚡ **Groq Ultra-Fast ReAct Agent Engine**  
  Uses `llama-3.1-8b-instant` (~100ms response time) with automatic 429 rate-limit fallback and text-embedded tool call parsers (`<desktop_apps>...</desktop_apps>`).
- 💬 **Continuous Multi-Turn Follow-Up Mode**  
  After responding, NEXA automatically keeps listening for follow-up commands so you can have fluid, multi-step conversations without repeating the wake word.
- 🌐 **Live Web Search & Weather Engine**  
  Real-time DuckDuckGo live web search and instant weather reporting via `wttr.in` for any location.
- 📁 **Multi-Partition File System Access**  
  Full access across `C:\`, `D:\`, `E:\`, and Windows user directories (`Documents`, `Downloads`, `Desktop`, `Pictures`, `Videos`) with safe permission handling.
- 🖥️ **5-Layer Windows Application Launcher**  
  Launches Windows apps via URI protocols (`spotify:`, `calculator:`), executables, Microsoft Store UWP apps (`shell:AppsFolder`), and AppData paths.
- ⚙️ **Safe Shell Execution Engine**  
  Execute PowerShell and CMD terminal commands with automatic interactive safety prompts for sensitive commands.

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
                    │    llama-3.1-8b-instant     │
                    └──────┬───────────────┬──────┘
                           │               │
            ┌──────────────▼──────┐ ┌──────▼──────────────┐
            │   Tool Executions   │ │  Piper Neural TTS   │
            │ • desktop_apps      │ │  Voice Synthesis    │
            │ • filesystem        │ └─────────────────────┘
            │ • web               │
            │ • run_command       │
            └─────────────────────┘
```

---

## 🛠️ Prerequisites

1. **Go**: Version `1.20` or higher.
2. **Python**: Version `3.10` or higher with required packages:
   ```bash
   pip install sounddevice numpy openwakeword onnxruntime
   ```
3. **Groq API Key**: Get a free API key at [console.groq.com](https://console.groq.com/keys).

---

## 🚀 Installation & Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-username/nexa.git
cd nexa
```

### 2. Configure Environment (`.env`)

Create or edit `.env` in the root directory:

```env
JARVIS_LLM_PROVIDER=groq
GROQ_API_KEY=gsk_your_groq_api_key_here
JARVIS_GROQ_MODEL=llama-3.1-8b-instant
JARVIS_ENABLE_TTS=true
```

### 3. Build Executable

```bash
go build -o jarvis.exe ./cmd/jarvis/
```

---

## 🎮 Usage Guide

### 🎤 1. Hands-Free Voice Activation Mode (Recommended)

Start NEXA's persistent voice listening engine:

```powershell
.\jarvis.exe listen
```

- Say **"Hey Nexa"** or **"Nexa"** to activate.
- Ask your command (e.g., *"Open Spotify"*, *"Check weather in London"*).
- NEXA will respond and immediately open a **continuous follow-up window** so you can speak your next command right away!

---

### 💬 2. Interactive Terminal Chat Mode

```powershell
.\jarvis.exe chat -t
```
Runs an interactive colored REPL interface in PowerShell with text-to-speech output enabled (`-t`).

---

### ❓ 3. Single Question (CLI Mode)

```powershell
.\jarvis.exe ask "Open drive E in File Manager" -t
```

---

## 🔧 Registered Tools & Capabilities

| Tool | Actions | Description |
|------|---------|-------------|
| **`desktop_apps`** | `launch`, `close`, `list`, `focus` | Launch desktop apps, UWP store apps, File Explorer drives (`C:\`, `D:\`, `E:\`), close processes, list running windows. |
| **`filesystem`** | `list_dir`, `read_file`, `write_file`, `append_file`, `copy`, `move`, `delete`, `search`, `find_files`, `get_info` | Complete file manager operations with automatic drive letter normalization (`D:` → `D:\`) and safe permission error handling. |
| **`web`** | `search`, `fetch`, `weather` | Live web search via DuckDuckGo POST scraping, URL content extraction, and real-time weather reports via `wttr.in`. |
| **`run_command`** | `execute` | Safely execute PowerShell and CMD commands with built-in interactive confirmation for system-altering commands. |

---

## 📁 Project Structure

```
.
├── cmd/
│   └── jarvis/           # Main CLI application entry point (cobra commands)
├── config/               # Environment & system prompt configurations
├── core/                 # ReAct agent execution loop & hybrid tool parsers
├── llm/                  # Groq API client with automatic rate-limit fallbacks
├── tools/                # Extensible tool registry
│   ├── apps/             # 5-Layer Windows application launcher & Explorer control
│   ├── filesystem/       # Multi-drive file system manager
│   ├── terminal/         # Safe PowerShell command execution tool
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
