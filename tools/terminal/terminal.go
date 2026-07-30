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

type RiskLevel int

const (
	RiskAllow RiskLevel = iota
	RiskConfirm
	RiskDeny
)

var (
	// Critical destructive commands that are strictly FORBIDDEN (DENY)
	denyPatterns = []string{
		"rm -rf /",
		"rm -rf c:",
		"rm -rf c:\\",
		"format c:",
		"diskpart",
		"rd /s /q c:",
		"del /f /s /q c:",
		":(){ :|:& };:",
		"mkfs",
		"dd if=",
	}

	// Medium/High risk commands that require user confirmation (CONFIRM)
	confirmPatterns = []string{
		"rm -rf",
		"rm -r",
		"rmdir",
		"remove-item",
		"del /f",
		"del /s",
		"format",
		"shutdown",
		"reboot",
		"reg delete",
		"stop-process",
		"taskkill",
		"git reset --hard",
		"git push --force",
	}

	// Safe read-only commands that bypass confirmation (ALLOW)
	allowPrefixes = []string{
		"dir", "ls", "pwd", "cd", "echo", "whoami", "ipconfig", "ifconfig",
		"git status", "git log", "git diff", "git branch", "git show",
		"docker ps", "docker images", "node -v", "npm -v", "go version", "go env",
		"get-childitem", "get-process", "get-service", "get-location",
	}
)

// TerminalTool executes shell commands on the user's system.
type TerminalTool struct {
	// ConfirmFunc is called before executing dangerous commands.
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
	return "Execute a shell command on the user's system. " +
		"Command Risk Analyzer categorizes commands into ALLOW (safe), CONFIRM (requires prompt), and DENY (blocked)."
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
				"description": "Optional working directory for the command.",
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

	// Analyze Command Risk
	risk := analyzeRisk(params.Command)
	switch risk {
	case RiskDeny:
		return fmt.Sprintf("⛔ Command DENIED: Execution of %q is blocked by Risk Analyzer for system protection.", params.Command), nil

	case RiskConfirm:
		if t.ConfirmFunc == nil || !t.ConfirmFunc(params.Command) {
			return fmt.Sprintf("⚠️ Command CANCELLED: Execution of %q requires user confirmation.", params.Command), nil
		}

	case RiskAllow:
		// Safe command, proceed directly
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

// analyzeRisk categorizes shell commands into RiskAllow, RiskConfirm, or RiskDeny.
func analyzeRisk(command string) RiskLevel {
	lower := strings.TrimSpace(strings.ToLower(command))

	// 1. Check DENY list
	for _, pattern := range denyPatterns {
		if strings.Contains(lower, pattern) {
			return RiskDeny
		}
	}

	// 2. Check ALLOW list
	for _, prefix := range allowPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return RiskAllow
		}
	}

	// 3. Check CONFIRM list
	for _, pattern := range confirmPatterns {
		if strings.Contains(lower, pattern) {
			return RiskConfirm
		}
	}

	// Default to RiskConfirm for unknown write/exec commands
	return RiskConfirm
}
