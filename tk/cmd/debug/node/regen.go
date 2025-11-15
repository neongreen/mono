package node

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var RegenCmd = &cobra.Command{
	Use:   "regen",
	Short: "Regenerate the node ID (use with caution)",
	Long: `Regenerates the node ID for this installation.

WARNING: This will change the node ID, which means:
- New tasks will have different IDs (tk-N-<new-node>)
- New events will have different IDs (ev-N-<new-node>)
- Old task IDs (tk-N-<old-node>) will remain unchanged
- You should record a node.alias event to link the old and new node IDs

Only use this command if you have a node ID collision with another machine.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Get current node ID
		oldNodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		fmt.Printf("Current node ID: %s\n", oldNodeID)
		fmt.Println("\nWARNING: This will change your node ID.")
		fmt.Println("- Existing task IDs will remain unchanged")
		fmt.Println("- New tasks will use the new node ID")
		fmt.Println("- You should sync before and after this operation")
		fmt.Print("\nContinue? (yes/no): ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}

		// Generate new node ID
		newNodeID, err := db.RegenerateNodeID()
		if err != nil {
			return err
		}

		fmt.Printf("\nNew node ID: %s\n", newNodeID)
		fmt.Println("\nNote: Consider recording this change by creating a node.alias event")
		fmt.Println("to link the old and new node IDs for UI purposes.")

		return nil
	},
}
