package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var groupShowCmd = &cobra.Command{
	Use:   "group-show",
	Short: "Show group details",
	Long: `Display detailed information about a group.

Example:
  tk group show q-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		groupID := args[0]

		// Query group details
		var primitive string
		var kind string
		var name string
		var metadata sql.NullString
		var removed int
		err = db.Db.QueryRow(`
			SELECT primitive, kind, name, metadata, removed
			FROM containers
			WHERE id = ?
		`, groupID).Scan(&primitive, &kind, &name, &metadata, &removed)
		if err != nil {
			return fmt.Errorf("group %q not found", groupID)
		}

		if primitive != string(types.PrimitiveGroup) {
			return fmt.Errorf("%q is a %s, not a group", groupID, primitive)
		}

		// Count members
		var memberCount int
		err = db.Db.QueryRow(`
			SELECT COUNT(*)
			FROM container_members
			WHERE container_id = ? AND removed = 0
		`, groupID).Scan(&memberCount)
		if err != nil {
			return fmt.Errorf("failed to count members: %w", err)
		}

		if jsonOutput {
			type GroupInfo struct {
				ID          string         `json:"id"`
				Name        string         `json:"name"`
				Kind        string         `json:"kind"`
				Primitive   string         `json:"primitive"`
				MemberCount int            `json:"member_count"`
				Removed     bool           `json:"removed"`
				Metadata    map[string]any `json:"metadata,omitempty"`
			}

			info := GroupInfo{
				ID:          groupID,
				Name:        name,
				Kind:        kind,
				Primitive:   primitive,
				MemberCount: memberCount,
				Removed:     removed == 1,
			}

			if metadata.Valid && metadata.String != "" && metadata.String != "null" {
				var metaMap map[string]any
				if err := json.Unmarshal([]byte(metadata.String), &metaMap); err == nil {
					info.Metadata = metaMap
				}
			}

			output, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(output))
		} else {
			// Display
			fmt.Printf("Group: %s\n", groupID)
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
		}

		return nil
	},
}

func init() {
	groupShowCmd.Flags().Bool("json", false, "Output as JSON")
}
