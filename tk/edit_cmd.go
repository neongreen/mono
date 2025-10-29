package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <task> <field> <value>",
	Short: "Edit task fields (number, title, status)",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		field := strings.ToLower(args[1])
		value := strings.Join(args[2:], " ")

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		return editTask(db, taskRef, field, value)
	},
}

func editTask(db *DB, taskRef, field, value string) error {
	taskUID, err := ResolveTaskReference(db, taskRef)
	if err != nil {
		return err
	}

	currentUser, err := getCurrentUser()
	if err != nil {
		return err
	}

	switch field {
	case "number":
		return editTaskNumber(db, taskUID, value, currentUser)
	case "title":
		return editTaskTitle(db, taskUID, value, currentUser)
	case "status":
		return editTaskStatus(db, taskUID, value, currentUser)
	default:
		return fmt.Errorf("unsupported field %s (supported: number, title, status)", field)
	}
}

func editTaskStatus(db *DB, taskUID string, value string, actor string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("status cannot be empty")
	}

	lamport, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	payload := TaskStatusSetPayload{
		TaskUUID: taskUID,
		Axis:     "generic",
		State:    value,
		Role:     "human",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.status.set payload: %w", err)
	}

	event := Event{
		ID:        string(NewEventID()),
		TS:        lamport,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(EventKindTaskStatusSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert task.status.set event: %w", err)
	}

	displayID, err := RenderTaskDisplayID(db, taskUID)
	if err != nil {
		displayID = taskUID
	}

	fmt.Printf("Updated status for %s to %s\n", displayID, value)
	return nil
}

func editTaskNumber(db *DB, taskUID string, value string, actor string) error {
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil || number <= 0 {
		return fmt.Errorf("invalid number %q", value)
	}

	projectUID, err := projectForTask(db, taskUID)
	if err != nil {
		return err
	}

	lamport, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	payload := TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     number,
		Reason:     "edit",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.number.set payload: %w", err)
	}

	event := Event{
		ID:        string(NewEventID()),
		TS:        lamport,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(EventKindTaskNumberSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert task.number.set event: %w", err)
	}

	if err := db.ProjectTaskNumberSetEvent(event); err != nil {
		return fmt.Errorf("failed to project task.number.set event: %w", err)
	}

	displayID, err := RenderTaskDisplayID(db, taskUID)
	if err != nil {
		return err
	}

	fmt.Printf("Updated number for %s to %d\n", displayID, number)
	return nil
}

func editTaskTitle(db *DB, taskUID string, value string, actor string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("title cannot be empty")
	}

	lamport, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	payload := TaskTitleSetPayload{
		TaskUID: taskUID,
		Title:   value,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.title.set payload: %w", err)
	}

	event := Event{
		ID:        string(NewEventID()),
		TS:        lamport,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(EventKindTaskTitleSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert task.title.set event: %w", err)
	}

	if err := db.ProjectTaskTitleSetEvent(event); err != nil {
		return fmt.Errorf("failed to project task.title.set event: %w", err)
	}

	displayID, err := RenderTaskDisplayID(db, taskUID)
	if err != nil {
		return err
	}

	fmt.Printf("Updated title for %s\n", displayID)
	return nil
}

func projectForTask(db *DB, taskUID string) (string, error) {
	var projectUID string
	err := db.db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, taskUID).Scan(&projectUID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("task %s not found in tasks table", taskUID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to load task %s: %w", taskUID, err)
	}
	return projectUID, nil
}
