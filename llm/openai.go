package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/heraji/jarvis/types"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAICompatible implements the LLM interface for OpenAI, OpenRouter, DeepSeek, and other OpenAI-compatible endpoints.
type OpenAICompatible struct {
	client *openai.Client
	model  string
	name   string
}

// NewOpenAI creates a provider for official OpenAI models (e.g. gpt-4o, gpt-4o-mini).
func NewOpenAI(apiKey, model string) (*OpenAICompatible, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required when using openai provider")
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	config := openai.DefaultConfig(apiKey)
	client := openai.NewClientWithConfig(config)
	return &OpenAICompatible{client: client, model: model, name: "OpenAI"}, nil
}

// NewOpenRouter creates a provider for OpenRouter API (e.g. deepseek/deepseek-chat, anthropic/claude-3.5-sonnet).
func NewOpenRouter(apiKey, model string) (*OpenAICompatible, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY environment variable is required when using openrouter provider")
	}
	if model == "" {
		model = "meta-llama/llama-3.3-70b-instruct"
	}
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://openrouter.ai/api/v1"
	client := openai.NewClientWithConfig(config)
	return &OpenAICompatible{client: client, model: model, name: "OpenRouter"}, nil
}

// NewDeepSeek creates a provider for official DeepSeek API (e.g. deepseek-chat, deepseek-reasoner).
func NewDeepSeek(apiKey, model string) (*OpenAICompatible, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is required when using deepseek provider")
	}
	if model == "" {
		model = "deepseek-chat"
	}
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com"
	client := openai.NewClientWithConfig(config)
	return &OpenAICompatible{client: client, model: model, name: "DeepSeek"}, nil
}

// Chat sends conversation history and tools to the OpenAI-compatible API.
func (o *OpenAICompatible) Chat(messages []types.Message, tools []types.ToolDefinition) (*types.ChatResponse, error) {
	ctx := context.Background()

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

	req := openai.ChatCompletionRequest{
		Model:    o.model,
		Messages: oaiMessages,
	}

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

	resp, err := o.client.CreateChatCompletion(ctx, req)
	if err != nil && len(req.Tools) > 0 {
		// Fallback: Retry without tools if function calling schema fails
		req.Tools = nil
		resp, err = o.client.CreateChatCompletion(ctx, req)
	}
	if err != nil {
		return nil, fmt.Errorf("%s API error: %w", o.name, err)
	}

	if len(resp.Choices) == 0 {
		return &types.ChatResponse{}, nil
	}

	choice := resp.Choices[0]
	result := &types.ChatResponse{
		Content: choice.Message.Content,
	}

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
