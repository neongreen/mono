package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administrative commands for tk",
	Long:  `Administrative commands for tk database management and maintenance.`,
}

var rollbackV4Cmd = &cobra.Command{
	Use:   "rollback-v4",
	Short: "Rollback v4 migration and restore v3 backup",
	Long: `Rollback the v4 migration by restoring the v3 backup.

This command:
1. Restores tk.db.v3.bak as tk.db
2. Resets meta.version_major = 3
3. Allows you to use v1/v2 binaries again

The v4 segments in v4/ remain untouched and can be ignored.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := GetDBPath()
		if err != nil {
			return err
		}

		// Perform rollback
		fmt.Println("Rolling back v4 migration...")
		fmt.Printf("Restoring backup from %s%s\n", path, v4BackupSuffix)

		if err := RollbackV4(path); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		// Open the restored database and reset version
		db, err := OpenDB(path)
		if err != nil {
			return fmt.Errorf("failed to open restored database: %w", err)
		}
		defer db.Close()

		// Set version back to 3
		if err := db.SetDBVersion(v3SpecVersion); err != nil {
			return fmt.Errorf("failed to reset version: %w", err)
		}

		fmt.Println("Rollback complete!")
		fmt.Println("You can now use v1/v2 tk binaries with this database.")
		return nil
	},
}

func init() {
	adminCmd.AddCommand(rollbackV4Cmd)
}
