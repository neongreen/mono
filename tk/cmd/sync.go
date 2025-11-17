package cmd

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/collision"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/remote"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var pushCmd = &cobra.Command{
	Use:   "push [remote-name]",
	Short: "Push local segments to remote",
	Long: `Push local segment files to a remote.

Examples:
  tk push icloud
  tk push icloud --space personal
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space, _ := cmd.Flags().GetString("space")
		exportAll, _ := cmd.Flags().GetBool("all")

		// Get remote name
		var remoteName string
		if len(args) > 0 {
			remoteName = args[0]
		}

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// If no remote name provided, try to find default
		if remoteName == "" {
			if len(config.Remotes) == 0 {
				return fmt.Errorf("no remotes configured")
			}
			if len(config.Remotes) == 1 {
				for name := range config.Remotes {
					remoteName = name
				}
			} else {
				return fmt.Errorf("multiple remotes configured; please specify which one")
			}
		}

		// Get remote config
		remoteConfig, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found", remoteName)
		}

		stateDir, err := config_pkg.GetStateDir()
		if err != nil {
			return err
		}

		if !remoteConfig.Push {
			return fmt.Errorf("push is disabled for remote '%s'", remoteName)
		}

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Push using business logic
		result, err := remote.Push(db, remote.PushParams{
			RemoteName:   remoteName,
			RemoteConfig: remoteConfig,
			Space:        space,
			StateDir:     stateDir,
			SyncConfig:   config.Sync,
			ExportAll:    exportAll,
		})
		if err != nil {
			return err
		}

		if result.SegmentsPushed == 0 {
			fmt.Println("No new segments to push")
		} else {
			fmt.Printf("Pushed %d segments, index updated\n", result.SegmentsPushed)
		}
		return nil
	},
}

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var pullCmd = &cobra.Command{
	Use:   "pull [remote-name]",
	Short: "Pull segments from remote",
	Long: `Pull segment files from a remote.

Examples:
  tk pull icloud
  tk pull icloud --space personal
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space, _ := cmd.Flags().GetString("space")

		// Get remote name
		var remoteName string
		if len(args) > 0 {
			remoteName = args[0]
		}

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// If no remote name provided, try to find default
		if remoteName == "" {
			if len(config.Remotes) == 0 {
				return fmt.Errorf("no remotes configured")
			}
			if len(config.Remotes) == 1 {
				for name := range config.Remotes {
					remoteName = name
				}
			} else {
				return fmt.Errorf("multiple remotes configured; please specify which one")
			}
		}

		// Get remote config
		remoteConfig, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found", remoteName)
		}

		if !remoteConfig.Pull {
			return fmt.Errorf("pull is disabled for remote '%s'", remoteName)
		}

		// Get state directory
		stateDir, err := config_pkg.GetStateDir()
		if err != nil {
			return err
		}

		// Pull using business logic
		result, err := remote.Pull(remoteName, remoteConfig, space, stateDir)
		if err != nil {
			return err
		}

		if result.SegmentsPulled == 0 {
			fmt.Println("No segments found on remote")
		} else {
			fmt.Printf("Pulled %d segments\n", result.SegmentsPulled)
		}
		return nil
	},
}

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var syncCmd = &cobra.Command{
	Use:   "sync [remote-name]",
	Short: "Sync with remote (pull → ingest → export → push)",
	Long: `Sync local events with a remote.

This command performs the full sync flow:
1. Pull segments from remote
2. Ingest events from remote segments
3. Push local events to remote

Examples:
  tk sync icloud
  tk sync icloud --space personal
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space, _ := cmd.Flags().GetString("space")

		// Get remote name
		var remoteName string
		if len(args) > 0 {
			remoteName = args[0]
		}

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// If no remote name provided, try to find default
		if remoteName == "" {
			if len(config.Remotes) == 0 {
				return fmt.Errorf("no remotes configured")
			}
			if len(config.Remotes) == 1 {
				for name := range config.Remotes {
					remoteName = name
				}
			} else {
				return fmt.Errorf("multiple remotes configured; please specify which one")
			}
		}

		// Get remote config
		remoteConfig, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found", remoteName)
		}

		stateDir, err := config_pkg.GetStateDir()
		if err != nil {
			return err
		}

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check for node collisions first
		fmt.Println("Checking for node collisions...")
		if err := collision.CheckNodeCollision(db, remoteName, remoteConfig); err != nil {
			return fmt.Errorf("node collision detected: %w", err)
		}

		// 1. Pull
		fmt.Println("Pulling from remote...")
		if remoteConfig.Pull {
			pullCmd.Flags().Set("space", space)
			if err := pullCmd.RunE(pullCmd, []string{remoteName}); err != nil {
				return fmt.Errorf("pull failed: %w", err)
			}
		}

		// 2. Ingest
		fmt.Println("Ingesting events...")
		if err := IngestRemote(db, remoteName, remoteConfig); err != nil {
			return fmt.Errorf("ingest failed: %w", err)
		}

		// 3. Push
		fmt.Println("Pushing to remote...")
		if remoteConfig.Push {
			_, err := remote.Push(db, remote.PushParams{
				RemoteName:   remoteName,
				RemoteConfig: remoteConfig,
				Space:        space,
				StateDir:     stateDir,
				SyncConfig:   config.Sync,
			})
			if err != nil {
				return fmt.Errorf("push failed: %w", err)
			}
		}

		fmt.Println("Sync complete")
		return nil
	},
}

func init() {
	pushCmd.Flags().String("space", "personal", "Space to push")
	pushCmd.Flags().Bool("all", false, "Export all events before pushing (rebuild segments)")
	pullCmd.Flags().String("space", "personal", "Space to pull")
	syncCmd.Flags().String("space", "personal", "Space to sync")
}
