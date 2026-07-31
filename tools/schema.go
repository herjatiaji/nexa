package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParameterSchema defines constraints for a single tool parameter.
type ParameterSchema struct {
	Type        string   `json:"type"`                  // string, number, boolean, array, object
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`        // Allowed enum values
	Required    bool     `json:"required,omitempty"`
}

// ToolSchema defines full typed parameter constraints for a tool.
type ToolSchema struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Parameters  map[string]ParameterSchema `json:"parameters"`
	Required    []string                   `json:"required,omitempty"`
}

// ValidateArgs checks raw JSON arguments against the ToolSchema.
func (ts *ToolSchema) ValidateArgs(jsonArgs string) error {
	if jsonArgs == "" || jsonArgs == "{}" {
		if len(ts.Required) > 0 {
			return fmt.Errorf("missing required parameters: %v", ts.Required)
		}
		return nil
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}

	// Check required fields
	for _, reqField := range ts.Required {
		if _, exists := args[reqField]; !exists {
			return fmt.Errorf("missing required parameter: %s", reqField)
		}
	}

	// Check parameter constraints & enum values
	for key, val := range args {
		paramSchema, exists := ts.Parameters[key]
		if !exists {
			continue // Allow unspecified extra parameters for flexibility
		}

		strVal := fmt.Sprintf("%v", val)

		// Check enum constraints
		if len(paramSchema.Enum) > 0 {
			validEnum := false
			for _, allowed := range paramSchema.Enum {
				if strings.EqualFold(strVal, allowed) {
					validEnum = true
					break
				}
			}
			if !validEnum {
				return fmt.Errorf("invalid value '%s' for parameter '%s'. Allowed values: %v", strVal, key, paramSchema.Enum)
			}
		}
	}

	return nil
}
