package cmd

import (
	"fmt"
	"os"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/spf13/cobra"
)

var untrackCmd = &cobra.Command{
	Use:   "untrack <name>",
	Short: "Stop tracking a folder",
	Long: `Stop tracking a folder and optionally remove its copy from conf directory.

By default, this removes the folder copy and manifest. Use --keep-copy to preserve
the folder copy in ~/.config/conf/<name>/.

Examples:
  conf untrack my-docs              # Remove folder copy and manifest
  conf untrack my-docs --keep-copy  # Remove manifest but keep folder copy`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		folderName := args[0]
		keepCopy, _ := cmd.Flags().GetBool("keep-copy")

		// Load config
		conf, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		configDir, err := config.ConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}

		// Check if folder is tracked
		if _, exists := conf.Folders[folderName]; !exists {
			return fmt.Errorf("folder %s is not tracked", folderName)
		}

		fmt.Printf("Untracking folder: %s\n", folderName)

		// Remove folder copy unless --keep-copy is set
		if !keepCopy {
			folderCopyPath := config.FolderCopyPath(configDir, folderName)
			if _, err := os.Stat(folderCopyPath); err == nil {
				if err := os.RemoveAll(folderCopyPath); err != nil {
					return fmt.Errorf("failed to remove folder copy: %w", err)
				}
				fmt.Printf("  ✓ Removed folder copy at %s\n", folderCopyPath)
			}
		} else {
			fmt.Printf("  ✓ Keeping folder copy (--keep-copy specified)\n")
		}

		// Remove manifest file
		manifestPath := config.FolderManifestPath(configDir, folderName)
		if _, err := os.Stat(manifestPath); err == nil {
			if err := os.Remove(manifestPath); err != nil {
				return fmt.Errorf("failed to remove manifest: %w", err)
			}
			fmt.Printf("  ✓ Removed manifest at %s\n", manifestPath)
		}

		// Remove from main config
		conf.RemoveFolder(folderName)
		if err := conf.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("  ✓ Removed from config\n")

		fmt.Printf("\n✓ Folder %s is no longer tracked\n", folderName)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(untrackCmd)
	untrackCmd.Flags().Bool("keep-copy", false, "Keep the folder copy in conf directory")
}
