# JARVIS — Personal AI Desktop Assistant

A personal AI assistant built in Go that can understand commands, plan actions, execute tools, speak with authentic voice, and automate workflows.

## Features

- 🤖 **AI-Powered** — Uses LLM (Gemini/Groq/Ollama) for intelligent responses
- 🔊 **Voice Activation & TTS** — Hands-free wake word activation ("Hey Jarvis") with authentic British voice synthesis
- 🖥️ **Desktop App Control** — Launch, close, list, and focus Windows desktop applications
- 📁 **Full File Manager** — Read, write, append, copy, move, delete, create dirs, and find files
- 🔧 **Terminal Tools** — Execute shell commands safely with confirmation prompts
- 💬 **Interactive Chat** — Colored terminal UI with multi-line input support

## Quick Start

### 1. Configure `.env`

Copy or edit `.env`:
```env
JARVIS_LLM_PROVIDER=groq
GROQ_API_KEY=gsk_your_groq_api_key_here
JARVIS_ENABLE_TTS=true
```

### 2. Run Commands

```bash
# Hands-Free Voice Activation Mode (Say "Hey Jarvis" to trigger)
./jarvis listen

# Interactive chat mode (with voice)
./jarvis chat -t

# Single question
./jarvis ask "Buka kalkulator dan list semua file di folder ini" -t

# Voice Sample & Voice List
./jarvis voice sample
./jarvis voice list
```

## Usage Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `jarvis listen` | Hands-Free | Continuous background wake word activation mode ("Hey Jarvis") |
| `jarvis chat [-t]` | Interactive | Interactive chat REPL mode (optional `-t` for TTS voice output) |
| `jarvis ask "..." [-t]` | One-off | Ask a single question (optional `-t` for TTS voice output) |
| `jarvis voice sample` | Voice Test | Play iconic JARVIS voice greeting sample |
| `jarvis voice list` | System Voices | List all installed speech synthesis voices |

## Capabilities & Tools

1. **`desktop_apps`**: Launch apps (`code`, `chrome`, `calc`, `notepad`, `spotify`, `explorer`), close processes, list running apps, focus windows.
2. **`filesystem`**: Read/write/append files, create dirs, delete files/folders, copy/move, view metadata (`get_info`), search text, find files by pattern (`find_files`).
3. **`run_command`**: Execute any PowerShell / shell command safely.

## License

MIT
