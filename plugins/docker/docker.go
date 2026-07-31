package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/heraji/jarvis/config"
	"github.com/heraji/jarvis/tools"
)

// DockerPlugin provides Docker container management capabilities to NEXA.
type DockerPlugin struct{}

func NewDockerPlugin() *DockerPlugin {
	return &DockerPlugin{}
}

func (dp *DockerPlugin) Name() string        { return "nexa-plugin-docker" }
func (dp *DockerPlugin) Description() string { return "Docker container management plugin for NEXA" }
func (dp *DockerPlugin) Version() string     { return "1.0.0" }
func (dp *DockerPlugin) Init(cfg *config.Config) error { return nil }

func (dp *DockerPlugin) Tools() []tools.Tool {
	return []tools.Tool{&DockerTool{}}
}

// DockerTool executes docker commands.
type DockerTool struct{}

func (dt *DockerTool) Name() string { return "docker" }
func (dt *DockerTool) Description() string {
	return "Manage Docker containers (ps, inspect, logs, start, stop)"
}

func (dt *DockerTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Docker action to perform (ps, logs, start, stop, inspect)",
				"enum":        []string{"ps", "logs", "start", "stop", "inspect"},
			},
			"container": map[string]interface{}{
				"type":        "string",
				"description": "Target container name or ID",
			},
		},
		"required": []string{"action"},
	}
}

type dockerArgs struct {
	Action    string `json:"action"`
	Container string `json:"container"`
}

func (dt *DockerTool) Execute(argsJSON string) (string, error) {
	var args dockerArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	var cmd *exec.Cmd
	switch args.Action {
	case "ps":
		cmd = exec.Command("docker", "ps", "-a", "--format", "table {{.ID}}\t{{.Names}}\t{{.Status}}")
	case "logs":
		if args.Container == "" {
			return "", fmt.Errorf("container parameter required for action 'logs'")
		}
		cmd = exec.Command("docker", "logs", "--tail", "20", args.Container)
	case "start":
		if args.Container == "" {
			return "", fmt.Errorf("container parameter required for action 'start'")
		}
		cmd = exec.Command("docker", "start", args.Container)
	case "stop":
		if args.Container == "" {
			return "", fmt.Errorf("container parameter required for action 'stop'")
		}
		cmd = exec.Command("docker", "stop", args.Container)
	case "inspect":
		if args.Container == "" {
			return "", fmt.Errorf("container parameter required for action 'inspect'")
		}
		cmd = exec.Command("docker", "inspect", "--format", "{{.State.Status}}", args.Container)
	default:
		return "", fmt.Errorf("unknown docker action: %s", args.Action)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Docker command failed: %v\nOutput: %s", err, string(out)), nil
	}

	return string(out), nil
}
