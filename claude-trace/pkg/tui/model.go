package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neongreen/mono/claude-trace/pkg/storage"
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

	switch m.mode {
	case modeView:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	case modeAnnotate:
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
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
