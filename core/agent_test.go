package core

import (
	"testing"
)

func TestParseTextToolCalls(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedArgs string
	}{
		{
			name:         "Opening Tag Attribute Format (User Error Case)",
			input:        `<filesystem {"action": "list_dir", "path": "E:\\Project Freelance\\be-gorillacoach"}></filesystem>`,
			expectedName: "filesystem",
			expectedArgs: `{"action": "list_dir", "path": "E:\\Project Freelance\\be-gorillacoach"}`,
		},
		{
			name:         "Malformed Double Braces & Missing Closing Tag",
			input:        `<filesystem {"action": "list_dir", "path": "E:\\Project Freelance\\be-gorillacoach"}}`,
			expectedName: "filesystem",
			expectedArgs: `{"action": "list_dir", "path": "E:\\Project Freelance\\be-gorillacoach"}`,
		},
		{
			name:         "Vision Self-Closing Tag",
			input:        `<vision {"action": "inspect_active_window"}/>`,
			expectedName: "vision",
			expectedArgs: `{"action": "inspect_active_window"}`,
		},
		{
			name:         "Desktop Apps Opening Tag Attribute Format",
			input:        `<desktop_apps {"action": "launch", "app_name": "explorer", "arguments": "E:\\Project Freelance\\be-gorillacoach"}></desktop_apps>`,
			expectedName: "desktop_apps",
			expectedArgs: `{"action": "launch", "app_name": "explorer", "arguments": "E:\\Project Freelance\\be-gorillacoach"}`,
		},
		{
			name:         "XML Tag Wrapper Format",
			input:        `<desktop_apps>{"action": "launch"}</desktop_apps>`,
			expectedName: "desktop_apps",
			expectedArgs: `{"action": "launch"}`,
		},
		{
			name:         "Function Tag Format",
			input:        `<function=filesystem>{"action": "read_file", "path": "test.txt"}</function>`,
			expectedName: "filesystem",
			expectedArgs: `{"action": "read_file", "path": "test.txt"}`,
		},
		{
			name:         "Function Call Syntax",
			input:        `filesystem({"action": "create_dir", "path": "C:\\test"})`,
			expectedName: "filesystem",
			expectedArgs: `{"action": "create_dir", "path": "C:\\test"}`,
		},
		{
			name:         "XML Attribute Format with Double Brackets (Filesystem)",
			input:        `<filesystem action="list_dir" path="E:\Project Freelance\be-gorillacoach">></filesystem>`,
			expectedName: "filesystem",
			expectedArgs: `{"action":"list_dir","path":"E:\\Project Freelance\\be-gorillacoach"}`,
		},
		{
			name:         "XML Attribute Format with Double Brackets (Desktop Apps)",
			input:        `<desktop_apps action="launch" app_name="explorer" arguments="E:\Project Freelance\be-gorillacoach">></desktop_apps>`,
			expectedName: "desktop_apps",
			expectedArgs: `{"action":"launch","app_name":"explorer","arguments":"E:\\Project Freelance\\be-gorillacoach"}`,
		},
		{
			name:         "Trailing Semicolon Inside Tag Body (User Case)",
			input:        `<desktop_apps>{"action":"launch","app_name":"chrome","arguments":"https://www.youtube.com/results?search_query=Oasis+Definitely+Maybe"};</desktop_apps>`,
			expectedName: "desktop_apps",
			expectedArgs: `{"action":"launch","app_name":"chrome","arguments":"https://www.youtube.com/results?search_query=Oasis+Definitely+Maybe"}`,
		},
		{
			name:         "Positional String Function Call (User Case 1)",
			input:        `desktop_apps('launch', 'app_name', 'spotify')`,
			expectedName: "desktop_apps",
			expectedArgs: `{"action":"launch","app_name":"spotify"}`,
		},
		{
			name:         "Positional String Function Call with URI (User Case 2)",
			input:        `desktop_apps('desktop_apps', 'app_name', 'spotify', 'spotify:search:Wonderwall')`,
			expectedName: "desktop_apps",
			expectedArgs: `{"action":"launch","app_name":"spotify","arguments":"spotify:search:Wonderwall"}`,
		},
		{
			name:         "JSON Block Format",
			input:        `{"tool": "web_search", "arguments": {"query": "golang"}}`,
			expectedName: "web_search",
			expectedArgs: `{"query": "golang"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := parseTextToolCalls(tt.input)
			if len(calls) == 0 {
				t.Fatalf("Expected tool call for %s, got 0", tt.name)
			}
			if calls[0].Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, calls[0].Name)
			}
			if calls[0].Arguments != tt.expectedArgs {
				t.Errorf("Expected args %s, got %s", tt.expectedArgs, calls[0].Arguments)
			}
		})
	}
}
