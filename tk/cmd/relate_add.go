package cmd

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var relateAddCmd = &cobra.Command{
	Use:   "relate-add [source-task] [relation-type] [target-task]",
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
  tk relate-add tk-1 blocks tk-2
  tk relate-add tk-3 subtask tk-1
  tk relate-add tk-4 related tk-5 --note "Both implement auth"`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcTaskID := args[0]
		relationType := args[1]
		dstTaskID := args[2]
		note, _ := cmd.Flags().GetString("note")

		// Validate relation type
		validTypes := []string{"blocks", "blocked_by", "subtask", "parent", "related", "duplicate_of", "supersedes"}
		valid := slices.Contains(validTypes, relationType)
		if !valid {
			return fmt.Errorf("invalid relation type %q, must be one of: %s", relationType, strings.Join(validTypes, ", "))
		}

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve both task IDs to UUIDs
		srcUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(srcTaskID))
		if err != nil {
			return fmt.Errorf("failed to resolve source task %q: %w", srcTaskID, err)
		}

		dstUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(dstTaskID))
		if err != nil {
			return fmt.Errorf("failed to resolve target task %q: %w", dstTaskID, err)
		}

		// Normalize inverse relation types to canonical forms
		// blocked_by(a,b) -> blocks(b,a)
		// parent(a,b) -> subtask(a,b) (a is parent, b is child)
		normalizedType := relationType
		normalizedSrc := srcUUID
		normalizedDst := dstUUID

		switch relationType {
		case "blocked_by":
			normalizedType = "blocks"
			normalizedSrc = dstUUID
			normalizedDst = srcUUID
		case "parent":
			normalizedType = "subtask"
			// parent(a,b) means a is parent of b, which is subtask(a,b)
			// Keep src and dst as-is
		}

		srcUUID = normalizedSrc
		dstUUID = normalizedDst
		relationType = normalizedType

		// Get current user
		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Generate event ID
		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return err
		}

		// Get next Lamport timestamp
		lamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		// Create relation.add event
		payload := types.RelationAddPayload{
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

		srcDisplay, err := database.RenderTaskDisplayID(db, srcUUID)
		if err != nil {
			srcDisplay = srcTaskID
		}
		dstDisplay, err := database.RenderTaskDisplayID(db, dstUUID)
		if err != nil {
			dstDisplay = dstTaskID
		}

		fmt.Printf("Added relation: %s %s %s\n", srcDisplay, relationType, dstDisplay)
		return nil
	},
}

func init() {
	relateAddCmd.Flags().String("note", "", "Optional note for the relation")
}
