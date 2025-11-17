package cmd

import (
	"encoding/json"
	"testing"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestCalculateColumnWidths_Basic(t *testing.T) {
	tasks := []*types.Task{
		{
			TaskDisplayID: "test-1",
			TaskUUID:      "uuid1",
			Title:         "Test task",
			Axes: map[string]types.AxisStatus{
				"generic": {Effective: "done"},
			},
			Metadata: map[string]types.MetadataStatus{
				"priority": {Effective: json.RawMessage(`1`)},
			},
		},
	}

	displayIDs := map[string]string{
		"uuid1": "test-1",
	}

	constraints := ColumnConstraints{
		TermWidth:      80,
		ShowAliases:    false,
		LabelsMaxWidth: 10,
		TitleMinWidth:  30,
		SeparatorWidth: 3,
		PaddingPerCell: 2,
	}

	widths := CalculateColumnWidths(tasks, displayIDs, constraints)

	// ID should be at least len("test-1") = 6
	if widths.ID < 6 {
		t.Errorf("ID width too small: got %d, want at least 6", widths.ID)
	}

	// Status should be at least len("done") = 4
	if widths.Status < 4 {
		t.Errorf("Status width too small: got %d, want at least 4", widths.Status)
	}

	// Priority should be at least 1
	if widths.Priority < 1 {
		t.Errorf("Priority width too small: got %d, want at least 1", widths.Priority)
	}

	// Labels should be exactly LabelsMaxWidth
	if widths.Labels != 10 {
		t.Errorf("Labels width incorrect: got %d, want 10", widths.Labels)
	}

	// Title should be at least TitleMinWidth
	if widths.Title < 30 {
		t.Errorf("Title width too small: got %d, want at least 30", widths.Title)
	}

	// Total width should not exceed terminal width (with some tolerance for separators)
	// ID + Status + P + Labels + Title + separators + padding
	totalUsed := widths.ID + widths.Status + widths.Priority + widths.Labels + widths.Title
	totalUsed += 4 * constraints.SeparatorWidth // 4 separators
	totalUsed += 5 * constraints.PaddingPerCell // 5 columns

	if totalUsed > constraints.TermWidth+10 { // Allow 10 char tolerance
		t.Errorf("Total width %d exceeds terminal width %d", totalUsed, constraints.TermWidth)
	}
}

func TestCalculateColumnWidths_WithAliases(t *testing.T) {
	tasks := []*types.Task{
		{
			TaskDisplayID: "test-1",
			TaskUUID:      "uuid1",
			Title:         "Test task",
			Aliases:       []string{"t", "task"},
			Axes: map[string]types.AxisStatus{
				"generic": {Effective: "wip"},
			},
		},
	}

	displayIDs := map[string]string{
		"uuid1": "test-1",
	}

	constraints := ColumnConstraints{
		TermWidth:      100,
		ShowAliases:    true,
		LabelsMaxWidth: 10,
		TitleMinWidth:  30,
		SeparatorWidth: 3,
		PaddingPerCell: 2,
	}

	widths := CalculateColumnWidths(tasks, displayIDs, constraints)

	// Aliases should be at least len("t, task") = 7
	if widths.Aliases < 7 {
		t.Errorf("Aliases width too small: got %d, want at least 7", widths.Aliases)
	}

	if !widths.HasAliases {
		t.Error("HasAliases should be true")
	}
}

func TestCalculateColumnWidths_LongIDs(t *testing.T) {
	tasks := []*types.Task{
		{
			TaskDisplayID: "very-long-project-name-123",
			TaskUUID:      "uuid1",
			Title:         "Task",
			Axes:          map[string]types.AxisStatus{},
		},
	}

	displayIDs := map[string]string{
		"uuid1": "very-long-project-name-123",
	}

	constraints := DefaultColumnConstraints(80, false)

	widths := CalculateColumnWidths(tasks, displayIDs, constraints)

	// ID should accommodate the long ID
	expectedIDWidth := len("very-long-project-name-123")
	if widths.ID < expectedIDWidth {
		t.Errorf("ID width too small for long ID: got %d, want at least %d", widths.ID, expectedIDWidth)
	}

	// Title should still get at least minimum width
	if widths.Title < constraints.TitleMinWidth {
		t.Errorf("Title width too small: got %d, want at least %d", widths.Title, constraints.TitleMinWidth)
	}
}

func TestCalculateColumnWidths_MinimumHeaderWidths(t *testing.T) {
	// Empty tasks should still respect header widths
	tasks := []*types.Task{}
	displayIDs := map[string]string{}
	constraints := DefaultColumnConstraints(80, false)

	widths := CalculateColumnWidths(tasks, displayIDs, constraints)

	// Minimum widths for headers
	if widths.ID < 2 {
		t.Errorf("ID width too small for header: got %d, want at least 2", widths.ID)
	}
	if widths.Status < 6 {
		t.Errorf("Status width too small for header: got %d, want at least 6", widths.Status)
	}
	if widths.Priority < 1 {
		t.Errorf("Priority width too small: got %d, want at least 1", widths.Priority)
	}
}

func TestCalculateColumnWidths_NarrowTerminal(t *testing.T) {
	tasks := []*types.Task{
		{
			TaskDisplayID: "test-1",
			TaskUUID:      "uuid1",
			Title:         "Task",
		},
	}

	displayIDs := map[string]string{
		"uuid1": "test-1",
	}

	// Very narrow terminal
	constraints := ColumnConstraints{
		TermWidth:      40,
		ShowAliases:    false,
		LabelsMaxWidth: 10,
		TitleMinWidth:  30,
		SeparatorWidth: 3,
		PaddingPerCell: 2,
	}

	widths := CalculateColumnWidths(tasks, displayIDs, constraints)

	// Title should still get minimum width even if terminal is narrow
	if widths.Title < constraints.TitleMinWidth {
		t.Errorf("Title width too small: got %d, want at least %d", widths.Title, constraints.TitleMinWidth)
	}
}
