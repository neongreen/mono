package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [tool]",
	Short: "Sync configuration with iCloud Drive",
	Long: `Merge local configuration with iCloud Drive using simple Last-Write-Wins strategy.

This command downloads config from iCloud, merges with local config, and uploads
the merged result back to iCloud. Conflicts are resolved using Last-Write-Wins
based on file modification times.

iCloud Drive location:
  ~/Library/Mobile Documents/com~apple~CloudDocs/conf/

  Tool configs are stored as:
    - jj.toml (for jj config)
    - mise.toml (for mise config)
    - etc.

Examples:
  conf sync           # Sync all tools
  conf sync jj        # Sync only jj config
  conf sync --dry-run # Preview what would be synced`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to load conf config: %v\n", err)
			os.Exit(1)
		}

		configDir, err := config.ConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get config directory: %v\n", err)
			os.Exit(1)
		}

		var toolsToSync []string
		if len(args) == 1 {
			toolsToSync = []string{args[0]}
		} else {
			// Sync all tools
			for toolName := range conf.Tools {
				toolsToSync = append(toolsToSync, toolName)
			}
		}

		for _, toolName := range toolsToSync {
			if err := syncTool(conf, configDir, toolName, dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Error syncing %s: %v\n", toolName, err)
				os.Exit(1)
			}
		}

		if !dryRun {
			fmt.Println("\n✓ Sync complete")
		}
	},
}

// syncTool syncs a tool's config with iCloud Drive
func syncTool(conf *config.Config, configDir, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("Syncing %s configuration...\n", toolName)

	// Load sync metadata
	metadata, err := sync.LoadSyncMetadata(configDir)
	if err != nil {
		return fmt.Errorf("failed to load sync metadata: %w", err)
	}

	// Get local values
	localValues, err := sync.GetLocalValues(conf, toolName)
	if err != nil {
		return fmt.Errorf("failed to get local values: %w", err)
	}

	// Download from iCloud
	icloudValues, err := sync.DownloadFromICloud(toolName)
	if err != nil {
		return fmt.Errorf("failed to download from iCloud: %w", err)
	}

	if icloudValues == nil && len(localValues) == 0 {
		fmt.Printf("  %s: No configuration to sync\n", toolName)
		return nil
	}

	// Handle first-time sync cases
	if icloudValues == nil {
		// No iCloud data, upload local
		fmt.Printf("  No iCloud data found, uploading local config (%d values)\n", len(localValues))
		if !dryRun {
			if err := sync.UploadToICloud(toolName, localValues); err != nil {
				return fmt.Errorf("failed to upload to iCloud: %w", err)
			}

			// Update metadata
			perToolPath := filepath.Join(configDir, toolName+".toml")
			if localHash, err := sync.ComputeFileHash(perToolPath); err == nil {
				metadata.UpdateToolState(toolName, localHash, localHash)
				metadata.Save(configDir)
			}

			fmt.Printf("  ✓ Uploaded to iCloud\n")
		}
		return nil
	}

	if len(localValues) == 0 {
		// No local data, download from iCloud
		fmt.Printf("  No local data found, downloading from iCloud (%d values)\n", len(icloudValues))
		if !dryRun {
			// Set values in config
			for k, v := range icloudValues {
				conf.SetToolValue(toolName, k, v)
			}

			if err := conf.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			// Update metadata
			icloudPath, _ := sync.ICloudDrivePath()
			icloudFilePath := filepath.Join(icloudPath, toolName+".toml")
			if icloudHash, err := sync.ComputeFileHash(icloudFilePath); err == nil {
				metadata.UpdateToolState(toolName, icloudHash, icloudHash)
				metadata.Save(configDir)
			}

			fmt.Printf("  ✓ Downloaded from iCloud\n")
		}
		return nil
	}

	// Both have data - need to merge
	localCount := len(config.FlattenValues(localValues))
	icloudCount := len(config.FlattenValues(icloudValues))
	fmt.Printf("  Local: %d values, iCloud: %d values\n", localCount, icloudCount)

	// Get file modification times for LWW
	perToolPath := filepath.Join(configDir, toolName+".toml")
	icloudPath, _ := sync.ICloudDrivePath()
	icloudFilePath := filepath.Join(icloudPath, toolName+".toml")

	localStat, _ := os.Stat(perToolPath)
	icloudStat, _ := os.Stat(icloudFilePath)

	localMtime := localStat.ModTime().Unix()
	icloudMtime := icloudStat.ModTime().Unix()

	// Merge configs
	merged := sync.MergeConfigs(localValues, icloudValues, localMtime, icloudMtime)

	fmt.Printf("  Merged: %d values\n", len(merged))

	if dryRun {
		fmt.Printf("  Would upload merged config to iCloud\n")
		fmt.Printf("  Would update local config\n")
	} else {
		// Upload merged result to iCloud
		if err := sync.UploadToICloud(toolName, merged); err != nil {
			return fmt.Errorf("failed to upload merged config: %w", err)
		}

		// Update local config
		tool.Values = merged
		conf.Tools[toolName] = tool
		if err := conf.Save(); err != nil {
			return fmt.Errorf("failed to save local config: %w", err)
		}

		// Update metadata
		if icloudHash, err := sync.ComputeFileHash(icloudFilePath); err == nil {
			if localHash, err := sync.ComputeFileHash(perToolPath); err == nil {
				metadata.UpdateToolState(toolName, icloudHash, localHash)
				metadata.Save(configDir)
			}
		}

		fmt.Printf("  ✓ Synced with iCloud\n")
	}

	return nil
}
