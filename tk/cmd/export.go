package cmd

import (
	"fmt"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/remote"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [remote-name]",
	Short: "Export local events to segment files",
	Long: `Export unsent local events to segment files.

Examples:
  tk export icloud               # Export to icloud remote
  tk export icloud --space personal --all  # Export all events
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space, _ := cmd.Flags().GetString("space")
		exportAll, _ := cmd.Flags().GetBool("all")

		// If no remote name provided, use default from config
		var remoteName string
		if len(args) > 0 {
			remoteName = args[0]
		}

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// If no remote name and no default, error
		if remoteName == "" {
			// Try to find a default remote
			if len(config.Remotes) == 0 {
				return fmt.Errorf("no remotes configured; add one with 'tk remote add'")
			}
			if len(config.Remotes) == 1 {
				// Use the only remote
				for name := range config.Remotes {
					remoteName = name
				}
			} else {
				return fmt.Errorf("multiple remotes configured; please specify which one to use")
			}
		}

		// Get remote config
		remoteConfig, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found", remoteName)
		}

		// Get state directory
		stateDir, err := config_pkg.GetStateDir()
		if err != nil {
			return err
		}

		// Export using business logic
		result, err := remote.Export(db, remote.ExportParams{
			RemoteName:   remoteName,
			RemoteConfig: remoteConfig,
			Space:        space,
			ExportAll:    exportAll,
			StateDir:     stateDir,
			SyncConfig:   config.Sync,
		})
		if err != nil {
			return err
		}

		if result.EventsExported == 0 {
			fmt.Println("No new events to export")
		} else {
			// Print segments written (for user visibility)
			for _, segInfo := range result.SegmentFiles {
				fmt.Printf("Wrote segment: %s\n", segInfo.Rel)
			}
			fmt.Printf("Exported %d events in %d segments\n", result.EventsExported, result.SegmentsWritten)
		}
		return nil
	},
}

func init() {
	exportCmd.Flags().String("space", "personal", "Space to export")
	exportCmd.Flags().Bool("all", false, "Export all events (not just unsent)")
}
