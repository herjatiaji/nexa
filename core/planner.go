package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/heraji/jarvis/types"
)

// TaskStep represents a sub-task created by the Planner.
type TaskStep struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"` // pending, running, completed, failed
	Result      string `json:"result,omitempty"`
}

// Plan represents a complete multi-step task decomposition.
type Plan struct {
	Goal      string     `json:"goal"`
	Steps     []TaskStep `json:"steps"`
	CreatedAt string     `json:"created_at"`
}

// Planner generates structured execution plans for complex user requests.
type Planner struct {
	llm LLM
}

// NewPlanner initializes a new Task Planner using the LLM.
func NewPlanner(llm LLM) *Planner {
	return &Planner{llm: llm}
}

// CreatePlan analyzes a user prompt and breaks it down into structured sub-tasks.
func (p *Planner) CreatePlan(userPrompt string) (*Plan, error) {
	systemPrompt := `You are NEXA's Task Planner. Analyze the user request and break it down into a sequence of clear, ordered sub-tasks.
Respond ONLY with a JSON array of step descriptions, like this:
["Check docker status", "Inspect container logs", "Restart service if failed"]`

	messages := []types.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	resp, err := p.llm.Chat(messages, nil)
	if err != nil {
		return nil, fmt.Errorf("Planner LLM error: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	if idx := strings.Index(content, "["); idx != -1 {
		if lastIdx := strings.LastIndex(content, "]"); lastIdx != -1 && lastIdx > idx {
			content = content[idx : lastIdx+1]
		}
	}

	var stepDescs []string
	if err := json.Unmarshal([]byte(content), &stepDescs); err != nil {
		// Fallback single-step plan
		stepDescs = []string{userPrompt}
	}

	steps := make([]TaskStep, len(stepDescs))
	for i, desc := range stepDescs {
		steps[i] = TaskStep{
			ID:          i + 1,
			Description: desc,
			Status:      "pending",
		}
	}

	return &Plan{
		Goal:      userPrompt,
		Steps:     steps,
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}
