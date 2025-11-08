package cmd

import (
	"fmt"

	"github.com/neongreen/mono/lib/cli"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show what you have",
	Long: `Show all tools, repositories, and resources that have been installed via want.

This command will display:
  • Tools installed via want
  • Repositories cloned via want
  • Their current status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(cli.Muted("MVP: No requirements tracked yet"))
		fmt.Println("\nThis command will show:")
		fmt.Println("  • Tools installed via want")
		fmt.Println("  • Repositories cloned via want")
		fmt.Println("  • Their current status")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
