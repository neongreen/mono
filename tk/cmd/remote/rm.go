package remote

import (
	"fmt"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/remote"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
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

		// Remove remote using business logic
		if err := remote.RemoveRemote(config, name); err != nil {
			return err
		}

		// Save config
		if err := config_pkg.SaveConfig(config); err != nil {
			return err
		}

		fmt.Printf("Removed remote '%s'\n", name)
		return nil
	},
}
