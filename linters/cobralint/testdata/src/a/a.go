package testcmd

import (
	"a/cobra"
)

var goodCmd = &cobra.Command{
	Use:   "good",
	Short: "A command with json flag",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var badCmd = &cobra.Command{ // want `command "badCmd" \(use: "bad"\) missing required --json flag`
	Use:   "bad",
	Short: "A command without json flag",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var RootCmd = &cobra.Command{
	Use:   "test",
	Short: "Root command",
}

func init() {
	goodCmd.Flags().Bool("json", false, "Output as JSON")
	RootCmd.AddCommand(goodCmd)
	RootCmd.AddCommand(badCmd)
}
