package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/neongreen/mono/conf/pkg/config"
	claudetool "github.com/neongreen/mono/conf/pkg/tools/claude"
	jjtool "github.com/neongreen/mono/conf/pkg/tools/jj"
	misetool "github.com/neongreen/mono/conf/pkg/tools/mise"
	starshiptool "github.com/neongreen/mono/conf/pkg/tools/starship"
	"github.com/neongreen/mono/lib/cli"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [tool]",
	Short: "Import configuration values from target files into conf state",
	Long: `Read configuration values from target files (e.g., ~/.config/jj/config.toml) 
and import them into conf's state management (stored in ~/.config/conf/).

This is useful for:
- Migrating existing configurations to conf management
- Capturing manual changes made to config files
- Syncing local configurations into conf state

Examples:
  conf import           # Import all tools
  conf import jj        # Import only jj config
  conf import --dry-run # Preview what would be imported`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to load conf config: %v\n", err)
			os.Exit(1)
		}

		var toolsToImport []string
		if len(args) == 1 {
			toolsToImport = []string{args[0]}
		} else {
			// Import all tools
			for toolName := range conf.Tools {
				toolsToImport = append(toolsToImport, toolName)
			}
		}
		sort.Strings(toolsToImport)

		for _, toolName := range toolsToImport {
			if err := importTool(conf, toolName, dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Error importing %s: %v\n", toolName, err)
				os.Exit(1)
			}
		}

		if !dryRun {
			fmt.Println("\n✓ Import complete")
		}
	},
}

// importTool imports configuration values from a tool's target file into conf's state
func importTool(conf *config.Config, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("Importing %s configuration from %s...\n", toolName, tool.ConfigPath)

	// Get all values from the target config file
	values, err := getTargetConfigValues(toolName)
	if err != nil {
		return fmt.Errorf("failed to read target config: %w", err)
	}

	flatValues := config.FlattenValues(values)
	existingFlat := config.FlattenValues(tool.Values)

	if len(flatValues) == 0 {
		fmt.Printf("  No values found in %s\n", tool.ConfigPath)
		return nil
	}

	fmt.Printf("  Found %d values\n", len(flatValues))

	if dryRun {
		renderImportPreview(existingFlat, flatValues)
		// Compatibility concise lines
		for path, value := range flatValues {
			fmt.Printf("  Would import: %s.%s = %v\n", toolName, path, value)
		}
		return nil
	}

	// Import all values into conf state
	renderImportPreview(existingFlat, flatValues)
	for path, value := range flatValues {
		fmt.Printf("  ✓ Imported %s.%s = %v\n", toolName, path, value)
	}

	conf.MergeToolValues(toolName, values)

	// Save conf state
	if err := conf.Save(); err != nil {
		return fmt.Errorf("failed to save conf state: %w", err)
	}

	fmt.Printf("  ✓ Saved to conf state\n")

	return nil
}

// getTargetConfigValues reads all values from a tool's target config file
func getTargetConfigValues(toolName string) (map[string]any, error) {
	switch toolName {
	case "jj":
		jjTool, err := jjtool.NewJJTool()
		if err != nil {
			return nil, err
		}
		return jjTool.GetAllValues()
	case "claude":
		claudeTool, err := claudetool.NewClaudeTool()
		if err != nil {
			return nil, err
		}
		return claudeTool.GetAllValues()
	case "mise":
		miseTool, err := misetool.NewMiseTool()
		if err != nil {
			return nil, err
		}
		return miseTool.GetAllValues()
	case "starship":
		starshipTool, err := starshiptool.NewStarshipTool()
		if err != nil {
			return nil, err
		}
		return starshipTool.GetAllValues()
	default:
		return nil, fmt.Errorf("unsupported tool: %s", toolName)
	}
}

// renderImportPreview renders a table showing what would be imported vs current conf state.
func renderImportPreview(existingFlat map[string]any, incomingFlat map[string]any) {
	type row struct {
		path     string
		status   string
		incoming any
		current  any
	}

	var rows []row
	var added, updated, same int

	for path, incoming := range incomingFlat {
		current, hasCurrent := existingFlat[path]
		switch {
		case !hasCurrent:
			added++
			rows = append(rows, row{
				path:     path,
				status:   cli.Success("NEW"),
				incoming: incoming,
				current:  nil,
			})
		case fmt.Sprintf("%v", incoming) == fmt.Sprintf("%v", current):
			same++
			rows = append(rows, row{
				path:     path,
				status:   cli.Muted("SAME"),
				incoming: incoming,
				current:  current,
			})
		default:
			updated++
			rows = append(rows, row{
				path:     path,
				status:   cli.Warning("UPDATE"),
				incoming: incoming,
				current:  current,
			})
		}
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].path < rows[j].path })

	t := cli.NewTable(os.Stdout)
	t.AppendHeader(table.Row{"Path", "Status", "Incoming", "Current"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: 42},
		{Number: 2, WidthMax: 10},
		{Number: 3, WidthMax: 50, WidthMaxEnforcer: text.WrapSoft},
		{Number: 4, WidthMax: 50, WidthMaxEnforcer: text.WrapSoft},
	})

	for _, r := range rows {
		t.AppendRow(table.Row{
			cli.Key(r.path),
			r.status,
			cli.Value(formatValueShort(r.incoming)),
			cli.Muted(formatValueShort(r.current)),
		})
	}
	t.Render()

	fmt.Printf("\nSummary: %s new, %s updated, %s unchanged\n",
		cli.Successf("%d", added),
		cli.Warningf("%d", updated),
		cli.Mutedf("%d", same),
	)
}
