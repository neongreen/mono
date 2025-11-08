package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/neongreen/mono/tk/internal/utils"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv <task> <target>",
	Short: "Move a task to another project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		targetSpec := args[1]

		db, err := database.OpenExistingDB()
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

func parseMoveOptions(cmd *cobra.Command, targetSpec string) (tasks.MoveOptions, error) {
	keep, _ := cmd.Flags().GetBool("keep")
	auto, _ := cmd.Flags().GetBool("auto")
	forceFlag, _ := cmd.Flags().GetInt64("force")
	onCollision, _ := cmd.Flags().GetString("on-collision")

	if onCollision != "fail" && onCollision != "auto" {
		return tasks.MoveOptions{}, fmt.Errorf("invalid value for --on-collision: %s (expected fail or auto)", onCollision)
	}

	// Parse inline target number (alias:number syntax)
	var inlineNumber *int64
	if parts := strings.SplitN(targetSpec, ":", 2); len(parts) == 2 {
		n, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || n <= 0 {
			return tasks.MoveOptions{}, fmt.Errorf("invalid number %q in target %s", parts[1], targetSpec)
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
		return tasks.MoveOptions{}, fmt.Errorf("use only one of --keep, --auto, or --force")
	}

	opts := tasks.MoveOptions{
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
			return tasks.MoveOptions{}, fmt.Errorf("--force requires a positive integer")
		}
		opts.Mode = "force"
		opts.ForceNumber = forceFlag
	}

	return opts, nil
}

func moveTask(db *database.DB, taskRef string, targetSpec string, opts tasks.MoveOptions) error {
	// Resolve task reference
	taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
	if err != nil {
		return err
	}

	// Parse target project (strip :number suffix if present)
	targetRef := targetSpec
	if idx := strings.IndexRune(targetSpec, ':'); idx != -1 {
		targetRef = targetSpec[:idx]
	}

	toProjectUID, err := database.ResolveProjectByAlias(db, targetRef)
	if err != nil {
		return err
	}

	currentUser, err := utils.GetCurrentUser()
	if err != nil {
		return err
	}

	// Move the task using business logic
	if err := tasks.Move(db, taskUID, toProjectUID, opts, currentUser); err != nil {
		return err
	}

	// Display success message
	display, err := database.RenderTaskDisplayID(db, taskUID)
	if err != nil {
		return err
	}

	fmt.Printf("Moved task %s to %s\n", taskRef, display)
	return nil
}

func init() {
	mvCmd.Flags().Bool("keep", false, "Keep the existing task number in the new project")
	mvCmd.Flags().Bool("auto", false, "Auto-assign the next available number in the new project")
	mvCmd.Flags().Int64("force", 0, "Force a specific number in the new project")
	mvCmd.Flags().String("on-collision", "fail", "Collision handling strategy (fail|auto)")
}
