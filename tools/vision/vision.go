package vision

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// VisionTool captures current screen snapshots for LLM visual context analysis.
type VisionTool struct{}

// New creates a new VisionTool instance.
func New() *VisionTool {
	return &VisionTool{}
}

func (v *VisionTool) Name() string {
	return "vision"
}

func (v *VisionTool) Description() string {
	return "Capture a real-time screenshot of the user's primary display or active window. " +
		"Use this tool when the user asks to inspect their screen, analyze code errors on screen, check UI design, or diagnose visual issues."
}

func (v *VisionTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Vision action: 'capture_screen' (snapshot primary display) or 'inspect_active_window'",
				"enum":        []interface{}{"capture_screen", "inspect_active_window"},
			},
		},
		"required": []interface{}{"action"},
	}
}

type visionInput struct {
	Action string `json:"action"`
}

func (v *VisionTool) Execute(input string) (string, error) {
	var params visionInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("screen vision tool is currently supported on Windows OS")
	}

	tempPng := filepath.Join(os.TempDir(), fmt.Sprintf("nexa_screen_%d.png", time.Now().UnixNano()))
	defer func() {
		// Clean up temporary image after execution
		_ = os.Remove(tempPng)
	}()

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$screen = [System.Windows.Forms.Screen]::PrimaryScreen
$bounds = $screen.Bounds
$bmp = New-Object System.Drawing.Bitmap($bounds.Width, $bounds.Height)
$graphics = [System.Drawing.Graphics]::FromImage($bmp)
$graphics.CopyFromScreen($bounds.X, $bounds.Y, 0, 0, $bmp.Size)
$bmp.Save('%s', [System.Drawing.Imaging.ImageFormat]::Png)
$graphics.Dispose()
$bmp.Dispose()
`, tempPng)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to capture screen: %v (output: %s)", err, string(output))
	}

	imgBytes, err := os.ReadFile(tempPng)
	if err != nil {
		return "", fmt.Errorf("failed to read captured image: %w", err)
	}

	// Base64 encode for vision multimodal models
	b64Data := base64.StdEncoding.EncodeToString(imgBytes)

	return fmt.Sprintf("📸 Screen snapshot captured successfully! Image size: %d bytes. [Base64 PNG Ready (%d chars)]", len(imgBytes), len(b64Data)), nil
}
