package rules

import (
	"fmt"
	"strings"

	"github.com/heraji/jarvis/cognitive"
)

// DockerRule triggers proactive assistance suggestions when build or docker errors repeat.
type DockerRule struct{}

func NewDockerRule() *DockerRule {
	return &DockerRule{}
}

func (r *DockerRule) Name() string {
	return "docker_troubleshooting_rule"
}

func (r *DockerRule) Evaluate(ctx cognitive.CognitiveContext, memory cognitive.WorkingMemory) *cognitive.CognitiveRuleResult {
	if !ctx.ProblemDetected && ctx.Activity != "troubleshooting" {
		return nil
	}

	errorCount := 0
	var lastErrDesc string
	for _, p := range memory.OpenProblems {
		if !p.Resolved && (strings.Contains(strings.ToLower(p.Description), "docker") || strings.Contains(strings.ToLower(p.Description), "failed") || strings.Contains(strings.ToLower(p.Description), "error")) {
			errorCount++
			lastErrDesc = p.Description
		}
	}

	if errorCount >= 2 {
		return &cognitive.CognitiveRuleResult{
			Action:     "TOAST",
			Confidence: 0.85,
			Reason:     fmt.Sprintf("Detected %d build/docker issues recently", errorCount),
			Payload: map[string]string{
				"message": fmt.Sprintf("I noticed repeated build errors (%s). Would you like me to inspect logs?", lastErrDesc),
			},
		}
	}

	return nil
}
