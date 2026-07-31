# 🤖 NEXA v3.1.1 — Personal Cognitive Operating System & Observability Platform

[![Release](https://img.shields.io/badge/Version-v3.1.1-success?style=flat-square)](https://github.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react)](https://reactjs.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io/)
[![SQLite](https://img.shields.io/badge/Database-SQLite-003B57?style=flat-square&logo=sqlite)](https://sqlite.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.style=flat-square)](LICENSE)

**NEXA v3.1.1** is a **Personal Cognitive Operating System & AI Desktop Companion** built with Go, Wails v2, React, SQLite, and Python. It features a **Headless Background Daemon**, **Modular Brain Kernel**, **Redux-Style State Reducer**, **Perception Fusion Engine**, **Adaptive Cycle Scheduler**, **Policy Rule Engine**, **Weighted Decision Arbiter**, **Cognitive Trace Pipeline**, **Decision Explainer**, **React Brain Inspector UI**, **4-Level Autonomy Controller**, **Speech Intelligence Layer**, and **Windows Desktop Observer**.

---

## ✨ Execution Modes in v3.1.1

NEXA can run in 5 specialized execution modes:

| Command | Interface | Description |
| :--- | :--- | :--- |
| **`.\nexa.exe dashboard`** | 📊 **Full Dashboard Window** | Complete management UI (Chat stream, Memory editor, Brain Inspector, Dev Traces, Model router) |
| **`.\nexa.exe avatar`** | 🤖 **Floating Mascot Widget** | 320x320 transparent, frameless, always-on-top mascot widget on desktop |
| **`.\nexa.exe gui`** | 🚀 **Desktop Companion** | Default launcher (starts background daemon + floating mascot widget) |
| **`.\nexa.exe daemon`** | ⚙️ **Background Service** | Headless runtime (Brain Kernel, Voice listener, Memory, Agent engine, Observer, IPC server) |
| **`.\nexa.exe autostart`** | 🔌 **Windows Autostart** | Registers NEXA daemon to boot automatically on Windows startup (`HKCU\...\Run`) |

---

## 🧠 Cognitive Architecture & Observability (v3.1 & v3.1.1)

```
                         Raw Events (Voice, Vision, Desktop)
                                      │
                              Event Priority Queue
                              (events/priority.go)
                                      │
                              Perception Pipeline
                           (perception/perceiver.go)
                                      │
                           Perception Fusion Engine
                             (perception/fusion.go)
                                      │
                             Brain State Reducer
                              (brain/reducer.go)
                                      │
                             Brain Snapshot (immutable)
                                      │
              ┌───────────────────────┼───────────────────────┐
              │                       │                       │
      Working Memory           Context Module          Goal Manager
  (cognitive/working_memory.go) (cognitive/context.go)  (goals/manager.go)
              │                       │                       │
              └───────────────────────┼───────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
              Policy Engine    Social Module     Autonomy Controller
         (cognitive/policy.go) (cognitive/social.go) (autonomy/controller.go)
                    │                 │                 │
                    └─────────────────┼─────────────────┘
                                      │
                              Decision Arbiter
                              (brain/arbiter.go)
                                      │
                        Decision Explainer & Telemetry
                       (cognitive/explain/ & trace/)
                                      │
                        Dashboard Brain Inspector UI
                     (gui/frontend/src/components/)
```

### Key Cognitive Features:

- 🧠 **Brain Kernel (`brain/kernel.go`)**: Central orchestrator managing the cognitive tick cycle with Redux-style atomic state reduction (`brain/reducer.go`) — modules propose state changes without direct mutation.
- ⚡ **Adaptive Cognitive Cycle (`brain/cycle.go`)**: Dynamic cycle scheduling adjusting frequency based on user activity:
  - `ACTIVE` / Voice Dialogue: `200ms` – `500ms` near-realtime
  - `NORMAL` Work: `2s` standard tick
  - `IDLE`: `10s` relaxed tick
  - `SILENT` / Deep Idle: `30s` CPU-saving sleep
- 👁️ **Perception Fusion Engine (`perception/fusion.go`)**: Resolves conflicting inputs from Desktop (`user32.dll`), Voice, and Vision into unified semantic percepts with weighted multi-source confidence fusion.
- 📜 **Cognitive Trace Pipeline & Event Middleware (`cognitive/trace/`, `events/middleware.go`)**: Intercepts and records cycle phase traces (`perceive`, `update`, `think`, `decide`, `execute`) and system event execution durations.
- 💬 **Decision Explainer (`cognitive/explain/`)**: Translates decision scores, active rules, social boundary scores, and autonomy constraints into natural language explanations (*"NEXA memilih diam karena Anda sedang mengetik di VS Code (Perhatian: 91%)"*).
- 🔍 **React Brain Inspector Dashboard (`components/BrainInspector.tsx`)**: Interactive modal inspector accessible via the **🧠 BRAIN** top-bar button:
  - **Current State Card**: Active app, activity state, attention score %, autonomy level, tick latency.
  - **Decision Explainer Box**: Human-readable reasoning and decision factor breakdown.
  - **Real-Time Cognitive Trace Stream**: Step-by-step visual pipeline execution log.
- 🛡️ **Autonomy Controller & Budget (`autonomy/controller.go`)**: 4-level user-configurable autonomy (`Passive`, `Suggest`, `Assist`, `Agent`) with hourly action limits.
- 🗄️ **SQLite Telemetry Storage (`memory/db.go`)**: Persistent logging of `brain_events`, `cognitive_traces`, and `brain_snapshots` in `nexa_data.db`.

---

## 🌟 System Features

- 💎 **Native Custom EXE & System Tray Icon (`cmd/nexa/winres/`)**: Embedded Solarpunk Aqua-Emerald glass sphere mascot icon compiled into `nexa.exe` binary and active Windows System Tray.
- 📊 **Full Management Dashboard (`.\nexa.exe dashboard`)**: Full-featured desktop window with interactive Chat, Memory manager, Brain Inspector, Dev Traces drawer, and OS Security Permissions modal.
- 🤖 **Transparent Floating Mascot Avatar (`.\nexa.exe avatar`)**: Frameless, transparent, always-on-top 320x320 mascot widget that floats directly on your desktop.
- 🎙️ **NEXA Speech Intelligence Layer (`voice/speech/`)**: Expressive speech planner with lexical emotion detection, prosody control, British professional persona, and pure Go WAV post-processing (-3dBFS peak normalization, silence trimming, crossfading).
- 🖥️ **Windows Desktop Observer (`observer/desktop.go`)**: Monitors active focused window (`user32.dll`) without high CPU load.

---

## 🛠️ Prerequisites & Setup

1. **Go 1.20+**, **Node.js 18+**, **Wails v2 CLI**, **Python 3.10+**.

### Build Executable

```powershell
# 1. Build React frontend
cd gui/frontend; npm install; npx tsc; npx vite build; cd ../..

# 2. Build NEXA executable with embedded icon & Cognitive Foundation
go build -tags desktop,production -o nexa.exe ./cmd/nexa/

# 3. Launch NEXA Dashboard (Full Window with Brain Inspector)
.\nexa.exe dashboard

# 4. Launch NEXA Mascot Companion (Floating Widget)
.\nexa.exe gui
```

---

## 📄 License

This project is licensed under the **MIT License**.
