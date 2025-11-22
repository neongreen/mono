package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/sync"
	"github.com/neongreen/mono/lib/cli"
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

		sort.Strings(toolsToSync)

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
		if dryRun {
			renderSyncDiff("Upload", config.FlattenValues(localValues), nil)
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
		if dryRun {
			renderSyncDiff("Download", nil, config.FlattenValues(icloudValues))
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

	if dryRun {
		renderSyncPreview(localValues, icloudValues, merged, localMtime, icloudMtime)
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

// renderSyncDiff renders a simple table for one-sided sync preview (upload/download)
func renderSyncDiff(action string, upload map[string]any, download map[string]any) {
	var rows []struct {
		action string
		path   string
		value  any
	}

	for path, val := range upload {
		rows = append(rows, struct {
			action string
			path   string
			value  any
		}{action, path, val})
	}
	for path, val := range download {
		rows = append(rows, struct {
			action string
			path   string
			value  any
		}{action, path, val})
	}

	if len(rows) == 0 {
		fmt.Println("  No values to sync")
		return
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].path < rows[j].path })

	fmt.Println()
	fmt.Printf("  Preview (%s %d values):\n", strings.ToLower(action), len(rows))
	printSyncTable(rows)
}

// renderSyncPreview shows merged differences for dry-run sync
func renderSyncPreview(localValues, icloudValues, merged map[string]any, localMtime, icloudMtime int64) {
	localFlat := config.FlattenValues(localValues)
	icloudFlat := config.FlattenValues(icloudValues)
	mergedFlat := config.FlattenValues(merged)

	type change struct {
		path       string
		source     string
		localValue any
		cloudValue any
		finalValue any
	}

	var rows []change
	seen := make(map[string]struct{})
	for k := range localFlat {
		seen[k] = struct{}{}
	}
	for k := range icloudFlat {
		seen[k] = struct{}{}
	}

	for path := range seen {
		lVal, lOk := localFlat[path]
		cVal, cOk := icloudFlat[path]
		final := mergedFlat[path]

		if lOk && cOk && fmt.Sprintf("%v", lVal) == fmt.Sprintf("%v", cVal) {
			continue // no difference
		}

		source := "local"
		if cOk && (!lOk || icloudMtime > localMtime) {
			source = "icloud"
		}

		rows = append(rows, change{
			path:       path,
			source:     source,
			localValue: lVal,
			cloudValue: cVal,
			finalValue: final,
		})
	}

	if len(rows) == 0 {
		fmt.Println("  No differences; configs already aligned")
		return
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].path < rows[j].path })

	fmt.Println()
	fmt.Printf("  Differences (%d):\n", len(rows))
	tableRows := make([]struct {
		action string
		path   string
		value  any
	}, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, struct {
			action string
			path   string
			value  any
		}{
			action: row.source,
			path:   row.path,
			value:  row.finalValue,
		})
	}
	printSyncTable(tableRows)
}

// printSyncTable renders a compact table for sync previews
func printSyncTable(rows []struct {
	action string
	path   string
	value  any
}) {
	t := cli.NewTable(os.Stdout)
	t.AppendHeader(table.Row{"Action", "Path", "Value"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: 10},
		{Number: 2, WidthMax: 42},
		{Number: 3, WidthMax: 60, WidthMaxEnforcer: text.WrapSoft},
	})

	for _, r := range rows {
		action := cli.Warning(r.action)
		if strings.EqualFold(r.action, "upload") || strings.EqualFold(r.action, "local") {
			action = cli.Success(r.action)
		} else if strings.EqualFold(r.action, "download") || strings.EqualFold(r.action, "icloud") {
			action = cli.Warning(r.action)
		}
		t.AppendRow(table.Row{
			action,
			cli.Key(r.path),
			cli.Value(formatValueShort(r.value)),
		})
	}

	t.Render()
}
