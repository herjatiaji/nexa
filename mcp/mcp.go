package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// MCPRequest defines the standard JSON-RPC 2.0 request frame.
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse defines the standard JSON-RPC 2.0 response frame.
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError holds error details from an MCP server.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPToolDefinition represents a tool declared by an MCP server.
type MCPToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Client manages a connection to an external MCP server via stdio.
type Client struct {
	command string
	args    []string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  *bufio.Reader
	seqID   int64
	mu      sync.Mutex
}

// NewClient initializes a new MCP stdio client.
func NewClient(command string, args ...string) *Client {
	return &Client{
		command: command,
		args:    args,
	}
}

// Start launches the MCP server process and performs the JSON-RPC handshake.
func (c *Client) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cmd = exec.Command(c.command, c.args...)
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdoutPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	c.stdout = bufio.NewReader(stdoutPipe)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server process %s: %w", c.command, err)
	}

	// 1. Send initialize request
	initParams := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "NEXA-MCP-Client",
			"version": "1.0.0",
		},
	}

	_, err = c.callLocked("initialize", initParams)
	if err != nil {
		return fmt.Errorf("MCP initialize handshake failed: %w", err)
	}

	return nil
}

// ListTools queries the MCP server for available tools via tools/list.
func (c *Client) ListTools() ([]MCPToolDefinition, error) {
	respData, err := c.Call("tools/list", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var listResult struct {
		Tools []MCPToolDefinition `json:"tools"`
	}
	if err := json.Unmarshal(respData, &listResult); err != nil {
		return nil, fmt.Errorf("failed to parse tools/list result: %w", err)
	}

	return listResult.Tools, nil
}

// CallTool invokes a tool on the MCP server via tools/call.
func (c *Client) CallTool(name string, arguments map[string]interface{}) (string, error) {
	callParams := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}

	respData, err := c.Call("tools/call", callParams)
	if err != nil {
		return "", err
	}

	var callResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}

	if err := json.Unmarshal(respData, &callResult); err != nil {
		return string(respData), nil
	}

	if len(callResult.Content) > 0 {
		return callResult.Content[0].Text, nil
	}

	return string(respData), nil
}

// Call executes a thread-safe JSON-RPC 2.0 request.
func (c *Client) Call(method string, params interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.callLocked(method, params)
}

func (c *Client) callLocked(method string, params interface{}) (json.RawMessage, error) {
	id := atomic.AddInt64(&c.seqID, 1)

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	reqBytes = append(reqBytes, '\n')
	if _, err := c.stdin.Write(reqBytes); err != nil {
		return nil, fmt.Errorf("failed to write to MCP stdin: %w", err)
	}

	// Read line response from stdout
	line, err := c.stdout.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read from MCP stdout: %w", err)
	}

	var resp MCPResponse
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return nil, fmt.Errorf("invalid MCP response line: %w (line: %s)", err, line)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error [%d]: %s", resp.Error.Code, resp.Error.Message)
	}

	return resp.Result, nil
}

// Close terminates the MCP process gracefully.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}
