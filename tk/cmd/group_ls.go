package cmd

import (
	"database/sql"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var groupLsCmd = &cobra.Command{
	Use:   "group-ls",
	Short: "List groups or group members",
	Long: `List all groups, or list members of a specific group.

Examples:
  tk group list       # List all groups
  tk group list g-1   # List members of group g-1`,
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

		// If group ID provided, list members
		if len(args) == 1 {
			return listGroupMembers(db, args[0])
		}

		// Otherwise list all groups
		rows, err := db.Db.Query(`
			SELECT id, kind, name, metadata
			FROM containers
			WHERE primitive = ? AND removed = 0
			ORDER BY id
		`, types.PrimitiveGroup)
		if err != nil {
			return fmt.Errorf("failed to query groups: %w", err)
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
			fmt.Println("No groups found.")
			fmt.Println("\nCreate one with: tk group create <kind> <name>")
		}

		return nil
	},
}

func listGroupMembers(db *database.DB, groupID string) error {
	// Verify group exists
	var primitive string
	var groupName string
	var removed int
	err := db.Db.QueryRow(`
		SELECT primitive, name, removed
		FROM containers
		WHERE id = ?
	`, groupID).Scan(&primitive, &groupName, &removed)
	if err != nil {
		return fmt.Errorf("group %q not found", groupID)
	}

	if primitive != string(types.PrimitiveGroup) {
		return fmt.Errorf("%q is a %s, not a group", groupID, primitive)
	}

	if removed == 1 {
		return fmt.Errorf("group %q has been removed", groupID)
	}

	// Query members (unordered - groups have no position)
	rows, err := db.Db.Query(`
		SELECT item_id
		FROM container_members
		WHERE container_id = ? AND removed = 0
		ORDER BY item_id
	`, groupID)
	if err != nil {
		return fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	// Print header
	fmt.Printf("Group %s: %s\n", groupID, groupName)
	fmt.Println()
	fmt.Println("ITEM")
	fmt.Println("─────────────────")

	count := 0
	for rows.Next() {
		var itemID string

		if err := rows.Scan(&itemID); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Render task UID as display ID (tk-123) for user output
		displayID, err := database.RenderTaskDisplayID(db, itemID)
		if err != nil {
			// If rendering fails, show the raw UID (defensive)
			displayID = itemID
		}

		fmt.Println(displayID)
		count++
	}

	if count == 0 {
		fmt.Println("(empty)")
	} else {
		fmt.Printf("\n%d items in group\n", count)
	}

	return nil
}
