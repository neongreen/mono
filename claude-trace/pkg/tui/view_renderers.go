package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// viewList renders the trace list view
func (m Model) viewList() string {
	var s strings.Builder
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	s.WriteString(headerStyle.Render("Claude Code Trace Reviewer"))
	s.WriteString("\n\n")
	for i, trace := range m.traces {
		prefix := "  "
		if i == m.currentIdx {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%s", prefix, trace.Name)
		var activeTags []string // Add tags indicator

		for tag, active := range trace.Tags {
			if active {
				activeTags = append(activeTags, tag)
			}
		}
		if len(activeTags) > 0 {
			tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			line += tagStyle.Render(fmt.Sprintf(" [%s]", strings.Join(activeTags, ", ")))
		}
		s.WriteString(line)
		s.WriteString("\n")
	}
	s.WriteString("\n")
	s.WriteString(m.helpView())
	if m.message != "" && time.Since(m.messageTime) < 3*time.Second {
		s.WriteString("\n")
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		s.WriteString(msgStyle.Render(m.message))
	}
	return s.String()
}

// viewTrace renders the trace detail view
func (m Model) viewTrace() string {
	var s strings.Builder
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	s.WriteString(headerStyle.Render("Viewing Trace"))
	s.WriteString("\n\n")
	s.WriteString(m.viewport.
		View())
	s.WriteString("\n\n")
	s.WriteString(m.helpView())
	if m.message !=
		"" && time.
		Since(m.messageTime) < 3*time.Second {
		s.WriteString("\n")
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		s.WriteString(msgStyle.Render(m.message))
	}
	return s.String()
}

// viewAnnotate renders the annotation input view
func (m Model) viewAnnotate() string {
	var s strings.Builder
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	s.
		WriteString(headerStyle.Render("Add Notes"))
	s.WriteString("\n\n")
	s.WriteString(m.textarea.
		View())
	s.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Faint(true)
	s.WriteString(helpStyle.Render("ctrl+s: Save and return | esc: Cancel"))
	if m.message != "" && time.Since(m.
		messageTime,
	) < 3*time.Second {
		s.WriteString("\n")
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.
			Color("42"))
		s.WriteString(msgStyle.Render(m.message))
	}
	return s.String()
}

// helpView renders the help text for the current mode
func (m Model) helpView() string {
	helpStyle := lipgloss.NewStyle().Faint(true)
	switch m.mode {
	case modeList:
		return helpStyle.Render("↑/k: Up | ↓/j: Down | enter: View | g: Good | s: Suspicious | f: Frustration | b: Bug | w: Win | n: Add notes | q: Quit")
	case modeView:
		return helpStyle.Render("↑/↓: Scroll | g: Good | s: Suspicious | f: Frustration | b: Bug | w: Win | n: Add notes | S: Save | q: Back | esc: Back")
	case modeAnnotate:
		return helpStyle.Render("Type your notes... | ctrl+s: Save | esc: Cancel")
	}
	return ""
}
