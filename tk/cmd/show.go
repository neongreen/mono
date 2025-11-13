package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show [task-id...]",
	Aliases: []string{"view"},
	Short:   "Show task details",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		var tasks []*types.Task

		// Resolve and fetch all tasks
		for _, taskRef := range args {
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
			tasks = append(tasks, &taskCopy)
		}

		if jsonOutput {
			// For JSON, output array if multiple tasks, single object if one task
			if len(tasks) == 1 {
				output, err := json.MarshalIndent(tasks[0], "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal task: %w", err)
				}
				fmt.Println(string(output))
			} else {
				output, err := json.MarshalIndent(tasks, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal tasks: %w", err)
				}
				fmt.Println(string(output))
			}
		} else {
			// For human output, show each task with separator
			for i, task := range tasks {
				if i > 0 {
					fmt.Println("\n" + strings.Repeat("─", 60))
					fmt.Println()
				}
				renderTaskDetails(db, reducer, task)
			}
		}

		return nil
	},
}

// renderTaskDetails renders a human-readable view of a single task
func renderTaskDetails(db *database.DB, reducer *reducer.Reducer, task *types.Task) {
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
				// Render the display ID for the blocker
				blockerID, err := database.RenderTaskDisplayID(db, blocker.TaskUUID)
				if err != nil || blockerID == "" {
					blockerID = blocker.TaskUUID
				}
				title := ""
				if blocker.Title != "" {
					title = fmt.Sprintf(" - %s", blocker.Title)
				}
				fmt.Printf("  - %s%s\n", blockerID, title)
			}
		}
	}

	// Display all relations if present
	if task.Relations != nil {
		// Display tasks this one blocks (outgoing blocks)
		if len(task.Relations.Blocks.Out) > 0 {
			fmt.Println("Blocks:")
			for _, rel := range task.Relations.Blocks.Out {
				relID, err := database.RenderTaskDisplayID(db, rel.TaskUUID)
				if err != nil || relID == "" {
					relID = rel.TaskUUID
				}

				// Get title from reducer
				title := ""
				if relTask, ok := reducer.GetTask(rel.TaskUUID); ok {
					title = relTask.Title
				}

				if title != "" {
					if rel.Note != "" {
						fmt.Printf("  - %s - %s (%s)\n", relID, title, rel.Note)
					} else {
						fmt.Printf("  - %s - %s\n", relID, title)
					}
				} else {
					if rel.Note != "" {
						fmt.Printf("  - %s (%s)\n", relID, rel.Note)
					} else {
						fmt.Printf("  - %s\n", relID)
					}
				}
			}
		}

		// Display parent (if this is a subtask)
		if task.Relations.Subtask.Parent != "" {
			parentID, err := database.RenderTaskDisplayID(db, task.Relations.Subtask.Parent)
			if err != nil || parentID == "" {
				parentID = task.Relations.Subtask.Parent
			}

			// Get title from reducer
			title := ""
			if parentTask, ok := reducer.GetTask(task.Relations.Subtask.Parent); ok {
				title = parentTask.Title
			}

			if title != "" {
				fmt.Printf("Parent task: %s - %s\n", parentID, title)
			} else {
				fmt.Printf("Parent task: %s\n", parentID)
			}
		}

		// Display subtasks (children)
		if len(task.Relations.Subtask.Children) > 0 {
			fmt.Println("Subtasks:")
			for _, childUUID := range task.Relations.Subtask.Children {
				childID, err := database.RenderTaskDisplayID(db, childUUID)
				if err != nil || childID == "" {
					childID = childUUID
				}

				// Get title from reducer
				title := ""
				if childTask, ok := reducer.GetTask(childUUID); ok {
					title = childTask.Title
				}

				if title != "" {
					fmt.Printf("  - %s - %s\n", childID, title)
				} else {
					fmt.Printf("  - %s\n", childID)
				}
			}
		}

		// Display related tasks (combine out and in, avoid duplicates)
		relatedSet := make(map[string]string) // UUID -> note
		for _, rel := range task.Relations.Related.Out {
			relatedSet[rel.TaskUUID] = rel.Note
		}
		for _, rel := range task.Relations.Related.In {
			if _, exists := relatedSet[rel.TaskUUID]; !exists {
				relatedSet[rel.TaskUUID] = rel.Note
			}
		}

		if len(relatedSet) > 0 {
			fmt.Println("Related:")
			for uuid, note := range relatedSet {
				relID, err := database.RenderTaskDisplayID(db, uuid)
				if err != nil || relID == "" {
					relID = uuid
				}

				// Get title from reducer
				title := ""
				if relTask, ok := reducer.GetTask(uuid); ok {
					title = relTask.Title
				}

				if title != "" {
					if note != "" {
						fmt.Printf("  - %s - %s (%s)\n", relID, title, note)
					} else {
						fmt.Printf("  - %s - %s\n", relID, title)
					}
				} else {
					if note != "" {
						fmt.Printf("  - %s (%s)\n", relID, note)
					} else {
						fmt.Printf("  - %s\n", relID)
					}
				}
			}
		}

		// Display duplicates
		if len(task.Relations.Duplicate.Out) > 0 || len(task.Relations.Duplicate.In) > 0 {
			fmt.Println("Duplicates:")
			for _, rel := range task.Relations.Duplicate.Out {
				dupID, err := database.RenderTaskDisplayID(db, rel.TaskUUID)
				if err != nil || dupID == "" {
					dupID = rel.TaskUUID
				}

				title := ""
				if dupTask, ok := reducer.GetTask(rel.TaskUUID); ok {
					title = dupTask.Title
				}

				if title != "" {
					fmt.Printf("  - %s - %s\n", dupID, title)
				} else {
					fmt.Printf("  - %s\n", dupID)
				}
			}
			for _, rel := range task.Relations.Duplicate.In {
				dupID, err := database.RenderTaskDisplayID(db, rel.TaskUUID)
				if err != nil || dupID == "" {
					dupID = rel.TaskUUID
				}

				title := ""
				if dupTask, ok := reducer.GetTask(rel.TaskUUID); ok {
					title = dupTask.Title
				}

				if title != "" {
					fmt.Printf("  - %s - %s\n", dupID, title)
				} else {
					fmt.Printf("  - %s\n", dupID)
				}
			}
		}

		// Display supersedes
		if len(task.Relations.Supersedes.Out) > 0 {
			fmt.Println("Supersedes:")
			for _, rel := range task.Relations.Supersedes.Out {
				supID, err := database.RenderTaskDisplayID(db, rel.TaskUUID)
				if err != nil || supID == "" {
					supID = rel.TaskUUID
				}

				title := ""
				if supTask, ok := reducer.GetTask(rel.TaskUUID); ok {
					title = supTask.Title
				}

				if title != "" {
					fmt.Printf("  - %s - %s\n", supID, title)
				} else {
					fmt.Printf("  - %s\n", supID)
				}
			}
		}
		if len(task.Relations.Supersedes.In) > 0 {
			fmt.Println("Superseded by:")
			for _, rel := range task.Relations.Supersedes.In {
				supID, err := database.RenderTaskDisplayID(db, rel.TaskUUID)
				if err != nil || supID == "" {
					supID = rel.TaskUUID
				}

				title := ""
				if supTask, ok := reducer.GetTask(rel.TaskUUID); ok {
					title = supTask.Title
				}

				if title != "" {
					fmt.Printf("  - %s - %s\n", supID, title)
				} else {
					fmt.Printf("  - %s\n", supID)
				}
			}
		}
	}

	// Display containers (v6+)
	version, _ := db.GetDBVersion()
	if version >= 6 {
		rows, err := db.Db.Query(`
			SELECT c.id, c.primitive, c.kind, c.name
			FROM container_members cm
			JOIN containers c ON cm.container_id = c.id
			WHERE cm.item_id = ? AND cm.removed = 0 AND c.removed = 0
			ORDER BY c.primitive, c.id
		`, task.TaskUUID)
		if err == nil {
			defer rows.Close()

			var containers []string
			for rows.Next() {
				var id, primitive, kind, name string
				if err := rows.Scan(&id, &primitive, &kind, &name); err == nil {
					containers = append(containers, fmt.Sprintf("%s (%s)", id, primitive))
				}
			}

			if len(containers) > 0 {
				fmt.Printf("\nIn containers: %s\n", strings.Join(containers, ", "))
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

	// Display attachments
	if len(task.Attachments) > 0 {
		fmt.Println("\nAttachments:")
		for _, att := range task.Attachments {
			sizeStr := formatAttachmentSize(att.Size)
			fmt.Printf("  %s: %s (%s)\n", att.ID, att.Filename, sizeStr)
			if att.Description != "" {
				fmt.Printf("      %s\n", att.Description)
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

// formatAttachmentSize formats a file size for display
func formatAttachmentSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)

	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
