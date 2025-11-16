package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

type taskIdentity struct {
	TaskUID      string    `json:"task_uid"`
	DisplayID    string    `json:"display_id"`
	ProjectUID   string    `json:"project_uid"`
	ProjectAlias string    `json:"project_alias,omitempty"`
	Number       int64     `json:"number"`
	Collides     bool      `json:"collides"`
	CreatedNode  string    `json:"created_node"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	Title        string    `json:"title"`
}

var debugIdCmd = &cobra.Command{
	Use:   "debug-id <task-ref>",
	Short: "Show canonical identifiers for a task",
	Args:  cobra.ExactArgs(1),
	RunE:  idCmdImpl,
}

var idCmd = &cobra.Command{
	Use:   "id <task-ref>",
	Short: "Show canonical identifiers for a task (alias for debug-id)",
	Args:  cobra.ExactArgs(1),
	RunE:  idCmdImpl,
}

func idCmdImpl(cmd *cobra.Command, args []string) error {
	taskRef := args[0]
	jsonOutput, _ := cmd.Flags().GetBool("json")

	db, err := database.OpenExistingDB()
	if err != nil {
		return err
	}
	defer db.Close()

	identity, err := describeTask(db, taskRef)
	if err != nil {
		return err
	}

	if jsonOutput {
		// JSON output mode
		output, err := json.MarshalIndent(identity, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal task identity: %w", err)
		}
		fmt.Println(string(output))
	} else {
		// Human-readable output mode
		fmt.Printf("Task: %s\n", identity.DisplayID)
		fmt.Printf("Title: %s\n", identity.Title)
		fmt.Println()

		fmt.Println("Identifiers:")
		fmt.Printf("  Display ID:  %s\n", identity.DisplayID)
		fmt.Printf("  Task UID:    %s\n", identity.TaskUID)
		fmt.Println()

		fmt.Println("Project:")
		if identity.ProjectAlias != "" {
			fmt.Printf("  Alias:       %s\n", identity.ProjectAlias)
		}
		fmt.Printf("  Project UID: %s\n", identity.ProjectUID)
		fmt.Printf("  Number:      %d\n", identity.Number)
		if identity.Collides {
			fmt.Printf("  Collides:    Yes (warning: number collision detected)\n")
		} else {
			fmt.Printf("  Collides:    No\n")
		}
		fmt.Println()

		fmt.Println("Created:")
		fmt.Printf("  By:          %s\n", identity.CreatedBy)
		fmt.Printf("  At:          %s\n", identity.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Node:        %s\n", identity.CreatedNode)
	}

	return nil
}

func init() {
	debugIdCmd.Flags().Bool("json", false, "Output as JSON")
	idCmd.Flags().Bool("json", false, "Output as JSON")
}

func describeTask(db *database.DB, ref string) (*taskIdentity, error) {
	taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(ref))
	if err != nil {
		return nil, err
	}

	var projectUID, createdNode, title, createdBy string
	var createdAtUnix int64
	err = db.Db.QueryRow(`
        SELECT project_uid, created_node, title, created_at, created_by
        FROM tasks
        WHERE task_uid = ?
    `, taskUID).Scan(&projectUID, &createdNode, &title, &createdAtUnix, &createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to load task %s: %w", taskUID, err)
	}

	var number int64
	err = db.Db.QueryRow(`
        SELECT number FROM task_numbers WHERE task_uid = ?
    `, taskUID).Scan(&number)
	if err != nil {
		return nil, fmt.Errorf("failed to load task number for %s: %w", taskUID, err)
	}

	alias, err := database.PreferredAliasForProject(db, types.ProjectUID(projectUID))
	if err != nil {
		return nil, err
	}

	displayID, err := database.RenderTaskDisplayID(db, taskUID)
	if err != nil {
		displayID = taskUID
	}

	collides, err := database.HasNumberCollision(db, projectUID, number)
	if err != nil {
		return nil, err
	}

	return &taskIdentity{
		TaskUID:      taskUID,
		DisplayID:    displayID,
		ProjectUID:   projectUID,
		ProjectAlias: alias,
		Number:       number,
		Collides:     collides,
		CreatedNode:  createdNode,
		CreatedBy:    createdBy,
		CreatedAt:    time.Unix(createdAtUnix, 0),
		Title:        title,
	}, nil
}
