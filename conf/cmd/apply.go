package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/diff"
	"github.com/neongreen/mono/conf/pkg/folders"
	"github.com/neongreen/mono/conf/pkg/tools"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [tool-or-folder]",
	Short: "Apply desired state to target configuration files or folders",
	Long: `Sync the desired state from conf config to target files or folders.

Examples:
  conf apply          # Apply all tools and folders
  conf apply jj       # Apply only jj config
  conf apply my-docs  # Apply my-docs folder
  conf apply --dry-run jj  # Preview what would be applied`,
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

		if len(args) == 1 {
			name := args[0]

			// Check if it's a folder
			if _, exists := conf.Folders[name]; exists {
				if err := applyFolder(conf, configDir, name, dryRun); err != nil {
					fmt.Fprintf(os.Stderr, "Error applying folder %s: %v\n", name, err)
					os.Exit(1)
				}
				return
			}

			// Check if it's a tool
			if _, exists := conf.Tools[name]; exists {
				if err := applyTool(conf, name, dryRun); err != nil {
					fmt.Fprintf(os.Stderr, "Error applying %s: %v\n", name, err)
					os.Exit(1)
				}
				return
			}

			fmt.Fprintf(os.Stderr, "Error: %s is not a configured tool or folder\n", name)
			os.Exit(1)
		}

		// Apply all tools
		var toolNames []string
		for toolName := range conf.Tools {
			toolNames = append(toolNames, toolName)
		}
		sort.Strings(toolNames)

		for _, toolName := range toolNames {
			if err := applyTool(conf, toolName, dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Error applying %s: %v\n", toolName, err)
				os.Exit(1)
			}
		}

		// Apply all folders
		var folderNames []string
		for folderName := range conf.Folders {
			folderNames = append(folderNames, folderName)
		}
		sort.Strings(folderNames)

		for _, folderName := range folderNames {
			if err := applyFolder(conf, configDir, folderName, dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Error applying folder %s: %v\n", folderName, err)
				os.Exit(1)
			}
		}
	},
}

// applyTool applies desired state for a specific tool
func applyTool(conf *config.Config, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	if len(tool.Values) == 0 {
		fmt.Printf("%s: No values to apply\n", toolName)
		return nil
	}

	fmt.Printf("Applying %s configuration...\n", toolName)

	// Use the nested structure directly instead of flattening to strings
	// This avoids the need to parse quoted keys and is more efficient
	if dryRun {
		fmt.Printf("  Would apply %d top-level config sections\n", len(tool.Values))
	} else {
		// Capture the "before" state
		before, err := readFileContentSafe(tool.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read config file before applying: %w", err)
		}

		// Apply the configuration
		if err := tools.ApplyAllToolValues(toolName, tool.Values); err != nil {
			return fmt.Errorf("failed to apply %s configuration: %w", toolName, err)
		}

		// Capture the "after" state
		after, err := readFileContentSafe(tool.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read config file after applying: %w", err)
		}

		// Display diff if there were changes
		if diff.DisplayUnifiedDiff(before, after, tool.ConfigPath) {
			fmt.Printf("  ✓ Applied configuration successfully\n")
		} else {
			fmt.Printf("  ✓ No changes needed (already in sync)\n")
		}
	}

	return nil
}

// readFileContentSafe reads file content, returning empty string if file doesn't exist
func readFileContentSafe(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(content), nil
}

// applyFolder applies desired state from conf copy to source folder
func applyFolder(conf *config.Config, configDir, folderName string, dryRun bool) error {
	folder, exists := conf.GetFolder(folderName)
	if !exists {
		return fmt.Errorf("folder %s not configured", folderName)
	}

	fmt.Printf("Applying %s folder to %s...\n", folderName, folder.SourcePath)

	// Detect drift
	confPath := config.FolderCopyPath(configDir, folderName)
	drifts, err := folders.DetectDrift(folder.SourcePath, confPath)
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
		fmt.Printf("\nWould apply:\n")
		for _, drift := range drifts {
			fmt.Printf("  %s (%s)\n", drift.RelPath, drift.Status)
		}
		return nil
	}

	// Apply all drifted files
	if err := folders.ApplyAll(confPath, folder.SourcePath, drifts); err != nil {
		return fmt.Errorf("failed to apply: %w", err)
	}

	fmt.Printf("  ✓ Applied %d file(s)\n", len(drifts))
	return nil
}
