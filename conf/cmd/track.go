package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/neongreen/mono/conf/pkg/config"
	tomlv2 "github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track <source-path> --name <name>",
	Short: "Start tracking a folder for sync",
	Long: `Track a folder by copying it to conf's directory and creating a manifest.

The folder will be copied to ~/.config/conf/<name>/ and a manifest file will be
created at ~/.config/conf/<name>.toml that tracks the source path.

Examples:
  conf track ~/Documents/my-docs --name my-docs
  conf track ~/scripts --name scripts --exclude "*.tmp" --exclude ".DS_Store"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		folderName, _ := cmd.Flags().GetString("name")
		excludePatterns, _ := cmd.Flags().GetStringSlice("exclude")

		if folderName == "" {
			return fmt.Errorf("--name flag is required")
		}

		// Load config
		conf, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		configDir, err := config.ConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}

		// Expand source path
		expandedSource, err := config.ExpandPath(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to expand source path: %w", err)
		}

		// Check if source exists
		sourceInfo, err := os.Stat(expandedSource)
		if err != nil {
			return fmt.Errorf("source path does not exist: %w", err)
		}
		if !sourceInfo.IsDir() {
			return fmt.Errorf("source path is not a directory: %s", expandedSource)
		}

		// Check if folder is already tracked
		if _, exists := conf.Folders[folderName]; exists {
			return fmt.Errorf("folder %s is already tracked", folderName)
		}

		// Create folder copy path
		folderCopyPath := config.FolderCopyPath(configDir, folderName)
		if _, err := os.Stat(folderCopyPath); err == nil {
			return fmt.Errorf("folder copy already exists at %s", folderCopyPath)
		}

		fmt.Printf("Tracking folder: %s\n", expandedSource)
		fmt.Printf("  Name: %s\n", folderName)
		fmt.Printf("  Copy to: %s\n", folderCopyPath)

		// Copy folder to conf directory, excluding specified patterns
		if err := copyDir(expandedSource, folderCopyPath, excludePatterns); err != nil {
			return fmt.Errorf("failed to copy folder: %w", err)
		}
		fmt.Printf("  ✓ Copied folder\n")

		// Create folder config
		folderConfig := config.FolderConfig{
			Name:         folderName,
			SourcePath:   sourcePath, // Keep in tilde notation
			TrackedSince: time.Now().Format(time.RFC3339),
			Exclude:      excludePatterns,
		}

		// Save manifest file
		manifestPath := config.FolderManifestPath(configDir, folderName)
		manifestData, err := tomlv2.Marshal(folderConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal manifest: %w", err)
		}
		if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
			return fmt.Errorf("failed to write manifest: %w", err)
		}
		fmt.Printf("  ✓ Created manifest at %s\n", manifestPath)

		// Register in main config
		conf.SetFolder(folderName, folderConfig)
		if err := conf.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("  ✓ Registered in config\n")

		fmt.Printf("\n✓ Folder %s is now tracked\n", folderName)
		return nil
	},
}

// copyDir recursively copies a directory, skipping files that match exclude patterns
func copyDir(src, dst string, excludePatterns []string) error {
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
		// Check if entry should be excluded
		if shouldExclude(entry.Name(), excludePatterns) {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath, excludePatterns); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// shouldExclude checks if a filename matches any of the exclude patterns.
// Invalid patterns are silently ignored (treated as non-matching).
func shouldExclude(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		// Error from filepath.Match indicates invalid pattern syntax - we skip such patterns
		if matched, err := filepath.Match(pattern, filename); err == nil && matched {
			return true
		}
	}
	return false
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

func init() {
	RootCmd.AddCommand(trackCmd)
	trackCmd.Flags().String("name", "", "Name for the tracked folder (required)")
	trackCmd.Flags().StringSlice("exclude", []string{}, "Patterns to exclude from sync")
	trackCmd.MarkFlagRequired("name")
}
