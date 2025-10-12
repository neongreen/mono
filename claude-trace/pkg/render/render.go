package render

import (
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
	Content      string            `json:"content"`
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
	data := &TraceData{
		Name:         trace.Name,
		Path:         trace.Path,
		ModTime:      trace.ModTime,
		Content:      trace.Content,
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

	// Content
	sb.WriteString("## Trace Content\n\n")
	sb.WriteString("```\n")
	sb.WriteString(data.Content)
	sb.WriteString("\n```\n\n")

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
