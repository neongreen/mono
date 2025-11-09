package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// MoveOptions contains options for moving a task
type MoveOptions struct {
	Mode         string // "keep", "auto", or "force"
	ForceNumber  int64
	OnCollision  string // "fail" or "auto"
	TargetNumber *int64 // Inline number from alias:number syntax
}

// Move relocates a task from one project to another
func Move(db *database.DB, taskUID, toProjectUID string, opts MoveOptions, actor string, clk clock.Clock) error {
	fromProjectUID, oldNumber, err := GetProjectAndNumberForTask(db, taskUID)
	if err != nil {
		return err
	}

	// If moving to same project with keep mode and no explicit number, it's a no-op
	if fromProjectUID == toProjectUID && opts.Mode == "keep" && !hasExplicitNumber(opts) {
		return fmt.Errorf("task already in project")
	}

	lamport, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	// Determine number policy
	// IMPORTANT: We resolve "auto" mode HERE (at event creation time) to make it deterministic.
	// The projection will just read the number from the event payload.
	numberPolicy := types.NumberPolicyPayload{Mode: opts.Mode}
	switch opts.Mode {
	case "keep":
		numberPolicy.Number = oldNumber
		// Check for collision if in keep mode
		if opts.OnCollision == "auto" {
			collision, err := CheckNumberCollision(db, toProjectUID, oldNumber, taskUID)
			if err != nil {
				return err
			}
			if collision {
				// Compute the next available number NOW
				var maxNumber int64
				err = db.Db.QueryRow(`
					SELECT COALESCE(MAX(number), 0) FROM task_numbers
					WHERE project_uid = ?
				`, toProjectUID).Scan(&maxNumber)
				if err != nil {
					return fmt.Errorf("failed to compute next number: %w", err)
				}
				// Convert to force mode with the computed number
				numberPolicy.Mode = "force"
				numberPolicy.Number = maxNumber + 1
			}
		} else {
			// Fail on collision
			collision, err := CheckNumberCollision(db, toProjectUID, oldNumber, taskUID)
			if err != nil {
				return err
			}
			if collision {
				return fmt.Errorf("number %d already exists in target project; use --auto or --force", oldNumber)
			}
		}
	case "auto":
		// Resolve "auto" mode to a concrete number NOW (deterministic)
		var maxNumber int64
		err = db.Db.QueryRow(`
			SELECT COALESCE(MAX(number), 0) FROM task_numbers
			WHERE project_uid = ?
		`, toProjectUID).Scan(&maxNumber)
		if err != nil {
			return fmt.Errorf("failed to compute next number: %w", err)
		}
		// Convert to force mode with the computed number
		numberPolicy.Mode = "force"
		numberPolicy.Number = maxNumber + 1
	case "force":
		numberPolicy.Number = opts.ForceNumber
	default:
		return fmt.Errorf("unsupported number policy mode %s", opts.Mode)
	}

	payload := types.TaskRelocatePayload{
		TaskUID:        taskUID,
		FromProjectUID: fromProjectUID,
		ToProjectUID:   toProjectUID,
		NumberPolicy:   numberPolicy,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.relocate payload: %w", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        lamport,
		CreatedAt: clk.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskRelocate),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert task.relocate event: %w", err)
	}

	if err := db.ProjectTaskRelocateEvent(event); err != nil {
		return fmt.Errorf("failed to project task.relocate event: %w", err)
	}

	return nil
}

// hasExplicitNumber checks if the move options specify an explicit number
func hasExplicitNumber(opts MoveOptions) bool {
	return opts.Mode == "force" || (opts.TargetNumber != nil && *opts.TargetNumber > 0)
}
