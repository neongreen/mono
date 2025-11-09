package cmd

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <task> <field> <value>",
	Short: "Edit task fields (number, title, status)",
	Long: `Edit task fields (number, title, status).

For editing just the title, you can also use 'tk describe <task> <new-title>' as a shortcut.`,
	Args: cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		field := strings.ToLower(args[1])
		value := strings.Join(args[2:], " ")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve task reference
		taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
		if err != nil {
			return err
		}

		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Edit the field using business logic
		if err := tasks.EditField(db, taskUID, field, value, currentUser, &clock.RealClock{}); err != nil {
			return err
		}

		// Display success message
		displayID, err := database.RenderTaskDisplayID(db, taskUID)
		if err != nil {
			displayID = taskUID
		}

		switch field {
		case "status":
			if value == "" {
				fmt.Printf("Unset status for %s\n", displayID)
			} else {
				fmt.Printf("Updated status for %s to %s\n", displayID, value)
			}
		case "number":
			fmt.Printf("Updated number for %s to %s\n", displayID, value)
		case "title":
			fmt.Printf("Updated title for %s\n", displayID)
		}

		return nil
	},
}
