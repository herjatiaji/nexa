package gui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// LaunchGUI opens the NEXA Dashboard in a standalone app window.
func LaunchGUI() error {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	htmlPath := filepath.Join(cwd, "gui", "index.html")

	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		return fmt.Errorf("GUI file not found at %s", htmlPath)
	}

	if runtime.GOOS == "windows" {
		// Launch as standalone Chromium / Edge App Window
		appURL := fmt.Sprintf("file:///%s", filepath.ToSlash(htmlPath))
		cmd := exec.Command("msedge.exe", fmt.Sprintf("--app=%s", appURL), "--window-size=1280,800")
		if err := cmd.Start(); err != nil {
			// Fallback: Open with default Windows web browser
			fallbackCmd := exec.Command("cmd", "/c", "start", appURL)
			return fallbackCmd.Start()
		}
		return nil
	}

	// Fallback for Linux/macOS
	return exec.Command("open", htmlPath).Start()
}
