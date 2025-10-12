package tui

import (
	"claude-trace/pkg/storage"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// addTag toggles a tag on the current trace and adds an annotation
func (m *Model) addTag(tag string) {
	if m.currentIdx < len(m.traces) {
		trace := m.traces[m.currentIdx]
		trace.Tags[tag] = !trace.Tags[tag]
		trace.Annotations = append(trace.Annotations, storage.Annotation{Timestamp: time.Now(), Tag: tag})
		m.message = fmt.Sprintf("Toggled tag: %s", tag)
		m.messageTime = time.Now()
		if m.mode == modeView {
			m.updateViewport()
		}
	}
}

// updateViewport refreshes the viewport content with the current trace
func (m *Model) updateViewport() {
	if m.currentIdx >= len(m.
		traces) {
		return
	}
	trace :=
		m.traces[m.currentIdx]
	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Trace: %s\n",
		trace.Name)))
	content.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.
		Sprintf("Modified: %s\n",
			trace.ModTime.Format("2006-01-02 15:04:05"))))
	content.WriteString("\n")
	if len(trace.Tags) > 0 {
		content.WriteString(lipgloss.NewStyle().Bold(true).Render("Tags: "))
		var activeTags []string
		for tag, active := range trace.Tags {
			if active {
				activeTags = append(activeTags, tag)
			}
		}
		content.WriteString(strings.Join(activeTags, ", "))
		content.WriteString("\n\n")
	}
	if trace.FreeformNote != "" {
		content.WriteString(lipgloss.NewStyle().Bold(true).Render("Notes:\n"))
		content.WriteString(trace.FreeformNote)
		content.WriteString("\n\n")
	}
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Content:\n"))
	content.WriteString(strings.Repeat("─", 80))
	content.WriteString("\n")
	content.WriteString(trace.Content)
	m.viewport.SetContent(content.String())
}
