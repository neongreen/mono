package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/sync"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remotes",
}

var remoteAddCmd = &cobra.Command{
	Use:   "add [name] [type] [path]",
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

		if remoteType != "folder" {
			return fmt.Errorf("unsupported remote type: %s (only 'folder' is supported in v1)", remoteType)
		}

		// Expand home directory
		if strings.HasPrefix(path, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			path = filepath.Join(home, path[2:])
		}

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// Check if remote already exists
		if _, exists := config.Remotes[name]; exists {
			return fmt.Errorf("remote '%s' already exists", name)
		}

		// Add remote
		config.Remotes[name] = sync.RemoteConfig{
			Type:   remoteType,
			Path:   path,
			Spaces: []string{"personal"},
			Push:   true,
			Pull:   true,
		}

		// Save config
		if err := config_pkg.SaveConfig(config); err != nil {
			return err
		}

		fmt.Printf("Added remote '%s' (type: %s, path: %s)\n", name, remoteType, path)
		return nil
	},
}

var remoteLsCmd = &cobra.Command{
	Use:   "ls",
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

var remoteRmCmd = &cobra.Command{
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

func init() {
	remoteCmd.AddCommand(remoteAddCmd)

	remoteLsCmd.Flags().Bool("json", false, "Output as JSON")
	remoteCmd.AddCommand(remoteLsCmd)

	remoteCmd.AddCommand(remoteRmCmd)
}
