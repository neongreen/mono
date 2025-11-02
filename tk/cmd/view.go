package cmd

import (
	"encoding/json"
	"fmt"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view [task-id]",
	Short: "View task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		config, err := config_pkg.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return err
		}

		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
		if err != nil {
			return err
		}

		task, ok := reducer.GetTask(taskUUID)
		if !ok {
			return fmt.Errorf("task not found: %s", taskRef)
		}

		displayID, err := database.RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		taskCopy := *task
		taskCopy.TaskID = displayID

		_ = jsonOutput
		output, err := json.MarshalIndent(taskCopy, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal task: %w", err)
		}

		fmt.Println(string(output))
		return nil
	},
}
