package remote

import (
	"fmt"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/spf13/cobra"
)

var RmCmd = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a remote",
	Long: `Remove a configured remote.

Examples:
  tk remote rm icloud
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// Check if remote exists
		if _, exists := config.Remotes[name]; !exists {
			return fmt.Errorf("remote '%s' not found", name)
		}

		// Remove remote
		delete(config.Remotes, name)

		// Save config
		if err := config_pkg.SaveConfig(config); err != nil {
			return err
		}

		fmt.Printf("Removed remote '%s'\n", name)
		return nil
	},
}
