package cmd

import (
	"encoding/json"
	"fmt"

	debug_pkg "github.com/neongreen/mono/tk/cmd/debug"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var taskConflictsCmd = &cobra.Command{
	Use:   "task-conflicts",
	Short: "List task number collisions",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFlag, _ := cmd.Flags().GetString("project")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		var projectUID string
		if projectFlag != "" {
			resolved, err := database.ResolveProjectByAlias(db, projectFlag)
			if err != nil {
				return err
			}
			projectUID = resolved
		}

		collisions, err := debug_pkg.GetNumberCollisions(db, projectUID)
		if err != nil {
			return err
		}

		if jsonOutput {
			output, err := json.MarshalIndent(collisions, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal collisions: %w", err)
			}
			fmt.Println(string(output))
			return nil
		}

		if len(collisions) == 0 {
			if projectFlag != "" {
				fmt.Printf("No number collisions for project %s\n", projectFlag)
			} else {
				fmt.Println("No number collisions")
			}
			return nil
		}

		for _, collision := range collisions {
			alias := collision.ProjectAlias
			if alias == "" {
				alias = collision.ProjectUID
			}
			fmt.Printf("%s #%d\n", alias, collision.Number)
			for _, task := range collision.TaskDisplayIDs {
				fmt.Printf("  - %s\n", task)
			}
		}

		return nil
	},
}

func init() {
	taskConflictsCmd.Flags().String("project", "", "Project alias or UID to inspect")
	taskConflictsCmd.Flags().Bool("json", false, "Output as JSON")
}
