package render

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatToolArguments formats tool arguments in a human-readable way
func formatToolArguments(toolName string, input map[string]any) string {
	var sb strings.Builder
	switch strings.ToLower(toolName) {
	case "write":
		sb.WriteString(formatWriteToolArguments(input))
	default:
		inputJSON, _ := json.MarshalIndent(input, "", "  ")
		sb.WriteString("```json\n")
		sb.WriteString(string(inputJSON))
		sb.WriteString("\n```\n\n")
	}
	return sb.String()
}

// formatWriteToolArguments formats the "write" tool arguments nicely
func formatWriteToolArguments(input map[string]any) string {
	var sb strings.Builder

	filePath, hasFilePath := input["file_path"].(string)
	content, hasContent := input["content"].(string)
	if hasFilePath {
		sb.WriteString(
			fmt.Sprintf("**File:** `%s`\n\n",
				filePath))
	}
	if hasContent {
		sb.WriteString("**Content:**\n")
		sb.WriteString("```\n")
		sb.WriteString(content)
		sb.WriteString("\n```\n\n")
	}
	var unexpectedFields []string
	for key := range input {
		if key != "file_path" && key != "content" {
			unexpectedFields = append(unexpectedFields, key)
		}
	}
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
