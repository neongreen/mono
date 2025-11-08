package node

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current node ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := database.OpenExistingDB()
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

func init() {
	ShowCmd.Flags().Bool("json", false, "Output as JSON")
}
