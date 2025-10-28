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

type moveOptions struct {
	Mode         string
	ForceNumber  int64
	OnCollision  string
	TargetNumber *int64
}

var mvCmd = &cobra.Command{
	Use:   "mv <task> <target>",
	Short: "Move a task to another project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		targetSpec := args[1]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		opts, err := parseMoveOptions(cmd, targetSpec)
		if err != nil {
			return err
		}

		return moveTask(db, taskRef, targetSpec, opts)
	},
}

func parseMoveOptions(cmd *cobra.Command, targetSpec string) (moveOptions, error) {
	keep, _ := cmd.Flags().GetBool("keep")
	auto, _ := cmd.Flags().GetBool("auto")
	forceFlag, _ := cmd.Flags().GetInt64("force")
	onCollision, _ := cmd.Flags().GetString("on-collision")

	if onCollision != "fail" && onCollision != "auto" {
		return moveOptions{}, fmt.Errorf("invalid value for --on-collision: %s (expected fail or auto)", onCollision)
	}

	// Parse inline target number (alias:number syntax)
	var inlineNumber *int64
	if parts := strings.SplitN(targetSpec, ":", 2); len(parts) == 2 {
		n, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || n <= 0 {
			return moveOptions{}, fmt.Errorf("invalid number %q in target %s", parts[1], targetSpec)
		}
		inlineNumber = &n
	}

	modeCount := 0
	if keep {
		modeCount++
	}
	if auto {
		modeCount++
	}
	if cmd.Flags().Changed("force") || inlineNumber != nil {
		modeCount++
	}
	if modeCount > 1 {
		return moveOptions{}, fmt.Errorf("use only one of --keep, --auto, or --force")
	}

	opts := moveOptions{
		Mode:         "keep",
		OnCollision:  onCollision,
		TargetNumber: inlineNumber,
	}

	if auto {
		opts.Mode = "auto"
	}

	if keep {
		opts.Mode = "keep"
	}

	if inlineNumber != nil {
		opts.Mode = "force"
		opts.ForceNumber = *inlineNumber
	}

	if cmd.Flags().Changed("force") {
		if forceFlag <= 0 {
			return moveOptions{}, fmt.Errorf("--force requires a positive integer")
		}
		opts.Mode = "force"
		opts.ForceNumber = forceFlag
	}

	return opts, nil
}

func moveTask(db *DB, taskRef string, targetSpec string, opts moveOptions) error {
	taskUID, err := ResolveTaskReference(db, taskRef)
	if err != nil {
		return err
	}

	fromProjectUID, oldNumber, err := taskProjectAndNumber(db, taskUID)
	if err != nil {
		return err
	}

	targetRef := targetSpec
	if idx := strings.IndexRune(targetSpec, ':'); idx != -1 {
		targetRef = targetSpec[:idx]
	}

	toProjectUID, err := resolveProjectByAlias(db, targetRef)
	if err != nil {
		return err
	}

	if fromProjectUID == toProjectUID && opts.Mode == "keep" && !cmdExplicitNumber(opts) {
		return fmt.Errorf("task already in project %s", targetRef)
	}

	currentUser, err := getCurrentUser()
	if err != nil {
		return err
	}

	lamport, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	numberPolicy := NumberPolicyPayload{Mode: opts.Mode}
	switch opts.Mode {
	case "keep":
		numberPolicy.Number = oldNumber
		if opts.OnCollision == "auto" {
			collision, err := numberCollisionExists(db, toProjectUID, oldNumber, taskUID)
			if err != nil {
				return err
			}
			if collision {
				numberPolicy.Mode = "auto"
				numberPolicy.Number = 0
			}
		} else {
			collision, err := numberCollisionExists(db, toProjectUID, oldNumber, taskUID)
			if err != nil {
				return err
			}
			if collision {
				display, _ := RenderTaskDisplayID(db, taskUID)
				return fmt.Errorf("task %s would collide with existing number %d in target project; rerun with --auto or --force", display, oldNumber)
			}
		}
	case "auto":
		numberPolicy.Number = 0
	case "force":
		numberPolicy.Number = opts.ForceNumber
	default:
		return fmt.Errorf("unsupported number policy mode %s", opts.Mode)
	}

	payload := TaskRelocatePayload{
		TaskUID:        taskUID,
		FromProjectUID: fromProjectUID,
		ToProjectUID:   toProjectUID,
		NumberPolicy:   numberPolicy,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.relocate payload: %w", err)
	}

	event := Event{
		ID:        string(NewEventID()),
		TS:        lamport,
		CreatedAt: time.Now(),
		Actor:     currentUser,
		Role:      "human",
		Kind:      string(EventKindTaskRelocate),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert task.relocate event: %w", err)
	}

	if err := db.ProjectTaskRelocateEvent(event); err != nil {
		return fmt.Errorf("failed to project task.relocate event: %w", err)
	}

	display, err := RenderTaskDisplayID(db, taskUID)
	if err != nil {
		return err
	}

	fmt.Printf("Moved task %s to %s\n", taskRef, display)
	return nil
}

func cmdExplicitNumber(opts moveOptions) bool {
	return opts.Mode == "force" || (opts.TargetNumber != nil && *opts.TargetNumber > 0)
}

func taskProjectAndNumber(db *DB, taskUID string) (string, int64, error) {
	var projectUID string
	if err := db.db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, taskUID).Scan(&projectUID); err != nil {
		if err == sql.ErrNoRows {
			return "", 0, fmt.Errorf("task %s not found", taskUID)
		}
		return "", 0, fmt.Errorf("failed to lookup task %s: %w", taskUID, err)
	}

	var number int64
	if err := db.db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&number); err != nil {
		if err == sql.ErrNoRows {
			number = 0
		} else {
			return "", 0, fmt.Errorf("failed to lookup task number: %w", err)
		}
	}

	return projectUID, number, nil
}

func numberCollisionExists(db *DB, projectUID string, number int64, taskUID string) (bool, error) {
	var count int
	if err := db.db.QueryRow(`
		SELECT COUNT(*) FROM task_numbers
		WHERE project_uid = ? AND number = ? AND task_uid != ?
	`, projectUID, number, taskUID).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check collisions: %w", err)
	}
	return count > 0, nil
}

func init() {
	mvCmd.Flags().Bool("keep", false, "Keep the existing task number in the new project")
	mvCmd.Flags().Bool("auto", false, "Auto-assign the next available number in the new project")
	mvCmd.Flags().Int64("force", 0, "Force a specific number in the new project")
	mvCmd.Flags().String("on-collision", "fail", "Collision handling strategy (fail|auto)")
}
