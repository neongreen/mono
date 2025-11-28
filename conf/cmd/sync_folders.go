package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/folders"
	"github.com/neongreen/mono/conf/pkg/sync"
)

// syncFolder syncs a folder with iCloud Drive
func syncFolder(conf *config.Config, configDir, folderName string, dryRun bool) error {
	folder, exists := conf.GetFolder(folderName)
	if !exists {
		return fmt.Errorf("folder %s not configured", folderName)
	}

	fmt.Printf("Syncing folder %s...\n", folderName)

	// Paths
	confCopyPath := config.FolderCopyPath(configDir, folderName)
	icloudPath, err := sync.ICloudDrivePath()
	if err != nil {
		return fmt.Errorf("failed to get iCloud Drive path: %w", err)
	}
	icloudFolderPath := filepath.Join(icloudPath, folderName)

	// Step 1: Import drift from source to conf copy
	// This ensures conf copy is up-to-date with source before syncing to iCloud
	sourcePath := folder.SourcePath
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("  Warning: Source folder %s does not exist, skipping source sync\n", sourcePath)
		} else {
			return fmt.Errorf("failed to stat source folder: %w", err)
		}
	} else {
		// Detect drift between source and conf copy
		drifts, err := folders.DetectDriftWithExcludes(sourcePath, confCopyPath, folder.Exclude)
		if err != nil {
			return fmt.Errorf("failed to detect drift: %w", err)
		}

		if len(drifts) > 0 {
			fmt.Printf("  Found %d changes in source folder\n", len(drifts))
			if dryRun {
				fmt.Printf("  Would import changes from source to conf copy\n")
			} else {
				// Import changes from source
				if err := folders.ImportAll(sourcePath, confCopyPath, drifts); err != nil {
					return fmt.Errorf("failed to import from source: %w", err)
				}
				fmt.Printf("  ✓ Imported changes from source\n")
			}
		} else {
			fmt.Printf("  Source folder is in sync with conf copy\n")
		}
	}

	// Step 2: Sync conf copy with iCloud
	// Check if iCloud folder exists
	if _, err := os.Stat(icloudFolderPath); os.IsNotExist(err) {
		// No iCloud copy, upload conf copy
		fmt.Printf("  No iCloud copy found, uploading...\n")
		if !dryRun {
			if err := syncCopyDirToICloud(confCopyPath, icloudFolderPath); err != nil {
				return fmt.Errorf("failed to upload to iCloud: %w", err)
			}
			fmt.Printf("  ✓ Uploaded to iCloud\n")
		}
		return nil
	}

	// Detect drift between conf copy and iCloud
	// Note: We use the same exclude patterns for iCloud sync to maintain consistency
	icloudDrifts, err := folders.DetectDriftWithExcludes(confCopyPath, icloudFolderPath, folder.Exclude)
	if err != nil {
		return fmt.Errorf("failed to detect iCloud drift: %w", err)
	}

	if len(icloudDrifts) == 0 {
		fmt.Printf("  Conf copy and iCloud are already in sync\n")
		return nil
	}

	fmt.Printf("  Found %d differences between conf copy and iCloud\n", len(icloudDrifts))

	if dryRun {
		// Show what would be synced
		for _, drift := range icloudDrifts {
			action := ""
			switch drift.Status {
			case folders.StatusModified:
				action = "MODIFY"
			case folders.StatusAdded:
				action = "ADD"
			case folders.StatusDeleted:
				action = "DELETE"
			}
			fmt.Printf("    [%s] %s\n", action, drift.RelPath)
		}
		fmt.Printf("  Would merge and upload to iCloud\n")
		return nil
	}

	// Merge using Last-Write-Wins strategy
	// For each drift, compare modification times
	confStat, _ := os.Stat(confCopyPath)
	icloudStat, _ := os.Stat(icloudFolderPath)

	if icloudStat.ModTime().After(confStat.ModTime()) {
		// iCloud is newer, download to conf copy
		fmt.Printf("  iCloud copy is newer, downloading...\n")
		for _, drift := range icloudDrifts {
			switch drift.Status {
			case folders.StatusModified, folders.StatusDeleted:
				// File exists in iCloud, copy to conf
				if err := folders.ApplyFile(icloudFolderPath, confCopyPath, drift.RelPath); err != nil {
					return fmt.Errorf("failed to apply iCloud file %s: %w", drift.RelPath, err)
				}
			case folders.StatusAdded:
				// File added in conf but not in iCloud, remove from conf to match iCloud
				if err := folders.DeleteFile(confCopyPath, drift.RelPath); err != nil {
					return fmt.Errorf("failed to delete file %s: %w", drift.RelPath, err)
				}
			}
		}
		fmt.Printf("  ✓ Downloaded from iCloud\n")
	} else {
		// Conf copy is newer or same age, upload to iCloud
		fmt.Printf("  Conf copy is newer, uploading...\n")
		for _, drift := range icloudDrifts {
			switch drift.Status {
			case folders.StatusModified, folders.StatusAdded:
				// Copy from conf to iCloud
				if err := folders.ApplyFile(confCopyPath, icloudFolderPath, drift.RelPath); err != nil {
					return fmt.Errorf("failed to upload file %s: %w", drift.RelPath, err)
				}
			case folders.StatusDeleted:
				// File deleted in conf, delete from iCloud
				if err := folders.DeleteFile(icloudFolderPath, drift.RelPath); err != nil {
					return fmt.Errorf("failed to delete iCloud file %s: %w", drift.RelPath, err)
				}
			}
		}
		fmt.Printf("  ✓ Uploaded to iCloud\n")
	}

	return nil
}

// syncCopyDirToICloud recursively copies a directory to iCloud
func syncCopyDirToICloud(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory with same permissions
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := syncCopyDirToICloud(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := syncCopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// syncCopyFile copies a single file for sync
func syncCopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
