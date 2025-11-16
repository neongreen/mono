package cmd

import (
	"encoding/json"
	"fmt"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/spf13/cobra"
)

var remoteLsCmd = &cobra.Command{
	Use:   "remote-ls",
	Short: "List configured remotes",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		if len(config.Remotes) == 0 {
			if jsonOutput {
				fmt.Println("[]")
			} else {
				fmt.Println("No remotes configured.")
			}
			return nil
		}

		if jsonOutput {
			output, err := json.MarshalIndent(config.Remotes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal remotes: %w", err)
			}
			fmt.Println(string(output))
		} else {
			for name, remote := range config.Remotes {
				fmt.Printf("%s:\n", name)
				fmt.Printf("  Type: %s\n", remote.Type)
				fmt.Printf("  Path: %s\n", remote.Path)
				fmt.Printf("  Spaces: %v\n", remote.Spaces)
				fmt.Printf("  Push: %v\n", remote.Push)
				fmt.Printf("  Pull: %v\n", remote.Pull)
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	remoteLsCmd.Flags().Bool("json", false, "Output as JSON")
}
