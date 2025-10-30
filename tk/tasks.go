package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func createTask(db *DB, cmd *cobra.Command, title string) error {
	projectFlag, _ := cmd.Flags().GetString("project")

	currentUser, err := getCurrentUser()
	if err != nil {
		return err
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return err
	}

	projectUID, err := resolveProjectByAlias(db, projectFlag)
	if err != nil {
		return fmt.Errorf("project/alias %q not found. Create it first with: tk project create <name> --alias %s", projectFlag, projectFlag)
	}

	taskUID := NewTaskUID()

	// Compute proposed number (max + 1)
	var maxNumber int64
	err = db.db.QueryRow(`
		SELECT COALESCE(MAX(number), 0) FROM task_numbers 
		WHERE project_uid = ?
	`, projectUID).Scan(&maxNumber)
	if err != nil {
		return fmt.Errorf("failed to get max number: %w", err)
	}
	proposedNumber := maxNumber + 1

	payload := TaskCreatedPayload{
		TaskUID:        string(taskUID),
		ProjectUID:     projectUID,
		ProposedNumber: proposedNumber,
		CreatedNode:    nodeID,
		Title:          title,
		CreatedBy:      currentUser,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := Event{
		ID:        generateEventID(db),
		TS:        getNextLamportTimestamp(db),
		CreatedAt: time.Now(),
		Actor:     currentUser,
		Role:      "human",
		Kind:      string(EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	if err := db.ProjectTaskCreatedEvent(event); err != nil {
		return fmt.Errorf("failed to project task: %w", err)
	}

	numberPayload := TaskNumberSetPayload{
		TaskUID:    string(taskUID),
		ProjectUID: projectUID,
		Number:     proposedNumber,
		Reason:     "initial",
	}
	numberPayloadJSON, err := json.Marshal(numberPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal number payload: %w", err)
	}

	numberEvent := Event{
		ID:        generateEventID(db),
		TS:        getNextLamportTimestamp(db),
		CreatedAt: time.Now(),
		Actor:     currentUser,
		Role:      "human",
		Kind:      string(EventKindTaskNumberSet),
		Payload:   numberPayloadJSON,
	}

	if err := db.InsertEvent(numberEvent); err != nil {
		return fmt.Errorf("failed to insert number event: %w", err)
	}

	if err := db.ProjectTaskNumberSetEvent(numberEvent); err != nil {
		return fmt.Errorf("failed to project task number: %w", err)
	}

	displayID := fmt.Sprintf("%s-%d", projectFlag, proposedNumber)
	fmt.Printf("Created task %s: %s\n", displayID, title)
	return nil
}
