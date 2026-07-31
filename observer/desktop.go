package observer

import (
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/heraji/jarvis/events"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	pGetForegroundWindow         = user32.NewProc("GetForegroundWindow")
	pGetWindowTextW              = user32.NewProc("GetWindowTextW")
	pGetWindowThreadProcessId    = user32.NewProc("GetWindowThreadProcessId")
)

// ActiveWindowInfo captures details of the currently focused desktop window.
type ActiveWindowInfo struct {
	Title     string `json:"title"`
	AppName   string `json:"appName"`
	Timestamp string `json:"timestamp"`
}

// DesktopObserver tracks the currently focused active window on Windows
// using user32.dll without heavy CPU polling.
type DesktopObserver struct {
	bus            *events.EventBus
	lastTitle      string
	lastApp        string
	mu             sync.Mutex
	stopChan       chan struct{}
	running        bool
	OnWindowChange func(info ActiveWindowInfo)
}

// NewDesktopObserver creates a DesktopObserver bound to the system EventBus.
func NewDesktopObserver(bus *events.EventBus) *DesktopObserver {
	return &DesktopObserver{
		bus:      bus,
		stopChan: make(chan struct{}),
	}
}

// Start begins monitoring active window changes in a background goroutine.
func (o *DesktopObserver) Start() {
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return
	}
	o.running = true
	o.stopChan = make(chan struct{})
	o.mu.Unlock()

	go o.loop()
}

// Stop terminates the observation loop.
func (o *DesktopObserver) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.running {
		return
	}
	o.running = false
	close(o.stopChan)
}

func (o *DesktopObserver) loop() {
	ticker := time.NewTicker(1500 * time.Millisecond) // Check every 1.5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-o.stopChan:
			return
		case <-ticker.C:
			info := o.GetActiveWindow()
			if info.Title == "" && info.AppName == "" {
				continue
			}

			o.mu.Lock()
			changed := info.Title != o.lastTitle || info.AppName != o.lastApp
			if changed {
				o.lastTitle = info.Title
				o.lastApp = info.AppName
			}
			o.mu.Unlock()

			if changed {
				// Emit window.changed event to event bus
				if o.bus != nil {
					o.bus.Emit(events.EventWindowChanged, info, "desktop_observer")
				}
				if o.OnWindowChange != nil {
					o.OnWindowChange(info)
				}
			}
		}
	}
}

// GetActiveWindow inspects Windows user32 APIs to retrieve current focused window title & app.
func (o *DesktopObserver) GetActiveWindow() ActiveWindowInfo {
	if runtime.GOOS != "windows" {
		return ActiveWindowInfo{Title: "Desktop Environment", AppName: "System"}
	}

	hwnd, _, _ := pGetForegroundWindow.Call()
	if hwnd == 0 {
		return ActiveWindowInfo{}
	}

	// Get Window Title
	b := make([]uint16, 512)
	ret, _, _ := pGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	title := ""
	if ret > 0 {
		title = syscall.UTF16ToString(b[:ret])
	}

	// Derive application name from window title or process
	appName := deriveAppName(title)

	return ActiveWindowInfo{
		Title:     title,
		AppName:   appName,
		Timestamp: time.Now().Format("15:04:05"),
	}
}

// deriveAppName parses common application names from window titles.
func deriveAppName(title string) string {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "visual studio code") || strings.Contains(lower, "vscode") || strings.HasSuffix(lower, "- code"):
		return "Visual Studio Code"
	case strings.Contains(lower, "chrome"):
		return "Google Chrome"
	case strings.Contains(lower, "firefox"):
		return "Mozilla Firefox"
	case strings.Contains(lower, "edge"):
		return "Microsoft Edge"
	case strings.Contains(lower, "spotify"):
		return "Spotify"
	case strings.Contains(lower, "discord"):
		return "Discord"
	case strings.Contains(lower, "terminal") || strings.Contains(lower, "powershell") || strings.Contains(lower, "cmd"):
		return "Windows Terminal"
	case strings.Contains(lower, "sublime"):
		return "Sublime Text"
	case strings.Contains(lower, "notepad"):
		return "Notepad"
	case strings.Contains(lower, "docker"):
		return "Docker Desktop"
	default:
		parts := strings.Split(title, " - ")
		if len(parts) > 1 {
			return parts[len(parts)-1]
		}
		return title
	}
}
