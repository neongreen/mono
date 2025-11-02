package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var forgetCmd = &cobra.Command{
	Use:   "forget <requirement>",
	Short: "Remove from tracking (doesn't uninstall)",
	Long: `Remove a requirement from tracking without uninstalling it.

This command removes the requirement from want's tracking database
but does not uninstall or delete the actual tool or repository.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		requirement := args[0]
		fmt.Printf("MVP: Would forget: %s\n", requirement)
		fmt.Println("This command will remove the requirement from tracking without uninstalling.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(forgetCmd)
}
