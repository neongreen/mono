package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var relateCmd = &cobra.Command{
	Use:   "relate",
	Short: "Manage task relations",
	Long:  `Add, remove, or view relations between tasks (blocks, subtasks, related, etc.)`,
}

var relateAddCmd = &cobra.Command{
	Use:   "add [source-task] [relation-type] [target-task]",
	Short: "Add a relation between two tasks",
	Long: `Add a relation between two tasks.

Relation types:
  blocks        - Source task blocks target task
  blocked_by    - Source task is blocked by target task (inverse of blocks)
  subtask       - Target task is a subtask of source task
  parent        - Source task is a parent of target task (inverse of subtask)
  related       - Tasks are related
  duplicate_of  - Source task is a duplicate of target task
  supersedes    - Source task supersedes target task

Examples:
  tk relate add tk-1 blocks tk-2
  tk relate add tk-3 subtask tk-1
  tk relate add tk-4 related tk-5 --note "Both implement auth"`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcTaskID := args[0]
		relationType := args[1]
		dstTaskID := args[2]
		note, _ := cmd.Flags().GetString("note")

		// Validate relation type
		validTypes := []string{"blocks", "blocked_by", "subtask", "parent", "related", "duplicate_of", "supersedes"}
		valid := false
		for _, t := range validTypes {
			if relationType == t {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid relation type %q, must be one of: %s", relationType, strings.Join(validTypes, ", "))
		}

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve both task IDs to UUIDs
		srcUUID, err := db.ResolveTaskIDToUUID(srcTaskID)
		if err != nil {
			return fmt.Errorf("failed to resolve source task %q: %w", srcTaskID, err)
		}

		dstUUID, err := db.ResolveTaskIDToUUID(dstTaskID)
		if err != nil {
			return fmt.Errorf("failed to resolve target task %q: %w", dstTaskID, err)
		}

		// Get current user
		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		// Generate event ID
		eventID, err := GenerateEventID(db)
		if err != nil {
			return err
		}

		// Get next Lamport timestamp
		lamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		// Create relation.add event
		payload := RelationAddPayload{
			Src:  srcUUID,
			Type: relationType,
			Dst:  dstUUID,
			Note: note,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		now := time.Now()
		event := Event{
			ID:        eventID,
			TS:        lamportTS,
			CreatedAt: now,
			Actor:     currentUser,
			Role:      "human",
			Kind:      "relation.add",
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return err
		}

		fmt.Printf("Added relation: %s %s %s\n", srcTaskID, relationType, dstTaskID)
		return nil
	},
}

var relateRemoveCmd = &cobra.Command{
	Use:   "remove [source-task] [relation-type] [target-task]",
	Short: "Remove a relation between two tasks",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcTaskID := args[0]
		relationType := args[1]
		dstTaskID := args[2]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve both task IDs to UUIDs
		srcUUID, err := db.ResolveTaskIDToUUID(srcTaskID)
		if err != nil {
			return fmt.Errorf("failed to resolve source task %q: %w", srcTaskID, err)
		}

		dstUUID, err := db.ResolveTaskIDToUUID(dstTaskID)
		if err != nil {
			return fmt.Errorf("failed to resolve target task %q: %w", dstTaskID, err)
		}

		// Get current user
		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		// Generate event ID
		eventID, err := GenerateEventID(db)
		if err != nil {
			return err
		}

		// Get next Lamport timestamp
		lamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		// Create relation.remove event
		payload := RelationRemovePayload{
			Src:  srcUUID,
			Type: relationType,
			Dst:  dstUUID,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		now := time.Now()
		event := Event{
			ID:        eventID,
			TS:        lamportTS,
			CreatedAt: now,
			Actor:     currentUser,
			Role:      "human",
			Kind:      "relation.remove",
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return err
		}

		fmt.Printf("Removed relation: %s %s %s\n", srcTaskID, relationType, dstTaskID)
		return nil
	},
}

var dupCmd = &cobra.Command{
	Use:   "dup [task-a] [task-b]",
	Short: "Mark two tasks as duplicates",
	Long:  `Mark two tasks as duplicates of each other (adds duplicate_of relations in both directions).`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskA := args[0]
		taskB := args[1]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve both task IDs to UUIDs
		uuidA, err := db.ResolveTaskIDToUUID(taskA)
		if err != nil {
			return fmt.Errorf("failed to resolve task %q: %w", taskA, err)
		}

		uuidB, err := db.ResolveTaskIDToUUID(taskB)
		if err != nil {
			return fmt.Errorf("failed to resolve task %q: %w", taskB, err)
		}

		// Get current user
		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		// Create two relation.add events (bidirectional)
		for i, pair := range [][2]string{{uuidA, uuidB}, {uuidB, uuidA}} {
			eventID, err := GenerateEventID(db)
			if err != nil {
				return err
			}

			lamportTS, err := db.GetNextLamportTS()
			if err != nil {
				return err
			}

			payload := RelationAddPayload{
				Src:  pair[0],
				Type: "duplicate_of",
				Dst:  pair[1],
			}
			payloadJSON, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to marshal payload: %w", err)
			}

			now := time.Now()
			event := Event{
				ID:        eventID,
				TS:        lamportTS,
				CreatedAt: now,
				Actor:     currentUser,
				Role:      "human",
				Kind:      "relation.add",
				Payload:   payloadJSON,
			}

			if err := db.InsertEvent(event); err != nil {
				return err
			}

			// Small delay to ensure different timestamps
			if i == 0 {
				time.Sleep(time.Millisecond)
			}
		}

		fmt.Printf("Marked %s and %s as duplicates\n", taskA, taskB)
		return nil
	},
}

func init() {
	relateAddCmd.Flags().String("note", "", "Optional note for the relation")
	relateCmd.AddCommand(relateAddCmd)
	relateCmd.AddCommand(relateRemoveCmd)
}
