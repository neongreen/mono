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

var goodCmdWithVar = &cobra.Command{
	Use:   "goodvar",
	Short: "A command with json flag using BoolVar",
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
	// Regular Bool method
	goodCmd.Flags().Bool("json", false, "Output as JSON")

	// BoolVar method (first arg is destination, second is flag name)
	var jsonFlag bool
	goodCmdWithVar.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")

	RootCmd.AddCommand(goodCmd)
	RootCmd.AddCommand(goodCmdWithVar)
	RootCmd.AddCommand(badCmd)
}
