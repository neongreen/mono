package tui

import (
	"fmt"
	"github.com/neongreen/mono/claude-trace/pkg/storage"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// updateList handles key events in list mode
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

// updateView handles key events in view mode
func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q",
		"esc":
		m.mode = modeList
		return m, nil
	case
		"g":
		m.addTag(
			"good")
	case "s":
		m.addTag("suspicious")
	case "f":
		m.addTag("frustration")
	case "b":
		m.addTag("bug")
	case "w":
		m.addTag(
			"win")
	case "n":
		m.mode = modeAnnotate
		m.textarea.
			SetValue(m.
				traces[m.currentIdx].FreeformNote,
			)
		m.textarea.Focus()
	case "S":
		if err := storage.
			SaveAnnotations(m.traces[m.currentIdx]); err != nil {
			m.message = fmt.Sprintf("Error saving: %v",
				err)
		} else {
			m.message = "Annotations saved!"
		}
		m.messageTime = time.Now()
	}
	return m, nil
}

// updateAnnotate handles key events in annotation mode
func (m Model) updateAnnotate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.traces[m.currentIdx].FreeformNote = m.textarea.
			Value()
		m.mode = modeView
		m.updateViewport()
		return m, nil
	case "ctrl+s":
		m.traces[m.currentIdx].FreeformNote = m.textarea.Value()
		if err := storage.
			SaveAnnotations(m.
				traces[m.currentIdx]); err !=
			nil {
			m.
				message = fmt.Sprintf("Error saving: %v",

				err)
		} else {
			m.
				message =
				"Annotations saved!"
		}
		m.
			messageTime = time.
			Now()
		m.
			mode = modeView
		m.updateViewport()
		return m, nil
	}
	return m, nil
}
