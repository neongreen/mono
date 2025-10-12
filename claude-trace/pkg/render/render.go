package render

import (
	"claude-trace/pkg/parser"
	"claude-trace/pkg/storage"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TraceData represents the intermediate representation for rendering traces
type TraceData struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	ModTime      time.Time         `json:"mod_time"`
	ParsedTrace  *parser.ParsedTrace `json:"parsed_trace"`
	Tags         map[string]bool   `json:"tags"`
	FreeformNote string            `json:"freeform_note,omitempty"`
	Annotations  []AnnotationData  `json:"annotations,omitempty"`
}

// AnnotationData represents an annotation in the intermediate format
type AnnotationData struct {
	Timestamp time.Time `json:"timestamp"`
	Tag       string    `json:"tag,omitempty"`
	Note      string    `json:"note,omitempty"`
}

// ToTraceData converts a storage.Trace to the intermediate TraceData format
func ToTraceData(trace *storage.Trace) *TraceData {
	// Parse the trace content
	parsedTrace, err := parser.ParseTrace(trace.Content)
	if err != nil {
		// If parsing fails, create a minimal parsed trace
		parsedTrace = &parser.ParsedTrace{
			Summary: "Failed to parse trace",
			Items:   []parser.ConversationItem{},
		}
	}

	data := &TraceData{
		Name:         trace.Name,
		Path:         trace.Path,
		ModTime:      trace.ModTime,
		ParsedTrace:  parsedTrace,
		Tags:         trace.Tags,
		FreeformNote: trace.FreeformNote,
		Annotations:  make([]AnnotationData, len(trace.Annotations)),
	}

	// Convert annotations
	for i, ann := range trace.Annotations {
		data.Annotations[i] = AnnotationData{
			Timestamp: ann.Timestamp,
			Tag:       ann.Tag,
			Note:      ann.Note,
		}
	}

	return data
}

// RenderToJSON renders TraceData to JSON format
func RenderToJSON(data *TraceData) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// RenderToMarkdown renders TraceData to Markdown format
func RenderToMarkdown(data *TraceData) ([]byte, error) {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", data.Name))

	// Metadata
	sb.WriteString("## Metadata\n\n")
	sb.WriteString(fmt.Sprintf("- **Path:** `%s`\n", data.Path))
	sb.WriteString(fmt.Sprintf("- **Modified:** %s\n", data.ModTime.Format(time.RFC3339)))
	if data.ParsedTrace.SessionID != "" {
		sb.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", data.ParsedTrace.SessionID))
	}
	if data.ParsedTrace.Summary != "" {
		sb.WriteString(fmt.Sprintf("- **Summary:** %s\n", data.ParsedTrace.Summary))
	}

	// Tags
	if len(data.Tags) > 0 {
		sb.WriteString("- **Tags:** ")
		var tags []string
		for tag, active := range data.Tags {
			if active {
				tags = append(tags, fmt.Sprintf("`%s`", tag))
			}
		}
		sb.WriteString(strings.Join(tags, ", "))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Freeform Note
	if data.FreeformNote != "" {
		sb.WriteString("## Notes\n\n")
		sb.WriteString(data.FreeformNote)
		sb.WriteString("\n\n")
	}

	// Conversation
	sb.WriteString("## Conversation\n\n")
	renderConversationItems(&sb, data.ParsedTrace.Items)

	// Annotations History
	if len(data.Annotations) > 0 {
		sb.WriteString("## Annotation History\n\n")
		for _, ann := range data.Annotations {
			sb.WriteString(fmt.Sprintf("- **%s**", ann.Timestamp.Format(time.RFC3339)))
			if ann.Tag != "" {
				sb.WriteString(fmt.Sprintf(" - Tag: `%s`", ann.Tag))
			}
			if ann.Note != "" {
				sb.WriteString(fmt.Sprintf(" - Note: %s", ann.Note))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return []byte(sb.String()), nil
}

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
			// Skip summary and file history snapshots
			continue

		default:
			// Skip unknown types
			continue
		}
	}
}

// formatToolArguments formats tool arguments in a human-readable way
func formatToolArguments(toolName string, input map[string]interface{}) string {
	var sb strings.Builder
	
	// Special formatting for specific tools (case-insensitive)
	switch strings.ToLower(toolName) {
	case "write":
		sb.WriteString(formatWriteToolArguments(input))
	default:
		// Default JSON formatting for other tools
		inputJSON, _ := json.MarshalIndent(input, "", "  ")
		sb.WriteString("```json\n")
		sb.WriteString(string(inputJSON))
		sb.WriteString("\n```\n\n")
	}
	
	return sb.String()
}

// formatWriteToolArguments formats the "write" tool arguments nicely
func formatWriteToolArguments(input map[string]interface{}) string {
	var sb strings.Builder
	
	// Extract expected fields
	filePath, hasFilePath := input["file_path"].(string)
	content, hasContent := input["content"].(string)
	
	// Show file path
	if hasFilePath {
		sb.WriteString(fmt.Sprintf("**File:** `%s`\n\n", filePath))
	}
	
	// Show content in a code block
	if hasContent {
		sb.WriteString("**Content:**\n")
		sb.WriteString("```\n")
		sb.WriteString(content)
		sb.WriteString("\n```\n\n")
	}
	
	// Check for unexpected fields
	var unexpectedFields []string
	for key := range input {
		if key != "file_path" && key != "content" {
			unexpectedFields = append(unexpectedFields, key)
		}
	}
	
	// Show unexpected fields with warning
	if len(unexpectedFields) > 0 {
		sb.WriteString("⚠️ **Unexpected fields:** ")
		for i, field := range unexpectedFields {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("`%s`", field))
		}
		sb.WriteString("\n\n")
	}
	
	return sb.String()
}
