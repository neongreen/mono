package cmd

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/tk/internal/utils"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new [title]",
	Aliases: []string{"add"},
	Short:   "Create a new task",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		projectFlag, _ := cmd.Flags().GetString("project")
		title := args[0]
		
		// Auto-detect project from "project: title" format if -p not specified
		if projectFlag == "tk" { // Default project
			if idx := strings.Index(title, ": "); idx > 0 {
				prefix := title[:idx]
				restOfTitle := title[idx+2:]
				
				// Check if prefix is a valid project
				if _, err := database.ResolveProjectRef(db, types.NewProjectRef(prefix)); err == nil {
					projectFlag = prefix
					title = restOfTitle
				}
			}
		}

		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		result, err := database.CreateTask(db, database.CreateTaskParams{
			ProjectRef:  projectFlag,
			Title:       title,
			CurrentUser: currentUser,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Created task %s: %s\n", result.DisplayID, args[0])
		return nil
	},
}
