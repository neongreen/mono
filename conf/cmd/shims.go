package cmd

import (
	"fmt"
	"os"

	shimstool "github.com/neongreen/mono/conf/pkg/tools/shims"
	"github.com/spf13/cobra"
)

var shimsCmd = &cobra.Command{
	Use:   "shims",
	Short: "Manage command shims",
	Long: `Create and manage executable command shims in ~/.local/bin/conf-shims/.
	
Shims are executable scripts that act as aliases for longer commands.
They work across all shells (bash, zsh, fish) without modifying shell config files.

Add ~/.local/bin/conf-shims to your PATH to use the shims.`,
}

var shimsCreateCmd = &cobra.Command{
	Use:   "create [name] [command]",
	Short: "Create a new command shim",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		command := args[1]

		shimsTool, err := shimstool.NewShimsToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = shimsTool.CreateShim(name, command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Created shim: %s -> %s\n", name, command)
	},
}

var shimsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed command shims",
	Run: func(cmd *cobra.Command, args []string) {
		shimsTool, err := shimstool.NewShimsTool()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		shims, err := shimsTool.ListShims()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(shims) == 0 {
			fmt.Println("No shims found.")
			return
		}

		for _, shim := range shims {
			fmt.Printf("%-12s -> %s\n", shim.Name, shim.Command)
		}
	},
}

var shimsRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an existing command shim",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		shimsTool, err := shimstool.NewShimsToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = shimsTool.RemoveShim(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Removed shim: %s\n", name)
	},
}

func init() {
	shimsCmd.AddCommand(shimsCreateCmd)
	shimsCmd.AddCommand(shimsRemoveCmd)
	shimsCmd.AddCommand(shimsListCmd)
}
