package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var stackShowCmd = &cobra.Command{
	Use:   "stack-show",
	Short: "Show stack details",
	Long: `Display detailed information about a stack.

Example:
  tk stack show q-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		stackID := args[0]

		// Query stack details
		var primitive string
		var kind string
		var name string
		var metadata sql.NullString
		var removed int
		err = db.Db.QueryRow(`
			SELECT primitive, kind, name, metadata, removed
			FROM containers
			WHERE id = ?
		`, stackID).Scan(&primitive, &kind, &name, &metadata, &removed)
		if err != nil {
			return fmt.Errorf("stack %q not found", stackID)
		}

		if primitive != string(types.PrimitiveStack) {
			return fmt.Errorf("%q is a %s, not a stack", stackID, primitive)
		}

		// Count members
		var memberCount int
		err = db.Db.QueryRow(`
			SELECT COUNT(*)
			FROM container_members
			WHERE container_id = ? AND removed = 0
		`, stackID).Scan(&memberCount)
		if err != nil {
			return fmt.Errorf("failed to count members: %w", err)
		}

		// Display
		fmt.Printf("Stack: %s\n", stackID)
		fmt.Printf("Name: %s\n", name)
		fmt.Printf("Kind: %s\n", kind)
		fmt.Printf("Members: %d\n", memberCount)

		if removed == 1 {
			fmt.Println("Status: REMOVED")
		}

		if metadata.Valid && metadata.String != "" && metadata.String != "null" {
			fmt.Println("\nMetadata:")
			// Pretty-print JSON metadata
			var metaMap map[string]any
			if err := json.Unmarshal([]byte(metadata.String), &metaMap); err == nil {
				for k, v := range metaMap {
					fmt.Printf("  %s: %v\n", k, v)
				}
			} else {
				fmt.Printf("  %s\n", metadata.String)
			}
		}

		return nil
	},
}

func init() {
	stackShowCmd.Flags().Bool("json", false, "Output as JSON")
}
