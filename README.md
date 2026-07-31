# 🤖 NEXA v3.0 — AI Desktop Companion & Multi-Mode Platform

[![Release](https://img.shields.io/badge/Version-v3.0.0-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react)](https://reactjs.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io/)
[![SQLite](https://img.shields.io/badge/Database-SQLite-003B57?style=flat-square&logo=sqlite)](https://sqlite.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v3.0.0** is a **Living AI Desktop Companion Platform** built with Go, Wails v2, React, SQLite, and Python. It features a **Headless Background Daemon**, **Transparent Floating Mascot Avatar Window**, **Full Management Dashboard UI**, **IPC Event Bus (`\\.\pipe\nexa`)**, **Windows Desktop Observer (`user32.dll`)**, **Adaptive Intelligent Vision Loop**, **Awareness State Machine (`ACTIVE`, `PASSIVE`, `SILENT`)**, **Speech Intelligence Layer**, **Native Embedded EXE & System Tray Icon**, and **Windows Autostart Integration**.

---

## ✨ Execution Modes in v3.0.0

NEXA can run in 4 specialized execution modes:

| Command | Interface | Description |
| :--- | :--- | :--- |
| **`.\nexa.exe dashboard`** | 📊 **Full Dashboard Window** | Complete management UI (Chat stream, Memory editor, Traces inspector, Model router) |
| **`.\nexa.exe avatar`** | 🤖 **Floating Mascot Widget** | 320x320 transparent, frameless, always-on-top mascot widget on desktop |
| **`.\nexa.exe gui`** | 🚀 **Desktop Companion** | Default launcher (starts headless background service + floating mascot widget) |
| **`.\nexa.exe daemon`** | ⚙️ **Background Service** | Headless runtime (Voice listener, Memory, Agent engine, Observer, IPC server) |
| **`.\nexa.exe autostart`** | 🔌 **Windows Autostart** | Registers NEXA daemon to boot automatically on Windows startup (`HKCU\...\Run`) |

---

## 🌟 Key Features

- 💎 **Native Custom EXE & System Tray Icon (`cmd/nexa/winres/`)**: Embedded Solarpunk Aqua-Emerald glass sphere mascot icon compiled into `nexa.exe` binary and active Windows System Tray.
- 📊 **Full Management Dashboard (`.\nexa.exe dashboard`)**: Full-featured 1280x820 desktop window with interactive Chat, Memory manager, Dev Traces drawer, and OS Security Permissions modal.
- 🤖 **Transparent Floating Mascot Avatar (`.\nexa.exe avatar`)**: Frameless, transparent, always-on-top 320x320 mascot widget that floats directly on your desktop. Drag the mascot anywhere.
- ⚙️ **NEXA Background Daemon (`.\nexa.exe daemon`)**: Headless background service managing openWakeWord, Whisper STT, SQLite Memory, Agent ReAct loop, Tools, Desktop Observer, and IPC server (`127.0.0.1:59123`).
- 📡 **IPC Event Bus (`events/bus.go`, `events/pipe.go`)**: Thread-safe pub/sub EventBus and Windows Named Pipe server for real-time inter-process communication.
- 🖥️ **Windows Desktop Observer (`observer/desktop.go`)**: Monitors focused window (`user32.dll`) without high CPU load. Emits `window.changed` events (*"Focused: Visual Studio Code"*).
- 👁️ **Adaptive Vision Loop (`observer/vision.go`)**: Screen observation loop with dynamic sampling rate:
  - `SILENT` / `IDLE`: 1 capture / 60 seconds
  - `PASSIVE`: 1 capture / 30 seconds
  - `ACTIVE`: 1 capture / 5 seconds (or immediately on window change)
- 🧠 **Awareness State Machine (`observer/awareness.go`)**: Manages `ACTIVE`, `PASSIVE`, and `SILENT` modes so NEXA observes desktop activity quietly without making annoying or aggressive interruptions.
- 🎙️ **NEXA Speech Intelligence Layer (`voice/speech/`)**: Multi-stage expressive speech planner with lexical emotion detection, prosody control, British professional persona, and pure Go WAV post-processing (-3dBFS peak normalization, silence trimming, crossfading).

---

## 🏗️ System Architecture

```
                 Windows Boot
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
Voice Listener                 Floating Mascot / Dashboard
Memory                         State Machine Animations
Tools                          Lip Sync / Toast Bubbles
Observer / Vision              Event Bus IPC Server
```

---

## 🛠️ Prerequisites & Setup

1. **Go 1.20+**, **Node.js 18+**, **Wails v2 CLI**, **Python 3.10+**.

### Build Executable

```powershell
# 1. Build React frontend
cd gui/frontend; npm install; npx tsc; npx vite build; cd ../..

# 2. Build NEXA executable with embedded icon
go build -tags desktop,production -o nexa.exe ./cmd/nexa/

# 3. Launch NEXA Dashboard (Full Window)
.\nexa.exe dashboard

# 4. Launch NEXA Mascot Companion (Floating Widget)
.\nexa.exe gui
```

---

## 📄 License

This project is licensed under the **MIT License**.
