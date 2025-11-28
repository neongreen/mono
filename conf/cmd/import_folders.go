package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/folders"
)

// importFolder imports all drifted files from source folder to conf copy
func importFolder(conf *config.Config, configDir, folderName string, dryRun bool) error {
	folder, exists := conf.GetFolder(folderName)
	if !exists {
		return fmt.Errorf("folder %s not configured", folderName)
	}

	fmt.Printf("Importing %s folder from %s...\n", folderName, folder.SourcePath)

	// Check if source exists
	if _, err := os.Stat(folder.SourcePath); err != nil {
		return fmt.Errorf("source folder not found: %w", err)
	}

	// Detect drift
	confPath := config.FolderCopyPath(configDir, folderName)
	drifts, err := folders.DetectDriftWithExcludes(folder.SourcePath, confPath, folder.Exclude)
	if err != nil {
		return fmt.Errorf("failed to detect drift: %w", err)
	}

	if len(drifts) == 0 {
		fmt.Printf("  ✓ No drift detected\n")
		return nil
	}

	// Display drift summary
	fmt.Printf("  %s\n", folders.FormatDriftSummary(drifts))

	if dryRun {
		fmt.Printf("\nWould import:\n")
		for _, drift := range drifts {
			fmt.Printf("  %s (%s)\n", drift.RelPath, drift.Status)
		}
		return nil
	}

	// Import all drifted files
	if err := folders.ImportAll(folder.SourcePath, confPath, drifts); err != nil {
		return fmt.Errorf("failed to import: %w", err)
	}

	fmt.Printf("  ✓ Imported %d file(s)\n", len(drifts))
	return nil
}

// importFolderFile imports a specific file from source folder to conf copy
func importFolderFile(conf *config.Config, configDir, folderName, relPath string, dryRun bool) error {
	folder, exists := conf.GetFolder(folderName)
	if !exists {
		return fmt.Errorf("folder %s not configured", folderName)
	}

	fmt.Printf("Importing %s/%s from source...\n", folderName, relPath)

	// Check if source file exists
	sourcePath := filepath.Join(folder.SourcePath, relPath)
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("source file not found: %w", err)
	}

	if dryRun {
		fmt.Printf("  Would import: %s\n", relPath)
		return nil
	}

	// Import the file
	confPath := config.FolderCopyPath(configDir, folderName)
	if err := folders.ImportFile(folder.SourcePath, confPath, relPath); err != nil {
		return fmt.Errorf("failed to import: %w", err)
	}

	fmt.Printf("  ✓ Imported %s\n", relPath)
	return nil
}
