package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/utils"

	relate_pkg "github.com/neongreen/mono/tk/cmd/relate"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var relateCmd = &cobra.Command{
	Use:   "relate",
	Short: "Manage task relations",
	Long:  `Add, remove, or view relations between tasks (blocks, subtasks, related, etc.)`,
}

var dupCmd = &cobra.Command{
	Use:   "dup [task-a] [task-b]",
	Short: "Mark two tasks as duplicates",
	Long:  `Mark two tasks as duplicates of each other (adds duplicate_of relations in both directions).`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskA := args[0]
		taskB := args[1]

		db, err := database.OpenExistingDB()
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
		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create two relation.add events (bidirectional)
		for i, pair := range [][2]string{{uuidA, uuidB}, {uuidB, uuidA}} {
			eventID, err := database.GenerateEventID(db)
			if err != nil {
				return err
			}

			lamportTS, err := db.GetNextLamportTS()
			if err != nil {
				return err
			}

			payload := types.RelationAddPayload{
				Src:  pair[0],
				Type: "duplicate_of",
				Dst:  pair[1],
			}
			payloadJSON, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to marshal payload: %w", err)
			}

			now := time.Now()
			event := types.Event{
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
	relateCmd.AddCommand(relate_pkg.AddCmd)
	relateCmd.AddCommand(relate_pkg.RemoveCmd)
	relateCmd.AddCommand(relate_pkg.LsCmd)
}
