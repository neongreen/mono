package parser

import (
	"testing"
)

func TestIsJSONL(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "valid JSONL",
			content:  `{"type":"user","message":"hello"}` + "\n" + `{"type":"assistant","message":"hi"}`,
			expected: true,
		},
		{
			name:     "not JSONL",
			content:  "just plain text\nno JSON here",
			expected: false,
		},
		{
			name:     "single line",
			content:  `{"type":"user","message":"hello"}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJSONL(tt.content)
			if result != tt.expected {
				t.Errorf("isJSONL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseUserMessage(t *testing.T) {
	tests := []struct {
		name     string
		raw      map[string]interface{}
		expected *UserMessage
	}{
		{
			name: "simple user message",
			raw: map[string]interface{}{
				"cwd": "/home/user",
				"message": map[string]interface{}{
					"content": "hello world",
				},
			},
			expected: &UserMessage{
				Content: "hello world",
				CWD:     "/home/user",
			},
		},
		{
			name: "user message with array content",
			raw: map[string]interface{}{
				"cwd": "/home/user",
				"message": map[string]interface{}{
					"content": []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": "hello world",
						},
					},
				},
			},
			expected: &UserMessage{
				Content: "hello world", // Should now be handled correctly
				CWD:     "/home/user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseUserMessage(tt.raw)
			if result.CWD != tt.expected.CWD {
				t.Errorf("parseUserMessage().CWD = %v, want %v", result.CWD, tt.expected.CWD)
			}
			if result.Content != tt.expected.Content {
				t.Errorf("parseUserMessage().Content = %v, want %v", result.Content, tt.expected.Content)
			}
		})
	}
}

func TestParseJSONLItem_ToolResult(t *testing.T) {
	// This is the key test - tool results have type:"user" but contain tool_result in content
	raw := map[string]interface{}{
		"type":      "user",
		"cwd":       "/home/user",
		"uuid":      "test-uuid",
		"timestamp": "2025-10-05T10:18:44.315Z",
		"message": map[string]interface{}{
			"role": "user",
			"content": []interface{}{
				map[string]interface{}{
					"tool_use_id": "toolu_123",
					"type":        "tool_result",
					"content":     "No files found",
				},
			},
		},
	}

	item := parseJSONLItem(raw)

	if item == nil {
		t.Fatal("parseJSONLItem() returned nil")
	}

	if item.Type != ItemTypeToolResult {
		t.Errorf("parseJSONLItem().Type = %v, want %v", item.Type, ItemTypeToolResult)
	}

	if item.ToolResult == nil {
		t.Fatal("parseJSONLItem().ToolResult is nil")
	}

	if item.ToolResult.ToolUseID != "toolu_123" {
		t.Errorf("ToolResult.ToolUseID = %v, want %v", item.ToolResult.ToolUseID, "toolu_123")
	}

	if item.ToolResult.Content != "No files found" {
		t.Errorf("ToolResult.Content = %v, want %v", item.ToolResult.Content, "No files found")
	}
}

func TestParseJSONLItem_UserMessage(t *testing.T) {
	raw := map[string]interface{}{
		"type":      "user",
		"cwd":       "/home/user",
		"uuid":      "test-uuid",
		"timestamp": "2025-10-05T10:18:34.589Z",
		"message": map[string]interface{}{
			"role":    "user",
			"content": "Hello, how are you?",
		},
	}

	item := parseJSONLItem(raw)

	if item == nil {
		t.Fatal("parseJSONLItem() returned nil")
	}

	if item.Type != ItemTypeUser {
		t.Errorf("parseJSONLItem().Type = %v, want %v", item.Type, ItemTypeUser)
	}

	if item.UserMessage == nil {
		t.Fatal("parseJSONLItem().UserMessage is nil")
	}
}

func TestParseAssistantMessage(t *testing.T) {
	raw := map[string]interface{}{
		"message": map[string]interface{}{
			"model": "claude-sonnet-4",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Hello!",
				},
				map[string]interface{}{
					"type":     "thinking",
					"thinking": "Let me think...",
				},
				map[string]interface{}{
					"type": "tool_use",
					"id":   "tool-123",
					"name": "read_file",
					"input": map[string]interface{}{
						"path": "/test.txt",
					},
				},
			},
		},
	}

	result := parseAssistantMessage(raw)

	if result.Model != "claude-sonnet-4" {
		t.Errorf("Model = %v, want %v", result.Model, "claude-sonnet-4")
	}

	if len(result.Content) != 3 {
		t.Fatalf("len(Content) = %v, want 3", len(result.Content))
	}

	if result.Content[0].Type != ContentTypeText {
		t.Errorf("Content[0].Type = %v, want %v", result.Content[0].Type, ContentTypeText)
	}

	if result.Content[1].Type != ContentTypeThinking {
		t.Errorf("Content[1].Type = %v, want %v", result.Content[1].Type, ContentTypeThinking)
	}

	if result.Content[2].Type != ContentTypeToolUse {
		t.Errorf("Content[2].Type = %v, want %v", result.Content[2].Type, ContentTypeToolUse)
	}
}

func TestParseTrace_Integration(t *testing.T) {
	jsonlContent := `{"type":"summary","summary":"Test Session","sessionId":"session-123"}
{"type":"user","cwd":"/home","uuid":"msg-1","timestamp":"2025-10-05T10:00:00Z","message":{"role":"user","content":"Hello"}}
{"type":"assistant","uuid":"msg-2","timestamp":"2025-10-05T10:00:01Z","message":{"role":"assistant","model":"claude-sonnet-4","content":[{"type":"text","text":"Hi there"}]}}
{"type":"user","uuid":"msg-3","timestamp":"2025-10-05T10:00:02Z","message":{"role":"user","content":[{"tool_use_id":"tool-1","type":"tool_result","content":"Result data"}]}}
`

	trace, err := ParseTrace(jsonlContent)
	if err != nil {
		t.Fatalf("ParseTrace() error = %v", err)
	}

	if trace.SessionID != "session-123" {
		t.Errorf("SessionID = %v, want %v", trace.SessionID, "session-123")
	}

	if trace.Summary != "Test Session" {
		t.Errorf("Summary = %v, want %v", trace.Summary, "Test Session")
	}

	// Should have 4 items: summary, user message, assistant message, tool result
	if len(trace.Items) != 4 {
		t.Errorf("len(Items) = %v, want 4", len(trace.Items))
	}

	// Check that tool result is properly identified
	toolResultFound := false
	for _, item := range trace.Items {
		if item.Type == ItemTypeToolResult {
			toolResultFound = true
			if item.ToolResult.Content != "Result data" {
				t.Errorf("ToolResult.Content = %v, want 'Result data'", item.ToolResult.Content)
			}
		}
	}

	if !toolResultFound {
		t.Error("No tool result found in parsed items")
	}
}
