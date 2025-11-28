package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/folders"
	"github.com/neongreen/mono/lib/cli"
	"github.com/neongreen/mono/lib/configschema"
)

// renderSettingsTable renders a table of settings with colors and proper formatting
func renderSettingsTable(settings []configschema.SettingInfo, configPath string) {
	if len(settings) == 0 {
		fmt.Println("No settings available")
		return
	}

	t := cli.NewTable(os.Stdout)
	t.AppendHeader(table.Row{"Setting", "Type", "Value", "Description"})

	// Configure column widths for better display
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: false, WidthMax: 40},                                  // Setting path
		{Number: 2, AutoMerge: false, WidthMax: 15},                                  // Type
		{Number: 3, AutoMerge: false, WidthMax: 30, WidthMaxEnforcer: text.WrapSoft}, // Value
		{Number: 4, AutoMerge: false, WidthMax: 50, WidthMaxEnforcer: text.WrapSoft}, // Description
	})

	for _, setting := range settings {
		// Format the setting path with color
		pathStr := cli.Key(setting.Path)

		// Format type with color
		typeStr := cli.Type(setting.Type)

		// Format value with color and status
		var valueStr string
		if setting.IsSet {
			valueStr = cli.Successf("✓ %v", formatValue(setting.CurrentValue))
		} else if setting.Default != nil {
			valueStr = cli.Secondaryf("(default: %v)", formatValue(setting.Default))
		} else {
			valueStr = cli.Muted("(not set)")
		}

		// Add enum values if present
		if len(setting.Enum) > 0 {
			valueStr += "\n" + cli.Mutedf("values: %s", strings.Join(setting.Enum, ", "))
		}

		// Format description
		descStr := setting.Description
		if descStr == "" {
			descStr = cli.Muted("-")
		}

		t.AppendRow(table.Row{pathStr, typeStr, valueStr, descStr})
	}

	t.Render()

	// Show config file path at the bottom
	fmt.Println()
	fmt.Printf("Config file: %s\n", configPath)
}

// renderStatusTable shows drift between desired and actual state using a readable table
func renderStatusTable(toolName string, desired map[string]any, actual map[string]any, showInSync bool) {
	desiredFlat := config.FlattenValues(desired)
	actualFlat := config.FlattenValues(actual)

	// Build sorted set of all paths
	pathSet := make(map[string]struct{})
	for k := range desiredFlat {
		pathSet[k] = struct{}{}
	}
	for k := range actualFlat {
		pathSet[k] = struct{}{}
	}
	var paths []string
	for k := range pathSet {
		paths = append(paths, k)
	}
	sort.Strings(paths)

	t := cli.NewTable(os.Stdout)
	t.AppendHeader(table.Row{"Path", "Status", "Value"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: 42},
		{Number: 2, WidthMax: 14},
		{Number: 3, WidthMax: 70, WidthMaxEnforcer: text.WrapSoft},
	})

	var inSync, drift, missing, unmanaged int

	for _, path := range paths {
		desiredVal, hasDesired := desiredFlat[path]
		actualVal, hasActual := actualFlat[path]

		switch {
		case hasDesired && !hasActual:
			missing++
			t.AppendRow(table.Row{
				cli.Key(path),
				cli.Error("MISSING"),
				fmt.Sprintf("%s\n%s",
					formatLabeledValue("(desired)", desiredVal),
					formatLabeledValue("(local)", nil)),
			})
		case !hasDesired && hasActual:
			unmanaged++
			t.AppendRow(table.Row{
				cli.Key(path),
				cli.Warning("UNMANAGED"),
				formatLabeledValue("(local)", actualVal),
			})
		case hasDesired && hasActual:
			if fmt.Sprintf("%v", desiredVal) == fmt.Sprintf("%v", actualVal) {
				inSync++
				if showInSync {
					t.AppendRow(table.Row{
						cli.Key(path),
						cli.Success("IN SYNC"),
						formatLabeledValue("(desired/local)", desiredVal),
					})
				}
			} else {
				drift++
				t.AppendRow(table.Row{
					cli.Key(path),
					cli.Warning("DRIFT"),
					fmt.Sprintf("%s\n%s",
						formatLabeledValue("(desired)", desiredVal),
						formatLabeledValue("(local)", actualVal)),
				})
			}
		}
	}

	t.Render()

	fmt.Printf("\nSummary: %s in sync, %s drift, %s missing, %s unmanaged\n",
		cli.Successf("%d", inSync),
		cli.Warningf("%d", drift),
		cli.Errorf("%d", missing),
		cli.Mutedf("%d", unmanaged),
	)
	if !showInSync && inSync > 0 {
		fmt.Printf("Hidden %s in-sync values; rerun with --show-in-sync to include them.\n", cli.Successf("%d", inSync))
	}
}

// renderFolderDriftTable renders a table of folder drift
func renderFolderDriftTable(drifts []folders.FileDrift, showInSync bool) {
	if len(drifts) == 0 && !showInSync {
		return
	}

	// Sort drifts by path
	sort.Slice(drifts, func(i, j int) bool {
		return drifts[i].RelPath < drifts[j].RelPath
	})

	t := cli.NewTable(os.Stdout)
	t.AppendHeader(table.Row{"File", "Status", "Details"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: 50, WidthMaxEnforcer: text.WrapSoft},
		{Number: 2, WidthMax: 12},
		{Number: 3, WidthMax: 40, WidthMaxEnforcer: text.WrapSoft},
	})

	for _, drift := range drifts {
		if drift.Status == folders.StatusInSync && !showInSync {
			continue
		}

		var statusStr, detailsStr string

		switch drift.Status {
		case folders.StatusInSync:
			statusStr = cli.Success("IN SYNC")
			detailsStr = cli.Muted("Identical")
		case folders.StatusModified:
			statusStr = cli.Warning("MODIFIED")
			if drift.SourceMtime > drift.ConfMtime {
				detailsStr = cli.Warningf("Source newer")
			} else {
				detailsStr = cli.Warningf("Conf newer")
			}
		case folders.StatusAdded:
			statusStr = cli.Warning("ADDED")
			detailsStr = "New in source"
		case folders.StatusDeleted:
			statusStr = cli.Warning("DELETED")
			detailsStr = "Removed from source"
		}

		pathStr := cli.Key(drift.RelPath)
		if drift.IsDir {
			pathStr = pathStr + "/"
		}

		t.AppendRow(table.Row{pathStr, statusStr, detailsStr})
	}

	t.Render()
}
