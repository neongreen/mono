package parser

import (
	"strings"
	"time"
)

// parseJSONLItem parses a single JSONL item into a ConversationItem
func parseJSONLItem(raw map[string]any) *ConversationItem {
	itemType, _ := raw["type"].(string)
	item := &ConversationItem{UUID: getString(raw, "uuid")}
	if ts, ok := raw["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			item.Timestamp = t
		}
	}
	if message, ok := raw["message"].(map[string]any); ok {
		if content, ok := message["content"].([]any); ok {
			if len(content) > 0 {
				if first, ok := content[0].(map[string]any); ok {
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
	case "file-history-snapshot":
		item.Type = ItemTypeSnapshot
	default:
		return nil
	}
	return item
}

// parseUserMessage extracts user message data from raw JSONL item
func parseUserMessage(raw map[string]any) *UserMessage {
	msg := &UserMessage{CWD: getString(raw, "cwd")}
	if message, ok := raw["message"].(map[string]any); ok {
		if contentStr, ok := message["content"].(string); ok {
			msg.Content =
				contentStr
		} else if contentArr, ok := message["content"].([]any); ok { // Handle array content - extract text from text blocks
			var texts []string
			for _, c := range contentArr {
				if block, ok := c.(map[string]any); ok {
					if blockType := getString(block, "type"); blockType == "text" {
						texts = append(texts, getString(block,
							"text"))
					}
				}
			}
			msg.
				Content = strings.Join(texts,
				"\n")
		}
	}
	return msg
}

// parseAssistantMessage extracts assistant message data from raw JSONL item
func parseAssistantMessage(raw map[string]any) *AssistantMessage {
	msg := &AssistantMessage{Content: []ContentBlock{}}
	if message, ok := raw["message"].(map[string]any); ok {
		msg.Model = getString(message, "model")
		if content, ok := message["content"].([]any); ok {
			for _, c := range content {
				if block, ok := c.(map[string]any); ok {
					msg.Content = append(msg.Content, parseContentBlock(block))
				}
			}
		}
	}
	return msg
}

// parseContentBlock extracts content block data from raw JSONL item
func parseContentBlock(raw map[string]any) ContentBlock {
	blockType := getString(raw, "type")
	block := ContentBlock{}
	switch blockType {
	case "text":
		block.Type = ContentTypeText
		block.Text = getString(raw, "text")
	case
		"thinking":
		block.Type = ContentTypeThinking
		block.Thinking = getString(raw,
			"thinking",
		)
	case "tool_use":
		block.Type = ContentTypeToolUse
		block.
			ToolUse = &ToolUse{ID: getString(raw, "id"), Name: getString(raw, "name")}
		if input, ok := raw["input"].(map[string]any); ok {
			block.ToolUse.Input = input
		}
	}
	return block
}

// parseToolResult extracts tool result data from raw JSONL item
func parseToolResult(raw map[string]any) *ToolResult {
	result := &ToolResult{ToolUseID: getString(raw, "tool_use_id"), Content: getString(raw, "content")}
	if isError, ok :=
		raw["is_error"].(bool); ok {
		result.IsError = isError
	}
	return result
}

// getString safely extracts a string value from a map, returning empty string if not found
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
