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
	ItemTypeUser      ItemType = "user"
	ItemTypeAssistant ItemType = "assistant"
	ItemTypeToolResult ItemType = "tool_result"
	ItemTypeSummary   ItemType = "summary"
	ItemTypeSnapshot  ItemType = "snapshot"
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

func isJSONL(content string) bool {
	lines := strings.Split(content, "\n")
	jsonLineCount := 0

	for i, line := range lines {
		if i > 5 {
			break // Check first few lines
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			jsonLineCount++
		}
	}

	return jsonLineCount >= 2
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

func parseJSONLItem(raw map[string]interface{}) *ConversationItem {
	itemType, _ := raw["type"].(string)

	item := &ConversationItem{
		UUID: getString(raw, "uuid"),
	}

	// Parse timestamp
	if ts, ok := raw["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			item.Timestamp = t
		}
	}

	// Check if this is actually a tool result (even if type is "user")
	if message, ok := raw["message"].(map[string]interface{}); ok {
		if content, ok := message["content"].([]interface{}); ok {
			if len(content) > 0 {
				if first, ok := content[0].(map[string]interface{}); ok {
					if _, hasToolUseID := first["tool_use_id"]; hasToolUseID {
						item.Type = ItemTypeToolResult
						item.ToolResult = parseToolResult(first)
						return item
					}
				}
			}
		}
	}

	switch itemType {
	case "user":
		item.Type = ItemTypeUser
		item.UserMessage = parseUserMessage(raw)

	case "assistant":
		item.Type = ItemTypeAssistant
		item.AssistantMessage = parseAssistantMessage(raw)

	case "summary":
		item.Type = ItemTypeSummary
		// Summary is extracted at trace level

	case "file-history-snapshot":
		item.Type = ItemTypeSnapshot
		// Skip or handle minimally

	default:
		return nil
	}

	return item
}

func parseUserMessage(raw map[string]interface{}) *UserMessage {
	msg := &UserMessage{
		CWD: getString(raw, "cwd"),
	}

	if message, ok := raw["message"].(map[string]interface{}); ok {
		// Handle string content
		if contentStr, ok := message["content"].(string); ok {
			msg.Content = contentStr
		} else if contentArr, ok := message["content"].([]interface{}); ok {
			// Handle array content - extract text from text blocks
			var texts []string
			for _, c := range contentArr {
				if block, ok := c.(map[string]interface{}); ok {
					if blockType := getString(block, "type"); blockType == "text" {
						texts = append(texts, getString(block, "text"))
					}
				}
			}
			msg.Content = strings.Join(texts, "\n")
		}
	}

	return msg
}

func parseAssistantMessage(raw map[string]interface{}) *AssistantMessage {
	msg := &AssistantMessage{
		Content: []ContentBlock{},
	}

	if message, ok := raw["message"].(map[string]interface{}); ok {
		msg.Model = getString(message, "model")

		if content, ok := message["content"].([]interface{}); ok {
			for _, c := range content {
				if block, ok := c.(map[string]interface{}); ok {
					msg.Content = append(msg.Content, parseContentBlock(block))
				}
			}
		}
	}

	return msg
}

func parseContentBlock(raw map[string]interface{}) ContentBlock {
	blockType := getString(raw, "type")

	block := ContentBlock{}

	switch blockType {
	case "text":
		block.Type = ContentTypeText
		block.Text = getString(raw, "text")

	case "thinking":
		block.Type = ContentTypeThinking
		block.Thinking = getString(raw, "thinking")

	case "tool_use":
		block.Type = ContentTypeToolUse
		block.ToolUse = &ToolUse{
			ID:   getString(raw, "id"),
			Name: getString(raw, "name"),
		}
		if input, ok := raw["input"].(map[string]interface{}); ok {
			block.ToolUse.Input = input
		}
	}

	return block
}

func parseToolResult(raw map[string]interface{}) *ToolResult {
	result := &ToolResult{
		ToolUseID: getString(raw, "tool_use_id"),
		Content:   getString(raw, "content"),
	}

	if isError, ok := raw["is_error"].(bool); ok {
		result.IsError = isError
	}

	return result
}

func parseSimpleTrace(content string) (*ParsedTrace, error) {
	// For non-JSONL traces, create a simple representation
	return &ParsedTrace{
		Summary: "Legacy trace format",
		Items:   []ConversationItem{},
	}, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
