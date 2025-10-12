package parser

import (
	"encoding/json"
	"strings"
	"time"
)

// ParsedTrace represents the internal representation of a trace
type ParsedTrace struct {
	SessionID string
	Summary   string
	StartTime time.Time
	Items     []ConversationItem
}

// ConversationItem represents a single item in the conversation
type ConversationItem struct {
	Type      ItemType
	Timestamp time.Time
	UUID      string

	// User message fields
	UserMessage *UserMessage

	// Assistant message fields
	AssistantMessage *AssistantMessage

	// Tool result fields
	ToolResult *ToolResult
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
	Content string
	CWD     string
}

type AssistantMessage struct {
	Model   string
	Content []ContentBlock
}

type ContentBlock struct {
	Type     ContentType
	Text     string
	Thinking string
	ToolUse  *ToolUse
}

type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeThinking ContentType = "thinking"
	ContentTypeToolUse  ContentType = "tool_use"
)

type ToolUse struct {
	ID    string
	Name  string
	Input map[string]interface{}
}

type ToolResult struct {
	ToolUseID string
	Content   string
	IsError   bool
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

		var raw map[string]interface{}
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
