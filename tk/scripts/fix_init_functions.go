// fix_init_functions.go - Fix init() functions to use new variable names
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	files := []string{
		"cmd/remote_ls.go",
		"cmd/relate_add.go",
		"cmd/project_ls.go",
		"cmd/project_rm.go",
		"cmd/schema_add.go",
		"cmd/schema_export.go",
		"cmd/schema_ls.go",
		"cmd/meta_ls.go",
		"cmd/meta_set.go",
		"cmd/debug_doctor.go",
		"cmd/debug_events_ls.go",
		"cmd/debug_events_show.go",
		"cmd/debug_events_stats.go",
		"cmd/debug_node_show.go",
		"cmd/debug_rebuild.go",
		"cmd/debug_repair.go",
		"cmd/log_query.go",
		"cmd/log_search.go",
		"cmd/sync_status.go",
		"cmd/task_conflicts.go",
	}

	for _, file := range files {
		if err := fixInitFunction(file); err != nil {
			fmt.Fprintf(os.Stderr, "Error fixing %s: %v\n", file, err)
			os.Exit(1)
		}
		fmt.Printf("✓ Fixed init() in %s\n", file)
	}

	fmt.Println("\nAll init() functions fixed!")
}

func fixInitFunction(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	text := string(content)

	// Determine the correct command variable name from the file
	cmdVarName := getCmdVarName(filePath)

	// Find all old Cmd references in init() and replace them
	// Pattern: SomeCmd.Flags() or SomeCmd.Something
	oldCmdRe := regexp.MustCompile(`\b([A-Z]\w+Cmd)\.(Flags|PersistentFlags|AddCommand)`)

	text = oldCmdRe.ReplaceAllStringFunc(text, func(match string) string {
		// Only replace if it's in an init() function
		// For simplicity, replace all old-style Cmd references
		parts := strings.Split(match, ".")
		return cmdVarName + "." + parts[1]
	})

	if err := os.WriteFile(filePath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func getCmdVarName(filePath string) string {
	// Extract command name from filename
	// e.g., cmd/remote_add.go -> remoteAddCmd
	base := filepath.Base(filePath)
	base = strings.TrimSuffix(base, ".go")

	// Convert underscores to camelCase
	parts := strings.Split(base, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return strings.Join(parts, "") + "Cmd"
}
