# 🤖 NEXA — Personal AI Desktop & Voice Assistant

[![Release](https://img.shields.io/badge/Version-v1.0.0-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/Python-3.10%2B-3776AB?style=flat-square&logo=python)](https://python.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v1.0.0** is an ultra-fast, hands-free personal AI desktop assistant built in Go & Python. It features multi-LLM provider support (Groq, OpenAI, DeepSeek, OpenRouter, Gemini, Ollama), real-time neural wake word detection, OpenAI Whisper speech-to-text, Rhasspy Piper neural voice synthesis, live web searching, and deep Windows OS desktop automation into a seamless conversational agent.

---

## ✨ Key Features & Versioning

- 🏷️ **Version 1.0.0 Release**  
  Full semantic versioning, CLI flags for runtime provider/model switching (`-p` / `--provider`, `-m` / `--model`), and version diagnostics (`nexa version`).
- 🤖 **Multi-LLM Provider Engine**  
  Effortlessly switch between 6 major LLM providers:
  - ⚡ **Groq** (`llama-3.1-8b-instant`, `llama-3.3-70b-versatile`, `gemma2-9b-it`)
  - 🧠 **OpenAI** (`gpt-4o`, `gpt-4o-mini`, `o3-mini`)
  - 🌐 **OpenRouter** (`meta-llama/llama-3.3-70b-instruct`, `anthropic/claude-3.5-sonnet`)
  - 🔬 **DeepSeek** (`deepseek-chat`, `deepseek-reasoner`)
  - ♊ **Google Gemini** (`gemini-2.0-flash`, `gemini-1.5-pro`)
  - 🦙 **Ollama (Local)** (`llama3.1`, `qwen2.5`, `mistral`)
- 🎙️ **Hands-Free Neural Wake Word ("Hey Nexa")**  
  Powered by `openWakeWord` ONNX neural stream and dynamic VAD audio recording on Windows.
- 🗣️ **Neural Voice Synthesis (Piper TTS)**  
  High-fidelity offline British female voice output using `en_GB-alba-medium.onnx` via Rhasspy Piper.
- 💬 **Continuous Multi-Turn Follow-Up Mode**  
  After responding, NEXA automatically keeps listening for follow-up commands so you can have fluid, multi-step conversations without repeating the wake word.
- 🌐 **Live Web Search & Weather Engine**  
  Real-time DuckDuckGo live web search and instant weather reporting via `wttr.in` for any location.
- 📁 **Multi-Partition File System Access**  
  Full access across `C:\`, `D:\`, `E:\`, and Windows user directories (`Documents`, `Downloads`, `Desktop`, `Pictures`, `Videos`).

---

## 🚀 Environment Configuration (`.env`)

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

---

## 🎮 CLI Usage & Provider Override

### Switch Providers & Models on the Fly

Use `-p` (`--provider`) and `-m` (`--model`) flags to instantly switch models without editing `.env`:

```powershell
# Run with Groq 70B
.\nexa.exe ask "Who won the F1 championship?" -p groq -m llama-3.3-70b-versatile

# Run with OpenAI GPT-4o-mini
.\nexa.exe ask "Summarize this file" -p openai -m gpt-4o-mini

# Run with DeepSeek Chat
.\nexa.exe ask "Explain quantum computing" -p deepseek -m deepseek-chat

# Start Hands-Free Mode with OpenRouter
.\nexa.exe listen -p openrouter -m meta-llama/llama-3.3-70b-instruct
```

### 3. Build Executable

```bash
go build -o nexa.exe ./cmd/nexa/
```

---

## 🎮 Usage Guide

### 🎤 1. Hands-Free Voice Activation Mode (Recommended)

Start NEXA's persistent voice listening engine:

```powershell
.\nexa.exe listen
```

- Say **"Hey Nexa"** or **"Nexa"** to activate.
- Ask your command (e.g., *"Open Spotify"*, *"Check weather in London"*).
- NEXA will respond and immediately open a **continuous follow-up window** so you can speak your next command right away!

---

### 💬 2. Interactive Terminal Chat Mode

```powershell
.\nexa.exe chat -t
```
Runs an interactive colored REPL interface in PowerShell with text-to-speech output enabled (`-t`).

---

### ❓ 3. Single Question (CLI Mode)

```powershell
.\nexa.exe ask "Open drive E in File Manager" -t
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
│   └── nexa/             # Main CLI application entry point (cobra commands)
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
