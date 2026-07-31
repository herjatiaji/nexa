# 🤖 NEXA — Personal AI Desktop Companion & Speech Intelligence Platform

[![Release](https://img.shields.io/badge/Version-v2.1.0-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react)](https://reactjs.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io/)
[![SQLite](https://img.shields.io/badge/Database-SQLite-003B57?style=flat-square&logo=sqlite)](https://sqlite.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v2.1.0** is an advanced **Living AI Desktop Companion & Speech Intelligence Platform** built with Go, Wails v2, React, SQLite, and Python. It features the **NEXA Speech Intelligence Layer**, an **Emotion Engine & Mascot State Machine**, **Animated Mascot Facial Expressions**, **Floating Notification Speech Bubbles**, **Idle Pet Behaviors**, **Solarpunk Frutiger Aero UI**, **Plugin SDK Platform**, **OS-Style Permission Manager** (`🛡️ PERMISSIONS`), **Typed Parameter Schema Validation**, **SQLite Database Storage** (`nexa_data.db`), **Agent Tracing System** (`⚡ TRACES`), **Task Planning Engine**, **Semantic Vector Memory**, hands-free Voice Activation (`🎤 START VOICE`), System Tray integration, `Ctrl+Space` global hotkey, Whisper STT, Piper TTS, screen vision, and MCP support.

---

## ✨ Key Features in v2.1.0

### 🎙️ NEXA Speech Intelligence Layer (`voice/speech/`)
Replaces flat TTS with a multi-stage expressive speech synthesis architecture:
- 🧩 **Speech Planner (`speech/speech.go`)**: Segments responses into natural speech chunks (15–40 words). Uses single-chunk synthesis for short prompts and multi-chunk concatenation with 30ms crossfades for long explanations.
- 🎭 **Rule-Based Emotion Analyzer (`speech/emotion.go`)**: Analyzes text lexical patterns, punctuation density (`!`, `?`, `...`), and emojis to detect 7 distinct emotional tones: `neutral`, `happy`, `excited`, `thoughtful`, `urgent`, `sad`, and `confident` (0 API cost / 0 latency).
- 🎛️ **Prosody Controller (`speech/prosody.go`)**: Dynamically modulates Piper TTS synthesis parameters (`SpeedScale`, `PitchShift`, `SentenceSilence`, `Emphasis`) based on the detected emotion tag.
- 🇬🇧 **Personality Engine (`speech/personality.go`)**: Defines the NEXA persona (British professional, warm, articulate). Features `EmotionDamping` (0.55) to attenuate raw prosody changes so emotions sound natural, composed, and human-like rather than exaggerated.
- 🔊 **WAV Audio Post-Processor (`speech/postprocess.go`)**: Pure Go audio pipeline that peak-normalizes output to -3dBFS, trims leading/trailing silence, and crossfades multi-chunk audio files without external ffmpeg dependencies.

### 🤖 NEXA AI Mascot & Desktop Companion (`core/emotion.go`)
NEXA lives directly on your Windows desktop with dynamic facial expressions and real-time state machine indicators:
- `IDLE`: Relaxed aqua aura `( ◉ ◉ )`
- `LISTENING`: Glowing emerald green `( ⚡ ⚡ )`
- `THINKING`: Purple orbit rotation `( ◌ ◌ )`
- `EXECUTING`: Amber gold spark `( ⚙️ ⚙️ )`
- `SPEAKING`: Bright cyan audio wave `( 🔊 🔊 )`
- `HAPPY`: Cool lime sunglasses `( 😎 😎 )`
- `CONFUSED`: Coral red drop `( 😕 😕 )`
- `YAWN`: Soft aquamarine zzz after 3 minutes idle `( 💤 💤 )`

### 💬 Floating Speech Bubbles & Visual Emotion Badges
- **Speech Bubble Toasts (`FloatingCompanion.tsx`)**: Interactive glass toasts floating next to the mascot sphere when NEXA responds or executes tools (*"Opened Spotify! 😎"*, *"Listening for 'Hey Nexa'..."*).
- **Speech Emotion Indicator (`VoiceReactor.tsx`)**: Real-time color-coded badge showing current speech emotion and confidence level (e.g. `😊 Happy 92%`).

### 🌿 Solarpunk Frutiger Aero Aesthetic
- Skeuomorphic glass panels (`backdrop-filter: blur(30px)`), aquamarine/emerald ocean aurora backdrop, specular pill glass bubbles, and a liquid water droplet orb.

### ⚡ React Portal Modal Overlays
- Top-level modal overlays using `ReactDOM.createPortal` for **`⚡ TRACES`** and **`🛡️ PERMISSIONS`**, preventing layout clipping and CSS stacking context issues across window layers.

### 🔌 Plugin SDK & Core Infrastructure
- **Plugin SDK (`plugins/plugin.go`)**: Pre-bundled with `nexa-plugin-docker`.
- **Security Manager (`security/permissions.go`)**: Granular OS capability rules (`ALLOW`, `CONFIRM`, `DENY`).
- **SQLite DB (`memory/db.go`)**: CGO-free database with automatic legacy JSON migration.
- **Short-Term Context (`core/context.go`)**: Entity tracking for intuitive follow-up commands (*"Open it"*, *"Close the project"*).

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
     • FloatingCompanion (Solarpunk Mascot + 💬)      • core/agent.go (ReAct Loop)
     • VoiceReactor (orb + speech emotion badge)     • core/emotion.go (Emotion Engine)
     • MemoryPanel (semantic memory management)     • voice/speech/ (Speech Intelligence)
     • TraceInspector (⚡ TRACES React Portal)        • core/trace.go (Agent Tracer)
     • PermissionsModal (🛡️ PERMISSIONS Modal)       • core/planner.go (Task Planner)
     • StatusBar (provider & model telemetry)       • core/context.go (Short-Term State)
     • System Tray & Hotkey (Ctrl+Space)            • security/permissions.go (OS Rules)
                                                    • tools/schema.go (Parameter Validation)
                                                    • plugins/ (Plugin SDK Platform)
                                                    • memory/db.go (SQLite Database)
                                                    • llm/ (6 LLM Providers)
                                                    • voice/ (STT/TTS/Wake/Speech)
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

### 2. Run NEXA AI Desktop Companion

```powershell
.\nexa.exe gui
```

---

## 📄 License

This project is licensed under the **MIT License**.
