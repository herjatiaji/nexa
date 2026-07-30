package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// dangerousPatterns contains command patterns that require user confirmation.
var dangerousPatterns = []string{
	"rm -rf",
	"rm -r",
	"rmdir",
	"del /f",
	"del /s",
	"format",
	"shutdown",
	"reboot",
	"mkfs",
	"dd if=",
	"> /dev/",
	":(){ :|:& };:",
	"reg delete",
	"diskpart",
}

// TerminalTool executes shell commands on the user's system.
type TerminalTool struct {
	// ConfirmFunc is called before executing dangerous commands.
	// If nil, dangerous commands are blocked.
	ConfirmFunc func(command string) bool
}

// New creates a new TerminalTool.
func New() *TerminalTool {
	return &TerminalTool{}
}

func (t *TerminalTool) Name() string {
	return "run_command"
}

func (t *TerminalTool) Description() string {
	return "Execute a shell command on the user's system and return the output. " +
		"Use this to run programs, check system info, manage processes, or perform any terminal operation. " +
		"The command runs in the system's default shell (PowerShell on Windows, bash on Linux/macOS)."
}

func (t *TerminalTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Optional working directory for the command. Defaults to current directory.",
			},
		},
		"required": []interface{}{"command"},
	}
}

type terminalInput struct {
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir,omitempty"`
}

func (t *TerminalTool) Execute(input string) (string, error) {
	var params terminalInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.Command == "" {
		return "", fmt.Errorf("command is required")
	}

	// Check for dangerous commands
	if isDangerous(params.Command) {
		if t.ConfirmFunc == nil || !t.ConfirmFunc(params.Command) {
			return "⚠️ Command blocked: deemed potentially dangerous. User declined execution.", nil
		}
	}

	// Set up the command with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", params.Command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", params.Command)
	}

	if params.WorkingDir != "" {
		cmd.Dir = params.WorkingDir
	}

	// Capture combined output
	output, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(output))

	// Truncate long output
	const maxLen = 3000
	if len(result) > maxLen {
		result = result[:maxLen] + "\n\n... [output truncated, " + fmt.Sprintf("%d", len(string(output))-maxLen) + " bytes omitted]"
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("Command timed out after 30 seconds.\nPartial output:\n%s", result), nil
		}
		// Return output even on error (e.g., non-zero exit code)
		if result != "" {
			return fmt.Sprintf("Command exited with error: %v\nOutput:\n%s", err, result), nil
		}
		return fmt.Sprintf("Command failed: %v", err), nil
	}

	if result == "" {
		return "Command executed successfully (no output).", nil
	}

	return result, nil
}

// isDangerous checks if a command matches any dangerous patterns.
func isDangerous(command string) bool {
	lower := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
