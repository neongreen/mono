package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/diff"
	"github.com/neongreen/mono/conf/pkg/sync"
	"github.com/neongreen/mono/conf/pkg/tools"
)

// importTool imports configuration values from a tool's target file into conf's state
func importTool(conf *config.Config, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("Importing %s configuration from %s...\n", toolName, tool.ConfigPath)

	values, err := getTargetConfigValues(toolName)
	if err != nil {
		return fmt.Errorf("failed to read target config: %w", err)
	}

	flatValues := config.FlattenValues(values)

	if len(flatValues) == 0 {
		fmt.Printf("  No values found in %s\n", tool.ConfigPath)
		return nil
	}

	fmt.Printf("  Found %d values\n", len(flatValues))

	if dryRun {

		for path, value := range flatValues {
			fmt.Printf("  Would import: %s.%s = %v\n", toolName, path, value)
		}
	} else {

		for path, value := range flatValues {
			fmt.Printf("  ✓ Imported %s.%s = %v\n", toolName, path, value)
		}

		conf.MergeToolValues(toolName, values)

		if err := conf.Save(); err != nil {
			return fmt.Errorf("failed to save conf state: %w", err)
		}

		fmt.Printf("  ✓ Saved to conf state\n")
	}

	return nil
}

// syncTool syncs a tool's config with iCloud Drive
func syncTool(conf *config.Config, configDir, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("Syncing %s configuration...\n", toolName)

	metadata, err := sync.LoadSyncMetadata(configDir)
	if err != nil {
		return fmt.Errorf("failed to load sync metadata: %w", err)
	}

	localValues, err := sync.GetLocalValues(conf, toolName)
	if err != nil {
		return fmt.Errorf("failed to get local values: %w", err)
	}

	icloudValues, err := sync.DownloadFromICloud(toolName)
	if err != nil {
		return fmt.Errorf("failed to download from iCloud: %w", err)
	}

	if icloudValues == nil && len(localValues) == 0 {
		fmt.Printf("  %s: No configuration to sync\n", toolName)
		return nil
	}

	if icloudValues == nil {

		fmt.Printf("  No iCloud data found, uploading local config (%d values)\n", len(localValues))
		if !dryRun {
			if err := sync.UploadToICloud(toolName, localValues); err != nil {
				return fmt.Errorf("failed to upload to iCloud: %w", err)
			}

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

		fmt.Printf("  No local data found, downloading from iCloud (%d values)\n", len(icloudValues))
		if !dryRun {

			for k, v := range icloudValues {
				conf.SetToolValue(toolName, k, v)
			}

			if err := conf.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

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

	localCount := len(config.FlattenValues(localValues))
	icloudCount := len(config.FlattenValues(icloudValues))
	fmt.Printf("  Local: %d values, iCloud: %d values\n", localCount, icloudCount)

	perToolPath := filepath.Join(configDir, toolName+".toml")
	icloudPath, _ := sync.ICloudDrivePath()
	icloudFilePath := filepath.Join(icloudPath, toolName+".toml")

	localStat, _ := os.Stat(perToolPath)
	icloudStat, _ := os.Stat(icloudFilePath)

	localMtime := localStat.ModTime().Unix()
	icloudMtime := icloudStat.ModTime().Unix()

	merged := sync.MergeConfigs(localValues, icloudValues, localMtime, icloudMtime)

	fmt.Printf("  Merged: %d values\n", len(merged))

	if dryRun {
		fmt.Printf("  Would upload merged config to iCloud\n")
		fmt.Printf("  Would update local config\n")
	} else {

		if err := sync.UploadToICloud(toolName, merged); err != nil {
			return fmt.Errorf("failed to upload merged config: %w", err)
		}

		tool.Values = merged
		conf.Tools[toolName] = tool
		if err := conf.Save(); err != nil {
			return fmt.Errorf("failed to save local config: %w", err)
		}

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

// applyTool applies desired state for a specific tool
func applyTool(conf *config.Config, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	if tool.Values == nil || len(tool.Values) == 0 {
		fmt.Printf("%s: No values to apply\n", toolName)
		return nil
	}

	fmt.Printf("Applying %s configuration...\n", toolName)

	if dryRun {
		fmt.Printf("  Would apply %d top-level config sections\n", len(tool.Values))
	} else {

		before, err := readFileContentSafe(tool.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read config file before applying: %w", err)
		}

		if err := tools.ApplyAllToolValues(toolName, tool.Values); err != nil {
			return fmt.Errorf("failed to apply %s configuration: %w", toolName, err)
		}

		after, err := readFileContentSafe(tool.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read config file after applying: %w", err)
		}

		if diff.DisplayUnifiedDiff(before, after, tool.ConfigPath) {
			fmt.Printf("  ✓ Applied configuration successfully\n")
		} else {
			fmt.Printf("  ✓ No changes needed (already in sync)\n")
		}
	}

	return nil
}

// showToolStatus shows drift status for a specific tool
func showToolStatus(conf *config.Config, toolName string) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("%s status:\n", toolName)

	flatValues := config.FlattenValues(tool.Values)
	if len(flatValues) == 0 {
		fmt.Printf("  No managed values\n")
		return nil
	}

	actualValues, err := getTargetConfigValues(toolName)
	if err != nil {
		return fmt.Errorf("failed to read current %s configuration: %w", toolName, err)
	}
	actualFlat := config.FlattenValues(actualValues)

	hasChanges := false
	for path, desiredValue := range flatValues {
		actualValue, exists := actualFlat[path]

		if !exists {
			fmt.Printf("  %s: MISSING (desired: %v)\n", path, desiredValue)
			hasChanges = true
		} else if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", desiredValue) {
			fmt.Printf("  %s: DRIFT (desired: %v, actual: %v)\n", path, desiredValue, actualValue)
			hasChanges = true
		} else {
			fmt.Printf("  %s: IN SYNC (%v)\n", path, actualValue)
		}
	}

	if !hasChanges {
		fmt.Printf("  All values in sync\n")
	}

	return nil
}
