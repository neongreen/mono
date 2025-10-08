package tui

import (
	"claude-trace/pkg/storage"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	modeList mode = iota
	modeView
	modeAnnotate
)

// Model represents the TUI state
type Model struct {
	traces      []*storage.Trace
	currentIdx  int
	viewport    viewport.Model
	textarea    textarea.Model
	mode        mode
	width       int
	height      int
	message     string
	messageTime time.Time
}

// NewModel creates a new TUI model
func NewModel(traces []*storage.Trace) Model {
	// Load existing annotations for all traces
	for _, trace := range traces {
		storage.LoadAnnotations(trace)
	}

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	ta := textarea.New()
	ta.Placeholder = "Enter your notes here..."
	ta.Focus()

	return Model{
		traces:   traces,
		viewport: vp,
		textarea: ta,
		mode:     modeList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		m.textarea.SetWidth(msg.Width - 4)
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeList:
			return m.updateList(msg)
		case modeView:
			return m.updateView(msg)
		case modeAnnotate:
			return m.updateAnnotate(msg)
		}
	}

	if m.mode == modeView {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.mode == modeAnnotate {
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.currentIdx > 0 {
			m.currentIdx--
		}
	case "down", "j":
		if m.currentIdx < len(m.traces)-1 {
			m.currentIdx++
		}
	case "enter", " ":
		m.mode = modeView
		m.updateViewport()
	case "g":
		m.addTag("good")
	case "s":
		m.addTag("suspicious")
	case "f":
		m.addTag("frustration")
	case "b":
		m.addTag("bug")
	case "w":
		m.addTag("win")
	case "n":
		m.mode = modeAnnotate
		m.textarea.SetValue(m.traces[m.currentIdx].FreeformNote)
		m.textarea.Focus()
	}
	return m, nil
}

func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.mode = modeList
		return m, nil
	case "g":
		m.addTag("good")
	case "s":
		m.addTag("suspicious")
	case "f":
		m.addTag("frustration")
	case "b":
		m.addTag("bug")
	case "w":
		m.addTag("win")
	case "n":
		m.mode = modeAnnotate
		m.textarea.SetValue(m.traces[m.currentIdx].FreeformNote)
		m.textarea.Focus()
	case "S":
		if err := storage.SaveAnnotations(m.traces[m.currentIdx]); err != nil {
			m.message = fmt.Sprintf("Error saving: %v", err)
		} else {
			m.message = "Annotations saved!"
		}
		m.messageTime = time.Now()
	}
	return m, nil
}

func (m Model) updateAnnotate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.traces[m.currentIdx].FreeformNote = m.textarea.Value()
		m.mode = modeView
		m.updateViewport()
		return m, nil
	case "ctrl+s":
		m.traces[m.currentIdx].FreeformNote = m.textarea.Value()
		if err := storage.SaveAnnotations(m.traces[m.currentIdx]); err != nil {
			m.message = fmt.Sprintf("Error saving: %v", err)
		} else {
			m.message = "Annotations saved!"
		}
		m.messageTime = time.Now()
		m.mode = modeView
		m.updateViewport()
		return m, nil
	}
	return m, nil
}

func (m *Model) addTag(tag string) {
	if m.currentIdx < len(m.traces) {
		trace := m.traces[m.currentIdx]
		trace.Tags[tag] = !trace.Tags[tag] // Toggle tag
		trace.Annotations = append(trace.Annotations, storage.Annotation{
			Timestamp: time.Now(),
			Tag:       tag,
		})
		m.message = fmt.Sprintf("Toggled tag: %s", tag)
		m.messageTime = time.Now()
		if m.mode == modeView {
			m.updateViewport()
		}
	}
}

func (m *Model) updateViewport() {
	if m.currentIdx >= len(m.traces) {
		return
	}

	trace := m.traces[m.currentIdx]
	
	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Trace: %s\n", trace.Name)))
	content.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Modified: %s\n", trace.ModTime.Format("2006-01-02 15:04:05"))))
	content.WriteString("\n")

	// Show tags
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

	// Show freeform note
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

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.mode {
	case modeList:
		return m.viewList()
	case modeView:
		return m.viewTrace()
	case modeAnnotate:
		return m.viewAnnotate()
	}
	return ""
}

func (m Model) viewList() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	s.WriteString(headerStyle.Render("Claude Code Trace Reviewer"))
	s.WriteString("\n\n")

	// Show traces
	for i, trace := range m.traces {
		prefix := "  "
		if i == m.currentIdx {
			prefix = "> "
		}

		line := fmt.Sprintf("%s%s", prefix, trace.Name)
		
		// Add tags indicator
		var activeTags []string
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

func (m Model) viewTrace() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	s.WriteString(headerStyle.Render("Viewing Trace"))
	s.WriteString("\n\n")
	
	s.WriteString(m.viewport.View())
	s.WriteString("\n\n")
	s.WriteString(m.helpView())

	if m.message != "" && time.Since(m.messageTime) < 3*time.Second {
		s.WriteString("\n")
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		s.WriteString(msgStyle.Render(m.message))
	}

	return s.String()
}

func (m Model) viewAnnotate() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	s.WriteString(headerStyle.Render("Add Notes"))
	s.WriteString("\n\n")
	
	s.WriteString(m.textarea.View())
	s.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Faint(true)
	s.WriteString(helpStyle.Render("ctrl+s: Save and return | esc: Cancel"))

	if m.message != "" && time.Since(m.messageTime) < 3*time.Second {
		s.WriteString("\n")
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		s.WriteString(msgStyle.Render(m.message))
	}

	return s.String()
}

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
