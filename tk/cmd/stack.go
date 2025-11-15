package cmd

import (
	stack_pkg "github.com/neongreen/mono/tk/cmd/stack"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Parent command; JSON only required for read-only data commands
var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Manage stacks (LIFO containers)",
	Long:  `Create and manage stack containers (last in, first out).`,
}

func init() {
	stackCmd.AddCommand(stack_pkg.CreateCmd)
	stackCmd.AddCommand(stack_pkg.ListCmd)
	stackCmd.AddCommand(stack_pkg.ShowCmd)
	stackCmd.AddCommand(stack_pkg.PushCmd)
	stackCmd.AddCommand(stack_pkg.PopCmd)
	stackCmd.AddCommand(stack_pkg.RenameCmd)
	stackCmd.AddCommand(stack_pkg.RmCmd)
}
