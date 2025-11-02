package cmd

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new tk database",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := database.GetDBPath()
		if err != nil {
			return err
		}

		if database.DBExists(path) {
			return fmt.Errorf("database already exists at %s", path)
		}

		db, err := database.OpenDB(path)
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.InitDB(); err != nil {
			return err
		}

		fmt.Printf("Database initialized at %s\n", path)
		return nil
	},
}
