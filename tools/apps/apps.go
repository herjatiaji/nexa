package apps

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// AppsTool provides Windows Desktop application launching, control, listing, and window focus management.
type AppsTool struct{}

// New creates a new AppsTool.
func New() *AppsTool {
	return &AppsTool{}
}

func (a *AppsTool) Name() string {
	return "desktop_apps"
}

func (a *AppsTool) Description() string {
	return "Control desktop applications on the user's computer. Supported actions: " +
		"'launch' to open/start an application (e.g., code, chrome, notepad, spotify, docker, explorer, calc, msedge), " +
		"'close' to terminate a running application process, " +
		"'list' to list all currently running GUI desktop applications and processes, " +
		"'focus' to bring a running application window to the foreground."
}

func (a *AppsTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Desktop application action to perform",
				"enum":        []interface{}{"launch", "close", "list", "focus"},
			},
			"app_name": map[string]interface{}{
				"type":        "string",
				"description": "Application name, executable name, or path (e.g., code, chrome, notepad, spotify, calc, explorer)",
			},
			"arguments": map[string]interface{}{
				"type":        "string",
				"description": "Optional arguments/URL/file path to pass when launching the application (e.g., URL for browser or folder path for VS Code)",
			},
		},
		"required": []interface{}{"action"},
	}
}

type appsInput struct {
	Action    string `json:"action"`
	AppName   string `json:"app_name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

func (a *AppsTool) Execute(input string) (string, error) {
	var params appsInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("desktop_apps tool is currently optimized for Windows OS")
	}

	switch params.Action {
	case "launch":
		return a.launchApp(params.AppName, params.Arguments)
	case "close":
		return a.closeApp(params.AppName)
	case "list":
		return a.listRunningApps()
	case "focus":
		return a.focusApp(params.AppName)
	default:
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}
}

func (a *AppsTool) launchApp(appName, args string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("app_name is required for launch action")
	}

	cleanApp := strings.TrimSpace(strings.ToLower(appName))
	cleanApp = strings.TrimSuffix(cleanApp, ".exe")

	psScript := fmt.Sprintf(`
$name = "%s"
$args = "%s"

# Special Layer 0: File Manager, Partitions & Windows Folder Shortcuts
if ($name -eq "file manager" -or $name -eq "explorer" -or $name -eq "files" -or $name -eq "this pc" -or $name -eq "my computer") {
	try {
		if ($args -ne "") {
			Start-Process "explorer.exe" -ArgumentList $args -ErrorAction Stop
		} else {
			Start-Process "explorer.exe" -ErrorAction Stop
		}
		"OK"
		exit 0
	} catch {}
}

# Handle Drive Partitions (e.g. D, E, D:, E:, drive d, drive e, partition d)
if ($name -match '^(drive|partition)?\s*([a-zA-Z]):?$' -or $args -match '^[a-zA-Z]:\\?$') {
	try {
		$targetDrive = $name
		if ($args -ne "") { $targetDrive = $args }
		$targetDrive = $targetDrive -replace '(?i)drive|partition|\s', ''
		if ($targetDrive.Length -eq 1) { $targetDrive = $targetDrive + ":" }
		$targetPath = $targetDrive + "\"
		if (Test-Path $targetPath) {
			Start-Process "explorer.exe" -ArgumentList $targetPath -ErrorAction Stop
			"OK"
			exit 0
		}
	} catch {}
}

$folderMap = @{
	'documents' = "$env:USERPROFILE\Documents"
	'downloads' = "$env:USERPROFILE\Downloads"
	'desktop'   = "$env:USERPROFILE\Desktop"
	'pictures'  = "$env:USERPROFILE\Pictures"
	'videos'    = "$env:USERPROFILE\Videos"
	'music'     = "$env:USERPROFILE\Music"
}
if ($folderMap.ContainsKey($name)) {
	try {
		Start-Process "explorer.exe" -ArgumentList $folderMap[$name] -ErrorAction Stop
		"OK"
		exit 0
	} catch {}
}

# Layer 1: App Protocol URIs
$uriMap = @{
	'spotify' = 'spotify:'
	'calculator' = 'calculator:'
	'calc' = 'calculator:'
	'settings' = 'ms-settings:'
	'store' = 'ms-windows-store:'
	'mail' = 'outlookmail:'
	'calendar' = 'outlookcal:'
	'photos' = 'ms-photos:'
	'camera' = 'microsoft.windows.camera:'
}

if ($uriMap.ContainsKey($name)) {
	try {
		Start-Process $uriMap[$name] -ErrorAction Stop
		"OK"
		exit 0
	} catch {}
}

# Layer 2: Standard Executable Launch (detached)
try {
	if ($args -ne "") {
		Start-Process -FilePath $name -ArgumentList $args -WindowStyle Normal -ErrorAction Stop
	} else {
		Start-Process -FilePath $name -WindowStyle Normal -ErrorAction Stop
	}
	"OK"
	exit 0
} catch {}

# Layer 3: Executable with .exe extension
try {
	$exeName = "$name.exe"
	if ($args -ne "") {
		Start-Process -FilePath $exeName -ArgumentList $args -WindowStyle Normal -ErrorAction Stop
	} else {
		Start-Process -FilePath $exeName -WindowStyle Normal -ErrorAction Stop
	}
	"OK"
	exit 0
} catch {}

# Layer 4: Windows Start Apps Database (shell:AppsFolder) for UWP/Microsoft Store apps
try {
	$app = Get-StartApps | Where-Object { $_.Name -like "*$name*" -or $_.AppID -like "*$name*" } | Select-Object -First 1
	if ($app) {
		Start-Process "shell:AppsFolder\$($app.AppID)" -ErrorAction Stop
		"OK"
		exit 0
	}
} catch {}

# Layer 5: AppData WindowsApps path fallback
$userApps = "$env:LOCALAPPDATA\Microsoft\WindowsApps\$name.exe"
if (Test-Path $userApps) {
	Start-Process -FilePath $userApps -ErrorAction Stop
	"OK"
	exit 0
}

throw "Application '$name' could not be launched via URI, Executable, or Start Menu database."
`, cleanApp, args)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to launch %s: %s", appName, strings.TrimSpace(string(out)))
	}

	if args != "" {
		return fmt.Sprintf("🚀 Successfully launched %s with arguments: %s", appName, args), nil
	}
	return fmt.Sprintf("🚀 Successfully launched application: %s", appName), nil
}

func (a *AppsTool) closeApp(appName string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("app_name is required for close action")
	}

	cleanName := strings.TrimSuffix(strings.ToLower(appName), ".exe")

	psScript := fmt.Sprintf(`
$procs = Get-Process -Name "*%s*" -ErrorAction SilentlyContinue
if ($procs) {
    $procs | ForEach-Object { $_.CloseMainWindow(); Start-Sleep -Milliseconds 200; if (-not $_.HasExited) { Stop-Process -Id $_.Id -Force } }
    "Closed $($procs.Count) process(es) matching %s"
} else {
    "No running process found matching %s"
}
`, cleanName, cleanName, cleanName)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to close %s: %v", appName, err)
	}

	return strings.TrimSpace(string(out)), nil
}

func (a *AppsTool) listRunningApps() (string, error) {
	psScript := `
Get-Process | Where-Object { $_.MainWindowTitle -ne "" } | 
    Select-Object ProcessName, Id, @{Name="RAM_MB"; Expression={[math]::Round($_.WorkingSet64 / 1MB, 1)}}, MainWindowTitle | 
    Sort-Object ProcessName | 
    Format-Table -AutoSize | Out-String
`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list running apps: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return "No active GUI desktop applications detected.", nil
	}

	return fmt.Sprintf("🖥️ Active Desktop Applications:\n\n%s", result), nil
}

func (a *AppsTool) focusApp(appName string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("app_name is required for focus action")
	}

	cleanName := strings.TrimSuffix(strings.ToLower(appName), ".exe")

	psScript := fmt.Sprintf(`
$code = @"
using System;
using System.Runtime.InteropServices;
public class Win32 {
    [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr hWnd);
    [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
}
"@
Add-Type -TypeDefinition $code -ErrorAction SilentlyContinue

$proc = Get-Process | Where-Object { $_.ProcessName -like "*%s*" -and $_.MainWindowHandle -ne 0 } | Select-Object -First 1
if ($proc) {
    [Win32]::ShowWindow($proc.MainWindowHandle, 9)
    [Win32]::SetForegroundWindow($proc.MainWindowHandle)
    "Focused window for process: $($proc.ProcessName)"
} else {
    "No window handle found for process matching %s"
}
`, cleanName, cleanName)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to focus %s: %v", appName, err)
	}

	return strings.TrimSpace(string(out)), nil
}

func normalizeAppName(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch lower {
	case "vscode", "vs code", "code":
		return "code"
	case "chrome", "google chrome":
		return "chrome"
	case "edge", "msedge", "microsoft edge":
		return "msedge"
	case "notepad":
		return "notepad"
	case "calc", "calculator":
		return "calc"
	case "explorer", "file explorer":
		return "explorer"
	case "docker", "docker desktop":
		return "docker"
	case "spotify":
		return "spotify"
	case "terminal", "powershell":
		return "powershell"
	case "cmd", "command prompt":
		return "cmd"
	default:
		return name
	}
}
