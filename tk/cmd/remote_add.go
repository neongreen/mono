package cmd

import (
	"fmt"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/remote"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var remoteAddCmd = &cobra.Command{
	Use:   "remote-add",
	Short: "Add a new remote",
	Long: `Add a new remote for syncing events.

Examples:
  tk remote add icloud folder ~/Library/Mobile\ Documents/com~apple~CloudDocs/tk-events
`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		remoteType := args[1]
		path := args[2]

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// Add remote using business logic
		if err := remote.AddRemote(config, name, remoteType, path); err != nil {
			return err
		}

		// Save config
		if err := config_pkg.SaveConfig(config); err != nil {
			return err
		}

		fmt.Printf("Added remote '%s' (type: %s, path: %s)\n", name, remoteType, path)
		return nil
	},
}
