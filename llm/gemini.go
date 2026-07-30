package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/heraji/jarvis/types"
	"google.golang.org/genai"
)

// Gemini implements the LLM interface using Google's Gemini API.
type Gemini struct {
	client *genai.Client
	model  string
}

// NewGemini creates a new Gemini LLM provider.
func NewGemini(apiKey, model string) (*Gemini, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Gemini{
		client: client,
		model:  model,
	}, nil
}

// Chat sends messages and tool definitions to Gemini and returns the response.
func (g *Gemini) Chat(messages []types.Message, tools []types.ToolDefinition) (*types.ChatResponse, error) {
	ctx := context.Background()

	// Build Gemini config
	config := &genai.GenerateContentConfig{}

	// Convert tool definitions to Gemini format
	if len(tools) > 0 {
		var funcDecls []*genai.FunctionDeclaration
		for _, t := range tools {
			fd := &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
			}
			if t.Parameters != nil {
				fd.Parameters = convertToGeminiSchema(t.Parameters)
			}
			funcDecls = append(funcDecls, fd)
		}
		config.Tools = []*genai.Tool{
			{FunctionDeclarations: funcDecls},
		}
	}

	// Set system instruction from the first system message
	var contents []*genai.Content
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			config.SystemInstruction = &genai.Content{
				Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
			}
		case "user":
			contents = append(contents, &genai.Content{
				Role:  "user",
				Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
			})
		case "assistant":
			parts := []*genai.Part{}
			if msg.Content != "" {
				parts = append(parts, genai.NewPartFromText(msg.Content))
			}
			contents = append(contents, &genai.Content{
				Role:  "model",
				Parts: parts,
			})
		case "tool":
			// Tool responses are sent as function responses
			var resultMap map[string]any
			if err := json.Unmarshal([]byte(msg.Content), &resultMap); err != nil {
				resultMap = map[string]any{"result": msg.Content}
			}
			contents = append(contents, &genai.Content{
				Role:  "user",
				Parts: []*genai.Part{genai.NewPartFromFunctionResponse(msg.Name, resultMap)},
			})
		}
	}

	// Send request to Gemini
	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	// Parse response
	result := &types.ChatResponse{}

	if resp == nil || len(resp.Candidates) == 0 {
		return result, nil
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil {
		return result, nil
	}

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			result.Content += part.Text
		}
		if part.FunctionCall != nil {
			argsJSON, err := json.Marshal(part.FunctionCall.Args)
			if err != nil {
				argsJSON = []byte("{}")
			}
			result.ToolCalls = append(result.ToolCalls, types.ToolCall{
				ID:        fmt.Sprintf("call_%s_%d", part.FunctionCall.Name, len(result.ToolCalls)),
				Name:      part.FunctionCall.Name,
				Arguments: string(argsJSON),
			})
		}
	}

	return result, nil
}

// convertToGeminiSchema converts a JSON Schema map to Gemini's Schema type.
func convertToGeminiSchema(params map[string]interface{}) *genai.Schema {
	schema := &genai.Schema{
		Type: genai.TypeObject,
	}

	if props, ok := params["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]*genai.Schema)
		for name, propRaw := range props {
			if prop, ok := propRaw.(map[string]interface{}); ok {
				propSchema := &genai.Schema{}
				if t, ok := prop["type"].(string); ok {
					switch t {
					case "string":
						propSchema.Type = genai.TypeString
					case "number":
						propSchema.Type = genai.TypeNumber
					case "integer":
						propSchema.Type = genai.TypeInteger
					case "boolean":
						propSchema.Type = genai.TypeBoolean
					case "array":
						propSchema.Type = genai.TypeArray
					}
				}
				if desc, ok := prop["description"].(string); ok {
					propSchema.Description = desc
				}
				if enumVals, ok := prop["enum"].([]interface{}); ok {
					for _, v := range enumVals {
						if s, ok := v.(string); ok {
							propSchema.Enum = append(propSchema.Enum, s)
						}
					}
				}
				schema.Properties[name] = propSchema
			}
		}
	}

	if req, ok := params["required"].([]interface{}); ok {
		for _, r := range req {
			if s, ok := r.(string); ok {
				schema.Required = append(schema.Required, s)
			}
		}
	}

	return schema
}
