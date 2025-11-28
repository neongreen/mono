package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/folders"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [tool-or-folder]",
	Short: "Show drift between desired and actual state",
	Long: `Compare desired state in conf config with actual values in target files.

Examples:
  conf status         # Show status for all tools and folders
  conf status jj      # Show status for jj only
  conf status my-docs # Show status for my-docs folder`,
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
			// Check if it's a tool or a folder
			if _, exists := conf.Tools[name]; exists {
				if err := showToolStatus(conf, name); err != nil {
					fmt.Fprintf(os.Stderr, "Error checking %s status: %v\n", name, err)
					os.Exit(1)
				}
				return
			}
			if _, exists := conf.Folders[name]; exists {
				if err := showFolderStatus(conf, configDir, name); err != nil {
					fmt.Fprintf(os.Stderr, "Error checking %s status: %v\n", name, err)
					os.Exit(1)
				}
				return
			}
			fmt.Fprintf(os.Stderr, "Error: %s is not a configured tool or folder\n", name)
			os.Exit(1)
		}

		// Show all tools
		var toolNames []string
		for toolName := range conf.Tools {
			toolNames = append(toolNames, toolName)
		}
		sort.Strings(toolNames)

		for _, toolName := range toolNames {
			if err := showToolStatus(conf, toolName); err != nil {
				fmt.Fprintf(os.Stderr, "Error checking %s status: %v\n", toolName, err)
				os.Exit(1)
			}
		}

		// Show all folders
		var folderNames []string
		for folderName := range conf.Folders {
			folderNames = append(folderNames, folderName)
		}
		sort.Strings(folderNames)

		for _, folderName := range folderNames {
			if err := showFolderStatus(conf, configDir, folderName); err != nil {
				fmt.Fprintf(os.Stderr, "Error checking %s status: %v\n", folderName, err)
				os.Exit(1)
			}
		}
	},
}

var statusShowInSync bool

// showToolStatus shows drift status for a specific tool
func showToolStatus(conf *config.Config, toolName string) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("%s status:\n\n", toolName)

	actualValues, err := getTargetConfigValues(toolName)
	if err != nil {
		return fmt.Errorf("failed to read current %s configuration: %w", toolName, err)
	}

	renderStatusTable(toolName, tool.Values, actualValues, statusShowInSync)
	fmt.Println()

	return nil
}

// showFolderStatus shows drift status for a specific folder
func showFolderStatus(conf *config.Config, configDir, folderName string) error {
	folder, exists := conf.GetFolder(folderName)
	if !exists {
		return fmt.Errorf("folder %s not configured", folderName)
	}

	fmt.Printf("%s folder status:\n", folderName)
	fmt.Printf("  Source: %s\n\n", folder.SourcePath)

	// Check if source exists
	if _, err := os.Stat(folder.SourcePath); err != nil {
		fmt.Printf("  ⚠️  Source folder not found: %s\n\n", folder.SourcePath)
		return nil
	}

	// Detect drift
	confPath := config.FolderCopyPath(configDir, folderName)
	drifts, err := folders.DetectDrift(folder.SourcePath, confPath)
	if err != nil {
		return fmt.Errorf("failed to detect drift: %w", err)
	}

	if len(drifts) == 0 {
		fmt.Printf("  ✓ No drift detected\n\n")
		return nil
	}

	// Display drift summary
	fmt.Printf("  %s\n\n", folders.FormatDriftSummary(drifts))

	// Display drift details
	renderFolderDriftTable(drifts, statusShowInSync)
	fmt.Println()

	return nil
}

func init() {
	RootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVar(&statusShowInSync, "show-in-sync", false, "Include entries already in sync")
}
