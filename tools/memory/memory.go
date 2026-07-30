package memorytool

import (
	"encoding/json"
	"fmt"

	"github.com/heraji/jarvis/memory"
)

// MemoryTool provides long-term memory operations for NEXA.
type MemoryTool struct {
	store *memory.MemoryStore
}

// New creates a new MemoryTool instance.
func New(store *memory.MemoryStore) *MemoryTool {
	return &MemoryTool{store: store}
}

func (m *MemoryTool) Name() string {
	return "memory"
}

func (m *MemoryTool) Description() string {
	return "Manage long-term memory & user facts. Supported actions: " +
		"'store' (save a fact/preference about the user, e.g. key='main_project', value='Kazeer'), " +
		"'get' (retrieve a saved fact by key), " +
		"'list' (view all remembered user facts), " +
		"'delete' (remove a remembered fact)."
}

func (m *MemoryTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Memory action: 'store', 'get', 'list', or 'delete'",
				"enum":        []interface{}{"store", "get", "list", "delete"},
			},
			"key": map[string]interface{}{
				"type":        "string",
				"description": "Fact key identifier (e.g., 'main_project', 'preferred_language', 'backend_tech')",
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "Fact value to store (required for 'store' action)",
			},
		},
		"required": []interface{}{"action"},
	}
}

type memoryInput struct {
	Action string `json:"action"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
}

func (m *MemoryTool) Execute(input string) (string, error) {
	var params memoryInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	switch params.Action {
	case "store":
		if params.Key == "" || params.Value == "" {
			return "", fmt.Errorf("key and value are required for store action")
		}
		if err := m.store.Set(params.Key, params.Value); err != nil {
			return "", fmt.Errorf("failed to store fact: %w", err)
		}
		return fmt.Sprintf("🧠 Remembered: %s = %s", params.Key, params.Value), nil

	case "get":
		if params.Key == "" {
			return "", fmt.Errorf("key is required for get action")
		}
		val, found := m.store.Get(params.Key)
		if !found {
			return fmt.Sprintf("No fact remembered for key %q", params.Key), nil
		}
		return fmt.Sprintf("Fact %s: %s", params.Key, val), nil

	case "list":
		facts := m.store.List()
		if len(facts) == 0 {
			return "No long-term memories stored yet.", nil
		}
		var lines []string
		for k, v := range facts {
			lines = append(lines, fmt.Sprintf("• %s: %s", k, v))
		}
		return fmt.Sprintf("🧠 Remembered User Facts:\n\n%s", fmt.Sprintf("%s", lines)), nil

	case "delete":
		if params.Key == "" {
			return "", fmt.Errorf("key is required for delete action")
		}
		if err := m.store.Delete(params.Key); err != nil {
			return "", fmt.Errorf("failed to delete fact: %w", err)
		}
		return fmt.Sprintf("Forgotten fact for key %q", params.Key), nil

	default:
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}
}
