package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/heraji/jarvis/types"
)

// Ollama implements the LLM interface using Ollama's REST API.
type Ollama struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllama creates a new Ollama LLM provider.
func NewOllama(baseURL, model string) (*Ollama, error) {
	return &Ollama{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}, nil
}

// ollamaChatRequest represents the Ollama /api/chat request body.
type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []ollamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
	Tools    []ollamaTool        `json:"tools,omitempty"`
}

type ollamaChatMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaTool struct {
	Type     string             `json:"type"`
	Function ollamaToolFunction `json:"function"`
}

type ollamaToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ollamaToolCall struct {
	Function ollamaToolCallFunction `json:"function"`
}

type ollamaToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ollamaChatResponse represents the Ollama /api/chat response body.
type ollamaChatResponse struct {
	Message ollamaChatMessage `json:"message"`
	Done    bool              `json:"done"`
}

// Chat sends messages and tool definitions to Ollama and returns the response.
func (o *Ollama) Chat(messages []types.Message, tools []types.ToolDefinition) (*types.ChatResponse, error) {
	// Convert messages to Ollama format
	var ollamaMessages []ollamaChatMessage
	for _, msg := range messages {
		ollamaMessages = append(ollamaMessages, ollamaChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Convert tool definitions to Ollama format
	var ollamaTools []ollamaTool
	for _, t := range tools {
		ollamaTools = append(ollamaTools, ollamaTool{
			Type: "function",
			Function: ollamaToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	// Build request
	reqBody := ollamaChatRequest{
		Model:    o.model,
		Messages: ollamaMessages,
		Stream:   false,
		Tools:    ollamaTools,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	// Send HTTP request
	url := fmt.Sprintf("%s/api/chat", o.baseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("ollama API error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response
	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	result := &types.ChatResponse{
		Content: ollamaResp.Message.Content,
	}

	// Parse tool calls from response
	for i, tc := range ollamaResp.Message.ToolCalls {
		argsJSON, err := json.Marshal(tc.Function.Arguments)
		if err != nil {
			argsJSON = []byte("{}")
		}
		result.ToolCalls = append(result.ToolCalls, types.ToolCall{
			ID:        fmt.Sprintf("ollama_call_%d", i),
			Name:      tc.Function.Name,
			Arguments: string(argsJSON),
		})
	}

	return result, nil
}
