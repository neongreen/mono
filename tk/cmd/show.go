package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show [task-id]",
	Aliases: []string{"view"},
	Short:   "Show task details",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		config, err := config_pkg.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return err
		}

		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
		if err != nil {
			return err
		}

		task, ok := reducer.GetTask(taskUUID)
		if !ok {
			return fmt.Errorf("task not found: %s", taskRef)
		}

		displayID, err := database.RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		taskCopy := *task
		taskCopy.TaskDisplayID = displayID

		if jsonOutput {
			output, err := json.MarshalIndent(taskCopy, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal task: %w", err)
			}
			fmt.Println(string(output))
		} else {
			renderTaskDetails(db, &taskCopy)
		}

		return nil
	},
}

// renderTaskDetails renders a human-readable view of a single task
func renderTaskDetails(db *database.DB, task *types.Task) {
	fmt.Printf("Task: %s\n", boldText(task.TaskDisplayID))

	if task.Title != "" {
		fmt.Printf("Title: %s\n", task.Title)
	}

	// Display status from generic axis
	if axis, ok := task.Axes["generic"]; ok && axis.Effective != "" {
		fmt.Printf("Status: %s\n", colorizeStatus(axis.Effective))
	}

	// Display aliases
	if len(task.Aliases) > 0 {
		var aliases []string
		for _, alias := range task.Aliases {
			aliases = append(aliases, database.FormatTaskID(db, alias))
		}
		sort.Strings(aliases)
		fmt.Printf("Aliases: %s\n", strings.Join(aliases, ", "))
	}

	// Display blocked status
	if task.Blocked {
		fmt.Printf("Blocked: %s\n", redText("yes"))
		if len(task.Blockers) > 0 {
			fmt.Println("Blocked by:")
			for _, blocker := range task.Blockers {
				blockerID := blocker.TaskDisplayID
				title := ""
				if blocker.Title != "" {
					title = fmt.Sprintf(" - %s", blocker.Title)
				}
				fmt.Printf("  - %s%s\n", blockerID, title)
			}
		}
	}

	// Display notes
	if len(task.Notes) > 0 {
		fmt.Println("\nNotes:")
		for i, note := range task.Notes {
			if i > 0 {
				fmt.Println()
			}
			if note.Actor != "" || !note.Timestamp.IsZero() {
				header := []string{}
				if note.Actor != "" {
					header = append(header, note.Actor)
				}
				if !note.Timestamp.IsZero() {
					header = append(header, note.Timestamp.Format("2006-01-02 15:04:05"))
				}
				fmt.Printf("  [%s]\n", strings.Join(header, " - "))
			}
			if note.Markdown != "" {
				// Indent note content
				lines := strings.SplitSeq(note.Markdown, "\n")
				for line := range lines {
					fmt.Printf("  %s\n", line)
				}
			}
		}
	}

	// Display metadata
	if len(task.Metadata) > 0 {
		fmt.Println("\nMetadata:")

		// Sort keys for consistent display
		var keys []string
		for key := range task.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			meta := task.Metadata[key]

			// Display effective value
			var effectiveValue any
			if err := json.Unmarshal(meta.Effective, &effectiveValue); err == nil {
				fmt.Printf("  %s: %v", key, formatMetadataValue(effectiveValue))

				// Count tentative claims
				tentativeCount := 0
				for _, claim := range meta.Claims {
					if claim.Tentative {
						tentativeCount++
					}
				}

				if tentativeCount > 0 {
					fmt.Printf(" (+ %d tentative claim", tentativeCount)
					if tentativeCount > 1 {
						fmt.Printf("s")
					}
					fmt.Printf(")")
				}
				fmt.Println()
			}
		}
	}
}

// formatMetadataValue formats a metadata value for display
func formatMetadataValue(value any) string {
	switch v := value.(type) {
	case []any:
		// Format arrays
		var items []string
		for _, item := range v {
			items = append(items, fmt.Sprintf("%v", item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case map[string]any:
		// Format objects as JSON
		data, _ := json.Marshal(v)
		return string(data)
	default:
		return fmt.Sprintf("%v", v)
	}
}
