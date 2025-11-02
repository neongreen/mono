package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"

	"github.com/spf13/cobra"
)

func createTask(db *database.DB, cmd *cobra.Command, title string) error {
	projectFlag, _ := cmd.Flags().GetString("project")

	currentUser, err := getCurrentUser()
	if err != nil {
		return err
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return err
	}

	projectUID, err := database.ResolveProjectRef(db, types.NewProjectRef(projectFlag))
	if err != nil {
		return fmt.Errorf("project/alias %q not found. Create it first with: tk project create <name> --alias %s", projectFlag, projectFlag)
	}

	taskUID := types.NewTaskUID()

	// Compute proposed number (max + 1)
	var maxNumber int64
	err = db.Db.QueryRow(`
		SELECT COALESCE(MAX(number), 0) FROM task_numbers
		WHERE project_uid = ?
	`, projectUID.String()).Scan(&maxNumber)
	if err != nil {
		return fmt.Errorf("failed to get max number: %w", err)
	}
	proposedNumber := maxNumber + 1

	payload := types.TaskCreatedPayload{
		TaskUID:        string(taskUID),
		ProjectUID:     projectUID.String(),
		ProposedNumber: proposedNumber,
		CreatedNode:    nodeID,
		Title:          title,
		CreatedBy:      currentUser,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := types.Event{
		ID:        generateEventID(db),
		TS:        getNextLamportTimestamp(db),
		CreatedAt: time.Now(),
		Actor:     currentUser,
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	if err := db.ProjectTaskCreatedEvent(event); err != nil {
		return fmt.Errorf("failed to project task: %w", err)
	}

	numberPayload := types.TaskNumberSetPayload{
		TaskUID:    string(taskUID),
		ProjectUID: projectUID.String(),
		Number:     proposedNumber,
		Reason:     "initial",
	}
	numberPayloadJSON, err := json.Marshal(numberPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal number payload: %w", err)
	}

	numberEvent := types.Event{
		ID:        generateEventID(db),
		TS:        getNextLamportTimestamp(db),
		CreatedAt: time.Now(),
		Actor:     currentUser,
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   numberPayloadJSON,
	}

	if err := db.InsertEvent(numberEvent); err != nil {
		return fmt.Errorf("failed to insert number event: %w", err)
	}

	if err := db.ProjectTaskNumberSetEvent(numberEvent); err != nil {
		return fmt.Errorf("failed to project task number: %w", err)
	}

	// Get a friendly display name (preferred alias, or project name, or UID as fallback)
	displayPrefix, err := database.PreferredAliasForProject(db, projectUID)
	if err != nil {
		return fmt.Errorf("failed to get display prefix: %w", err)
	}
	if displayPrefix == "" {
		// No alias found, try to get project name
		var projectName string
		err = db.Db.QueryRow(`SELECT name FROM projects WHERE project_uid = ?`, projectUID.String()).Scan(&projectName)
		if err == nil && projectName != "" {
			displayPrefix = projectName
		} else {
			// Fallback to UID
			displayPrefix = projectUID.String()
		}
	}

	displayID := fmt.Sprintf("%s-%d", displayPrefix, proposedNumber)
	fmt.Printf("Created task %s: %s\n", displayID, title)
	return nil
}
