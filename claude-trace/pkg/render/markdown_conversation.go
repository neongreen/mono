package render

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/claude-trace/pkg/parser"
)

// renderConversationItems renders conversation items in a human-readable format
func renderConversationItems(sb *strings.Builder, items []parser.ConversationItem) {
	messageNum := 0
	for _, item := range items {
		switch item.Type {
		case parser.ItemTypeUser:
			messageNum++
			sb.WriteString(fmt.Sprintf("### Message %d: User\n\n", messageNum))
			if !item.Timestamp.IsZero() {
				sb.WriteString(fmt.Sprintf("**Time:** %s\n\n", item.Timestamp.Format("2006-01-02 15:04:05")))
			}
			if item.UserMessage != nil {
				sb.WriteString(item.UserMessage.Content)
				if item.UserMessage.CWD != "" {
					sb.WriteString(fmt.Sprintf("\n\n*Working directory: %s*", item.UserMessage.CWD))
				}
			}
			sb.WriteString("\n\n---\n\n")
		case parser.ItemTypeAssistant:
			messageNum++
			sb.WriteString(fmt.Sprintf("### Message %d: Assistant", messageNum))
			if item.AssistantMessage != nil && item.AssistantMessage.Model != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", item.AssistantMessage.Model))
			}
			sb.WriteString("\n\n")
			if !item.Timestamp.IsZero() {
				sb.WriteString(fmt.Sprintf("**Time:** %s\n\n", item.Timestamp.Format("2006-01-02 15:04:05")))
			}
			if item.AssistantMessage != nil {
				for _, content := range item.AssistantMessage.Content {
					switch content.Type {
					case parser.ContentTypeThinking:
						sb.WriteString("<details>\n<summary>Thinking</summary>\n\n")
						sb.WriteString(content.Thinking)
						sb.WriteString("\n</details>\n\n")
					case parser.ContentTypeText:
						sb.WriteString(content.Text)
						sb.WriteString("\n\n")
					case parser.ContentTypeToolUse:
						sb.WriteString(fmt.Sprintf("**Tool Use:** `%s`\n\n", content.ToolUse.Name))
						if len(content.ToolUse.Input) > 0 {
							sb.WriteString(formatToolArguments(content.ToolUse.Name, content.ToolUse.Input))
						}
					}
				}
			}
			sb.WriteString("---\n\n")
		case parser.ItemTypeToolResult:
			if item.ToolResult != nil {
				sb.WriteString(fmt.Sprintf("**Tool Result:** %s\n\n", item.ToolResult.ToolUseID))
				if item.ToolResult.IsError {
					sb.WriteString("*Error:* ")
				}
				sb.WriteString("```\n")
				sb.WriteString(item.ToolResult.Content)
				sb.WriteString("\n```\n\n")
			}
			sb.WriteString("---\n\n")
		case parser.ItemTypeSummary, parser.ItemTypeSnapshot:
			continue
		default:
			continue
		}
	}
}
