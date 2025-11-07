package cmd

import (
	"github.com/neongreen/mono/tk/internal/types"
)

// ColumnWidths holds the calculated widths for table columns
type ColumnWidths struct {
	ID         int
	Aliases    int
	Status     int
	Priority   int
	Labels     int
	Title      int
	HasAliases bool
}

// ColumnConstraints defines the constraints for column sizing
type ColumnConstraints struct {
	TermWidth      int
	ShowAliases    bool
	LabelsMaxWidth int
	TitleMinWidth  int
	SeparatorWidth int // Width of " │ " separator between columns
	PaddingPerCell int // Additional padding within each cell
}

// DefaultColumnConstraints returns sensible defaults
func DefaultColumnConstraints(termWidth int, showAliases bool) ColumnConstraints {
	return ColumnConstraints{
		TermWidth:      termWidth,
		ShowAliases:    showAliases,
		LabelsMaxWidth: 10,
		TitleMinWidth:  30,
		SeparatorWidth: 1, // "│" - the table library adds spaces around it
		PaddingPerCell: 2, // The table library adds 1 space on each side of content
	}
}

// CalculateColumnWidths computes optimal column widths based on actual task data
func CalculateColumnWidths(tasks []*types.Task, displayIDs map[string]string, constraints ColumnConstraints) ColumnWidths {
	widths := ColumnWidths{
		HasAliases: constraints.ShowAliases,
	}

	// Find maximum widths needed for non-wrapping columns
	for _, task := range tasks {
		// ID width
		displayID := displayIDs[task.TaskUUID]
		if displayID == "" {
			displayID = task.TaskID
		}
		if len(displayID) > widths.ID {
			widths.ID = len(displayID)
		}

		// Status width
		if axis, ok := task.Axes["generic"]; ok {
			statusLen := len(axis.Effective)
			if statusLen > widths.Status {
				widths.Status = statusLen
			}
		}

		// Priority width
		if meta, ok := task.Metadata["priority"]; ok {
			// Priority is typically 1 char ("1", "2", etc)
			priorityLen := len(string(meta.Effective))
			// But let's be conservative and check actual length
			if priorityLen > 10 {
				priorityLen = 1 // JSON encoded, so single digit
			}
			if priorityLen > widths.Priority {
				widths.Priority = priorityLen
			}
		}

		// Aliases width (if showing)
		if constraints.ShowAliases && len(task.Aliases) > 0 {
			// Estimate: "alias1, alias2, ..."
			aliasesLen := 0
			for i, alias := range task.Aliases {
				if i > 0 {
					aliasesLen += 2 // ", "
				}
				// Rough estimate: project prefix + number
				aliasesLen += len(alias)
			}
			if aliasesLen > widths.Aliases {
				widths.Aliases = aliasesLen
			}
		}
	}

	// Ensure minimum widths for headers
	if widths.ID < 2 {
		widths.ID = 2 // "ID"
	}
	if widths.Status < 6 {
		widths.Status = 6 // "STATUS"
	}
	if widths.Priority < 1 {
		widths.Priority = 1 // "P"
	}
	if constraints.ShowAliases && widths.Aliases < 7 {
		widths.Aliases = 7 // "Aliases"
	}

	// Labels gets fixed max width
	widths.Labels = constraints.LabelsMaxWidth

	// Calculate space used by fixed columns
	numColumns := 5 // ID, Status, P, Labels, Title
	if constraints.ShowAliases {
		numColumns = 6
	}
	numSeparators := numColumns - 1

	usedSpace := widths.ID + widths.Status + widths.Priority + widths.Labels
	if constraints.ShowAliases {
		usedSpace += widths.Aliases
	}
	usedSpace += numSeparators * constraints.SeparatorWidth
	usedSpace += numColumns * constraints.PaddingPerCell

	// Title gets all remaining space, but at least TitleMinWidth
	remainingSpace := constraints.TermWidth - usedSpace
	widths.Title = max(remainingSpace, constraints.TitleMinWidth)

	return widths
}
