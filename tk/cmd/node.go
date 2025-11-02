package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage node ID",
}

var nodeShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current node ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		nodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		if jsonOutput {
			output := map[string]string{"node_id": nodeID}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal output: %w", err)
			}
			fmt.Println(string(data))
		} else {
			fmt.Println(nodeID)
		}
		return nil
	},
}

var nodeRegenCmd = &cobra.Command{
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
		db, err := OpenExistingDB()
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

func init() {
	nodeShowCmd.Flags().Bool("json", false, "Output as JSON")
	nodeCmd.AddCommand(nodeShowCmd)
	nodeCmd.AddCommand(nodeRegenCmd)
}
