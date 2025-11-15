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

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var exemptCmd = &cobra.Command{
	Use:   "exempt",
	Short: "A command exempt from json requirement",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var exemptInteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "An interactive command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// This comment doesn't have the directive
var stillBadCmd = &cobra.Command{ // want `command "stillBadCmd" \(use: "stillbad"\) missing required --json flag`
	Use:   "stillbad",
	Short: "A command without the directive",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
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
	RootCmd.AddCommand(exemptCmd)
	RootCmd.AddCommand(exemptInteractiveCmd)
	RootCmd.AddCommand(stillBadCmd)
}
