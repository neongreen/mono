package folders

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ImportFile imports a single file from source to conf copy
func ImportFile(sourcePath, confPath, relPath string) error {
	srcFile := filepath.Join(sourcePath, relPath)
	dstFile := filepath.Join(confPath, relPath)

	// Check if source file exists
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		return fmt.Errorf("source file does not exist: %w", err)
	}

	// If it's a directory, create it in conf
	if srcInfo.IsDir() {
		return os.MkdirAll(dstFile, srcInfo.Mode())
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dstFile), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Copy file
	return copyFile(srcFile, dstFile)
}

// ApplyFile applies a single file from conf copy to source
func ApplyFile(confPath, sourcePath, relPath string) error {
	srcFile := filepath.Join(confPath, relPath)
	dstFile := filepath.Join(sourcePath, relPath)

	// Check if conf file exists
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		return fmt.Errorf("conf file does not exist: %w", err)
	}

	// If it's a directory, create it in source
	if srcInfo.IsDir() {
		return os.MkdirAll(dstFile, srcInfo.Mode())
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dstFile), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Copy file
	return copyFile(srcFile, dstFile)
}

// DeleteFile deletes a file from the target path
func DeleteFile(targetPath, relPath string) error {
	filePath := filepath.Join(targetPath, relPath)
	return os.RemoveAll(filePath)
}

// ImportAll imports all drifted files from source to conf
func ImportAll(sourcePath, confPath string, drifts []FileDrift) error {
	for _, drift := range drifts {
		switch drift.Status {
		case StatusAdded, StatusModified:
			// Copy from source to conf
			if err := ImportFile(sourcePath, confPath, drift.RelPath); err != nil {
				return fmt.Errorf("failed to import %s: %w", drift.RelPath, err)
			}
		case StatusDeleted:
			// Delete from conf
			if err := DeleteFile(confPath, drift.RelPath); err != nil {
				return fmt.Errorf("failed to delete %s from conf: %w", drift.RelPath, err)
			}
		}
	}
	return nil
}

// ApplyAll applies all drifted files from conf to source
func ApplyAll(confPath, sourcePath string, drifts []FileDrift) error {
	for _, drift := range drifts {
		switch drift.Status {
		case StatusAdded:
			// File was added in source, delete from source (restore to conf state)
			if err := DeleteFile(sourcePath, drift.RelPath); err != nil {
				return fmt.Errorf("failed to delete %s from source: %w", drift.RelPath, err)
			}
		case StatusModified, StatusDeleted:
			// Copy from conf to source (deleted files exist in conf)
			if err := ApplyFile(confPath, sourcePath, drift.RelPath); err != nil {
				return fmt.Errorf("failed to apply %s: %w", drift.RelPath, err)
			}
		}
	}
	return nil
}

// copyFile copies a single file, preserving permissions
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}
