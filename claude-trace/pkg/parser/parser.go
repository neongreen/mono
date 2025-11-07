package parser

import (
	"encoding/json"
	"strings"
	"time"
)

// ParsedTrace represents the internal representation of a trace
type ParsedTrace struct {
	SessionID string             `json:"session_id"`
	Summary   string             `json:"summary"`
	StartTime time.Time          `json:"start_time"`
	Items     []ConversationItem `json:"items"`
}

// ConversationItem represents a single item in the conversation
type ConversationItem struct {
	Type      ItemType  `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	UUID      string    `json:"uuid,omitempty"`

	// User message fields
	UserMessage *UserMessage `json:"user_message,omitempty"`

	// Assistant message fields
	AssistantMessage *AssistantMessage `json:"assistant_message,omitempty"`

	// Tool result fields
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

type ItemType string

const (
	ItemTypeUser       ItemType = "user"
	ItemTypeAssistant  ItemType = "assistant"
	ItemTypeToolResult ItemType = "tool_result"
	ItemTypeSummary    ItemType = "summary"
	ItemTypeSnapshot   ItemType = "snapshot"
)

type UserMessage struct {
	Content string `json:"content"`
	CWD     string `json:"cwd,omitempty"`
}

type AssistantMessage struct {
	Model   string         `json:"model,omitempty"`
	Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
	Type     ContentType `json:"type"`
	Text     string      `json:"text,omitempty"`
	Thinking string      `json:"thinking,omitempty"`
	ToolUse  *ToolUse    `json:"tool_use,omitempty"`
}

type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeThinking ContentType = "thinking"
	ContentTypeToolUse  ContentType = "tool_use"
)

type ToolUse struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input,omitempty"`
}

type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// ParseTrace parses raw trace content into internal representation
func ParseTrace(content string) (*ParsedTrace, error) {
	// Detect format
	if isJSONL(content) {
		return parseJSONLTrace(content)
	}

	// For non-JSONL traces (like .log files), return simple representation
	return parseSimpleTrace(content)
}

func parseJSONLTrace(content string) (*ParsedTrace, error) {
	lines := strings.Split(content, "\n")
	trace := &ParsedTrace{
		Items: []ConversationItem{},
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue // Skip malformed lines
		}

		item := parseJSONLItem(raw)
		if item != nil {
			// Extract session ID and summary
			if trace.SessionID == "" {
				if sid, ok := raw["sessionId"].(string); ok {
					trace.SessionID = sid
				}
			}

			if item.Type == ItemTypeSummary && trace.Summary == "" {
				if summary, ok := raw["summary"].(string); ok {
					trace.Summary = summary
				}
			}

			trace.Items = append(trace.Items, *item)
		}
	}

	return trace, nil
}
