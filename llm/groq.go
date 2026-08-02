package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/heraji/jarvis/types"
	openai "github.com/sashabaranov/go-openai"
)

const groqBaseURL = "https://api.groq.com/openai/v1"

// Groq implements the LLM interface using Groq's OpenAI-compatible API.
type Groq struct {
	client *openai.Client
	model  string
}

// NewGroq creates a new Groq LLM provider.
func NewGroq(apiKey, model string) (*Groq, error) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = groqBaseURL
	client := openai.NewClientWithConfig(config)

	return &Groq{
		client: client,
		model:  model,
	}, nil
}

// Chat sends messages and tool definitions to Groq and returns the response.
func (g *Groq) Chat(messages []types.Message, tools []types.ToolDefinition) (*types.ChatResponse, error) {
	ctx := context.Background()

	// Convert messages to OpenAI format
	var oaiMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		oaiMsg := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		if msg.ToolCallID != "" {
			oaiMsg.ToolCallID = msg.ToolCallID
		} else if msg.Role == "tool" {
			oaiMsg.ToolCallID = "call_default"
		}
		if msg.Name != "" {
			oaiMsg.Name = msg.Name
		}
		// Include tool calls for assistant messages
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				oaiMsg.ToolCalls = append(oaiMsg.ToolCalls, openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				})
			}
		}
		oaiMessages = append(oaiMessages, oaiMsg)
	}

	// Build request
	req := openai.ChatCompletionRequest{
		Model:    g.model,
		Messages: oaiMessages,
	}

	// Convert tool definitions to OpenAI format
	if len(tools) > 0 {
		for _, t := range tools {
			paramsJSON, _ := json.Marshal(t.Parameters)
			var paramsDef json.RawMessage = paramsJSON

			req.Tools = append(req.Tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  paramsDef,
				},
			})
		}
	}

	// Send request with automatic 429 rate limit retry loop
	var resp openai.ChatCompletionResponse
	var err error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = g.client.CreateChatCompletion(ctx, req)
		if err == nil {
			break
		}

		errStr := err.Error()

		// If hit Rate Limit 429, wait and retry automatically
		if strings.Contains(errStr, "429") || strings.Contains(errStr, "Too Many Requests") || strings.Contains(errStr, "Rate limit") {
			if attempt < maxRetries {
				time.Sleep(time.Duration(3+attempt*2) * time.Second)
				continue
			}
		}

		// Fallback 1: Retry with high-capacity model (llama-3.1-8b-instant)
		if req.Model != "llama-3.1-8b-instant" {
			req.Model = "llama-3.1-8b-instant"
			resp, err = g.client.CreateChatCompletion(ctx, req)
			if err == nil {
				break
			}
		}

		// Fallback 2: If 400 Bad Request occurs (Groq function call formatting error), retry without tool definitions
		if len(req.Tools) > 0 {
			req.Tools = nil
			resp, err = g.client.CreateChatCompletion(ctx, req)
			if err == nil {
				break
			}
		}

		return nil, fmt.Errorf("groq API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return &types.ChatResponse{}, nil
	}

	choice := resp.Choices[0]
	result := &types.ChatResponse{
		Content: choice.Message.Content,
	}

	// Parse tool calls
	for i, tc := range choice.Message.ToolCalls {
		callID := tc.ID
		if callID == "" {
			callID = fmt.Sprintf("call_%d_%d", time.Now().UnixNano(), i)
		}
		result.ToolCalls = append(result.ToolCalls, types.ToolCall{
			ID:        callID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return result, nil
}
