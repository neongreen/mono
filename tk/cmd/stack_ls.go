package cmd

import (
	"database/sql"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var stackLsCmd = &cobra.Command{
	Use:   "stack-ls",
	Short: "List stacks or stack members",
	Long: `List all stacks, or list members of a specific stack.

Examples:
  tk stack list       # List all stacks
  tk stack list s-1   # List members of stack s-1`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check database version
		version, err := db.GetDBVersion()
		if err != nil {
			return err
		}
		if version < 6 {
			return fmt.Errorf("containers require database v6 or higher, but current version is v%d", version)
		}

		// If stack ID provided, list members
		if len(args) == 1 {
			return listStackMembers(db, args[0])
		}

		// Otherwise list all stacks
		rows, err := db.Db.Query(`
			SELECT id, kind, name, metadata
			FROM containers
			WHERE primitive = ? AND removed = 0
			ORDER BY id
		`, types.PrimitiveStack)
		if err != nil {
			return fmt.Errorf("failed to query stacks: %w", err)
		}
		defer rows.Close()

		// Print header
		fmt.Printf("%-8s %-12s %-30s\n", "ID", "KIND", "NAME")
		fmt.Println("────────────────────────────────────────────────────────────────")

		count := 0
		for rows.Next() {
			var id string
			var kind string
			var name string
			var metadata sql.NullString

			if err := rows.Scan(&id, &kind, &name, &metadata); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}

			fmt.Printf("%-8s %-12s %-30s\n", id, kind, name)
			count++
		}

		if count == 0 {
			fmt.Println("No stacks found.")
			fmt.Println("\nCreate one with: tk stack create <kind> <name>")
		}

		return nil
	},
}

func init() {
	stackLsCmd.Flags().Bool("json", false, "Output as JSON")
}

func listStackMembers(db *database.DB, stackID string) error {
	// Verify stack exists
	var primitive string
	var stackName string
	var removed int
	err := db.Db.QueryRow(`
		SELECT primitive, name, removed
		FROM containers
		WHERE id = ?
	`, stackID).Scan(&primitive, &stackName, &removed)
	if err != nil {
		return fmt.Errorf("stack %q not found", stackID)
	}

	if primitive != string(types.PrimitiveStack) {
		return fmt.Errorf("%q is a %s, not a stack", stackID, primitive)
	}

	if removed == 1 {
		return fmt.Errorf("stack %q has been removed", stackID)
	}

	// Query members (show in stack order - highest position first = top of stack)
	rows, err := db.Db.Query(`
		SELECT item_id, position
		FROM container_members
		WHERE container_id = ? AND removed = 0
		ORDER BY position DESC
	`, stackID)
	if err != nil {
		return fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	// Print header
	fmt.Printf("Stack %s: %s\n", stackID, stackName)
	fmt.Println()
	fmt.Printf("%-4s %s\n", "POS", "ITEM")
	fmt.Println("─────────────────────")

	count := 0
	for rows.Next() {
		var itemID string
		var position int64

		if err := rows.Scan(&itemID, &position); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Render task UID as display ID (tk-123) for user output
		displayID, err := database.RenderTaskDisplayID(db, itemID)
		if err != nil {
			// If rendering fails, show the raw UID (defensive)
			displayID = itemID
		}

		fmt.Printf("%-4d %s\n", position, displayID)
		count++
	}

	if count == 0 {
		fmt.Println("(empty)")
	} else {
		fmt.Printf("\n%d items in stack (top at position %d)\n", count, count)
	}

	return nil
}
