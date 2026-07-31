package gui

import (
	"fmt"
	"syscall"
	"unsafe"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	shell32          = syscall.NewLazyDLL("shell32.dll")
	pRegisterHotKey  = user32.NewProc("RegisterHotKey")
	pGetMessageW     = user32.NewProc("GetMessageW")
	pShowWindow      = user32.NewProc("ShowWindow")
	pSetForegroundWindow = user32.NewProc("SetForegroundWindow")

	// Shell_NotifyIconW for system tray
	pShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	pLoadIconW        = user32.NewProc("LoadIconW")

	// Constants
	modCtrl     = 0x0002
	vkSpace     = 0x20
	wmHotkey    = 0x0312

	// Shell_NotifyIcon constants
	nimAdd    = 0x00000000
	nimDelete = 0x00000004
	nifIcon   = 0x00000002
	nifTip    = 0x00000004
)

const hotkeyID = 1

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

// NOTIFYICONDATAW for system tray icon (simplified)
type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
}

// RegisterGlobalHotkey registers Ctrl+Space as a global hotkey to toggle the NEXA window.
func (a *App) RegisterGlobalHotkey() {
	go func() {
		ret, _, err := pRegisterHotKey.Call(0, hotkeyID, uintptr(modCtrl), uintptr(vkSpace))
		if ret == 0 {
			fmt.Printf("⚠️ Failed to register Ctrl+Space hotkey: %v\n", err)
			return
		}
		fmt.Println("⌨️  Global hotkey Ctrl+Space registered")

		var m msg
		for {
			ret, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
			if ret == 0 {
				break
			}
			if m.Message == uint32(wmHotkey) && m.WParam == hotkeyID {
				a.toggleWindow()
			}
		}
	}()
}

var windowVisible = true

func (a *App) toggleWindow() {
	if a.ctx == nil {
		return
	}
	if windowVisible {
		wailsRuntime.WindowHide(a.ctx)
		windowVisible = false
	} else {
		wailsRuntime.WindowShow(a.ctx)
		windowVisible = true
	}
}

// AddTrayIcon adds a basic system tray icon using Shell_NotifyIconW.
func (a *App) AddTrayIcon() {
	pGetModuleHandleW := user32.NewProc("GetModuleHandleW")

	go func() {
		// Load NEXA's embedded EXE icon handle from current process instance
		hInstance, _, _ := pGetModuleHandleW.Call(0)
		hIcon, _, _ := pLoadIconW.Call(hInstance, uintptr(1)) // Embedded MAIN icon resource ID #1
		if hIcon == 0 {
			// Fallback to IDI_APPLICATION if resource handle unavailable
			hIcon, _, _ = pLoadIconW.Call(0, uintptr(32512))
		}

		var nid notifyIconData
		nid.CbSize = uint32(unsafe.Sizeof(nid))
		nid.UID = 1
		nid.UFlags = uint32(nifIcon | nifTip)
		nid.HIcon = hIcon

		// Set tooltip "NEXA — AI Desktop Companion"
		tip := "NEXA — AI Desktop Companion"
		tipUTF16, _ := syscall.UTF16FromString(tip)
		copy(nid.SzTip[:], tipUTF16)

		pShellNotifyIconW.Call(uintptr(nimAdd), uintptr(unsafe.Pointer(&nid)))
		fmt.Println("🔔 System tray icon added (NEXA Custom Icon)")
	}()
}

// RemoveTrayIcon removes the system tray icon.
func (a *App) RemoveTrayIcon() {
	var nid notifyIconData
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.UID = 1
	pShellNotifyIconW.Call(uintptr(nimDelete), uintptr(unsafe.Pointer(&nid)))
}
