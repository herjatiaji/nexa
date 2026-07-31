# 🤖 NEXA v3.0 — AI Desktop Companion & Multi-Mode Platform

[![Release](https://img.shields.io/badge/Version-v3.0.0-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react)](https://reactjs.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io/)
[![SQLite](https://img.shields.io/badge/Database-SQLite-003B57?style=flat-square&logo=sqlite)](https://sqlite.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v3.0.0** evolves NEXA from a windowed application into a **Living AI Desktop Companion Platform**. It features a **Headless Background Daemon**, **Transparent Floating Mascot Avatar Window**, **IPC Event Bus (`\\.\pipe\nexa`)**, **Windows Desktop Observer (`user32.dll`)**, **Adaptive Intelligent Vision Loop**, **Awareness State Machine (`ACTIVE`, `PASSIVE`, `SILENT`)**, **Speech Intelligence Layer**, and **Windows Autostart Integration**.

---

## ✨ Application Modes in v3.0.0

NEXA can run in 4 specialized application modes:

```powershell
# 1. Headless Background Service Daemon (Voice, Memory, Agent, Observer, IPC)
.\nexa.exe daemon

# 2. Transparent Floating Mascot Avatar Window (Frameless & AlwaysOnTop)
.\nexa.exe avatar

# 3. Full Management UI Dashboard (Settings, Memory, Traces, Models)
.\nexa.exe dashboard

# 4. Default Companion Launcher (Background Daemon + Floating Mascot)
.\nexa.exe gui

# 5. Enable Windows Autostart on Boot (HKCU\...\Run)
.\nexa.exe autostart
```

---

## 🌟 Key Architecture Components

- ⚙️ **NEXA Background Daemon (`cmd/nexa/main.go daemon`)**: Runs as a headless Windows background service. Manages openWakeWord, Whisper STT, SQLite Memory, Agent ReAct loop, Tools, Desktop Observer, and IPC.
- 🤖 **Transparent Floating Mascot Avatar (`gui/avatar.go`)**: 320x320 frameless, transparent, always-on-top mascot widget that floats directly on your Windows desktop. Click and drag the mascot anywhere.
- 📡 **NEXA Event System & IPC (`events/bus.go`, `events/pipe.go`)**: Thread-safe internal pub/sub EventBus and Windows Named Pipe server (`\\.\pipe\nexa`) for real-time inter-process communication.
- 🖥️ **Windows Desktop Observer (`observer/desktop.go`)**: Monitors active focused window (`user32.dll`) without high CPU load. Emits `window.changed` events (*"User opened Visual Studio Code"*).
- 👁️ **Adaptive Intelligent Vision Loop (`observer/vision.go`)**: Screen observation loop with dynamic sampling:
  - `SILENT` / `IDLE`: 1 screenshot / 60s
  - `PASSIVE`: 1 screenshot / 30s
  - `ACTIVE`: 1 screenshot / 5s (or immediately on window change)
- 🧠 **Awareness State Machine (`observer/awareness.go`)**: Manages `ACTIVE`, `PASSIVE`, and `SILENT` modes so NEXA observes desktop activity quietly without making annoying or aggressive interruptions.
- 🎙️ **NEXA Speech Intelligence Layer (`voice/speech/`)**: Multi-stage expressive speech planner with lexical emotion detection, prosody control, British professional persona, and pure Go WAV post-processing (-3dBFS peak normalization, silence trimming, crossfading).

---

## 🏗️ High-Level System Architecture

```
                 Windows Startup
                       │
                       │
                 NEXA Runtime
              (Background Daemon)
                       │
        ┌──────────────┴──────────────┐
        │                             │
   Agent Engine                 Avatar Renderer
       Go                         Wails Window
        │                             │
 Voice Listener                 Floating Mascot
 Memory                         State Animations
 Tools                          Lip Sync / Toast
 Observer / Vision              Event Bus IPC
```

---

## 🛠️ Prerequisites & Setup

1. **Go 1.20+**, **Node.js 18+**, **Wails v2 CLI**, **Python 3.10+**.

### Build & Run

```powershell
# 1. Build React frontend
cd gui/frontend; npm install; npx tsc; npx vite build; cd ../..

# 2. Build NEXA executable
go build -tags desktop,production -o nexa.exe ./cmd/nexa/

# 3. Launch NEXA AI Companion
.\nexa.exe gui
```

---

## 📄 License

This project is licensed under the **MIT License**.
