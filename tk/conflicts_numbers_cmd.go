package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var conflictsNumbersCmd = &cobra.Command{
	Use:   "numbers",
	Short: "List task number collisions",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFlag, _ := cmd.Flags().GetString("project")

		db, err := openExistingDB(false)
		if err != nil {
			return err
		}
		defer db.Close()

		var projectUID string
		if projectFlag != "" {
			resolved, err := resolveProjectByAlias(db, projectFlag)
			if err != nil {
				return err
			}
			projectUID = resolved
		}

		collisions, err := getNumberCollisions(db, projectUID)
		if err != nil {
			return err
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
	conflictsNumbersCmd.Flags().String("project", "", "Project alias or UID to inspect")
	conflictsCmd.AddCommand(conflictsNumbersCmd)
}
