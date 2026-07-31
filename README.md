# 🤖 NEXA — Personal AI Desktop Companion & Voice Assistant

[![Release](https://img.shields.io/badge/Version-v2.0.0-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react)](https://reactjs.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io/)
[![SQLite](https://img.shields.io/badge/Database-SQLite-003B57?style=flat-square&logo=sqlite)](https://sqlite.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v2.0.0** transforms NEXA from a simple assistant into a **Living AI Desktop Companion** in Go, Wails v2, React, SQLite, and Python. It features a Solarpunk Frutiger Aero UI theme, an **Emotion Engine & Mascot State Machine**, **Animated Facial Expressions**, **Floating Notification Speech Bubbles**, **Idle Pet Behaviors**, **Plugin SDK Platform**, **OS-Style Permission Manager** (`🛡️ PERMISSIONS`), **Typed Parameter Schema Validation**, **SQLite Database Storage** (`nexa_data.db`), **Agent Tracing System** (`⚡ TRACES`), **Task Planning Engine**, **Semantic Vector Memory**, hands-free Voice Activation (`🎤 START VOICE`), System Tray integration, `Ctrl+Space` global hotkey, Whisper STT, Piper TTS, screen vision, and MCP support.

---

## ✨ Key Features in v2.0.0

- 🤖 **NEXA AI Desktop Companion & Emotion Engine (`core/emotion.go`)**  
  NEXA now has a living personality on your desktop! Features real-time state machine expressions:
  - `IDLE`: Relaxed aqua aura `( ◉ ◉ )`
  - `LISTENING`: Glowing emerald green `( ⚡ ⚡ )`
  - `THINKING`: Purple orbit rotation `( ◌ ◌ )`
  - `EXECUTING`: Amber gold spark `( ⚙️ ⚙️ )`
  - `SPEAKING`: Bright cyan audio wave `( 🔊 🔊 )`
  - `HAPPY`: Cool lime sunglasses `( 😎 😎 )`
  - `CONFUSED`: Coral red drop `( 😕 😕 )`
  - `YAWN`: Soft aquamarine zzz after 3m idle `( 💤 💤 )`

- 💬 **Floating Notification Speech Bubbles (`FloatingCompanion.tsx`)**  
  Interactive speech bubble toasts floating right next to the mascot sphere when NEXA responds or executes tools (*"Opened Spotify! 😎"*, *"Listening for 'Hey Nexa'..."*).

- 🌿 **Solarpunk Frutiger Aero Theme**  
  Glossy skeuomorphic glass panels, water droplet orb reflection, aquamarine/emerald aurora backdrop, and specular pill glass bubbles.

- 🔌 **Plugin SDK Platform (`plugins/plugin.go`)**  
  Modular Go plugin interface (`Plugin`) allowing custom third-party plugins. Pre-bundled with `nexa-plugin-docker`.

- 🛡️ **OS-Style Agent Permission Manager (`security/permissions.go`)**  
  Granular security access control (`ALLOW`, `CONFIRM`, `DENY`) for applications, files, and commands via the **`🛡️ PERMISSIONS`** modal.

- 📐 **Typed Tool Schema & Parameter Validation (`tools/schema.go`)**  
  Strict JSON Schema validation with `enum`, `required`, and `type` constraints.

- 🧠 **Short-Term Conversation Context Manager (`core/context.go`)**  
  Tracks session context entities (`last_app`, `last_path`). Resolves ambiguous references like *"Open the project"* or *"Close it"*.

- 🗄️ **SQLite Database Storage (`nexa_data.db`)**  
  CGO-free SQLite database (`modernc.org/sqlite`). Auto-migrates legacy `nexa_memory.json` data on startup.

- ⚡ **Agent Tracing System (`⚡ TRACES`)**  
  Inspect real-time reasoning thoughts, tool invocations, arguments, and raw output via the **`⚡ TRACES`** drawer.

- 🔔 **Windows System Tray & `Ctrl+Space` Global Hotkey**  
  Press **`Ctrl+Space`** at any time to toggle NEXA to the foreground.

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
     • VoiceReactor (animated orb + 🎤 Voice ON)     • core/emotion.go (Emotion Engine)
     • MemoryPanel (with live delete)               • core/trace.go (Agent Tracer)
     • TraceInspector (⚡ TRACES Drawer)             • core/planner.go (Task Planner)
     • PermissionsModal (🛡️ PERMISSIONS)            • core/context.go (Short-Term State)
     • StatusBar (provider & model info)            • security/permissions.go (OS Rules)
     • System Tray & Hotkey (Ctrl+Space)            • tools/schema.go (Parameter Validation)
                                                    • plugins/ (Plugin SDK Platform)
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

### 2. Run NEXA AI Desktop Companion

```powershell
.\nexa.exe gui
```

---

## 📄 License

This project is licensed under the **MIT License**.
