// flatten_commands.go - Script to flatten nested command structure
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Transformation struct {
	SourcePath string
	DestPath   string
	OldPackage string
	NewCommand string
}

func main() {
	transformations := []Transformation{
		// Remote commands
		{"cmd/remote/add.go", "cmd/remote_add.go", "remote", "remote-add"},
		{"cmd/remote/ls.go", "cmd/remote_ls.go", "remote", "remote-ls"},
		{"cmd/remote/rm.go", "cmd/remote_rm.go", "remote", "remote-rm"},

		// Relate commands
		{"cmd/relate/add.go", "cmd/relate_add.go", "relate", "relate-add"},
		{"cmd/relate/ls.go", "cmd/relate_ls.go", "relate", "relate-ls"},
		{"cmd/relate/remove.go", "cmd/relate_rm.go", "relate", "relate-rm"},

		// Project commands
		{"cmd/project/create.go", "cmd/project_create.go", "project", "project-create"},
		{"cmd/project/ls.go", "cmd/project_ls.go", "project", "project-ls"},
		{"cmd/project/rename.go", "cmd/project_rename.go", "project", "project-rename"},
		{"cmd/project/rm.go", "cmd/project_rm.go", "project", "project-rm"},

		// Queue commands
		{"cmd/queue/create.go", "cmd/queue_create.go", "queue", "queue-create"},
		{"cmd/queue/push.go", "cmd/queue_push.go", "queue", "queue-push"},
		{"cmd/queue/pop.go", "cmd/queue_pop.go", "queue", "queue-pop"},
		{"cmd/queue/list.go", "cmd/queue_ls.go", "queue", "queue-ls"},
		{"cmd/queue/show.go", "cmd/queue_show.go", "queue", "queue-show"},
		{"cmd/queue/rename.go", "cmd/queue_rename.go", "queue", "queue-rename"},
		{"cmd/queue/rm.go", "cmd/queue_rm.go", "queue", "queue-rm"},

		// Stack commands
		{"cmd/stack/create.go", "cmd/stack_create.go", "stack", "stack-create"},
		{"cmd/stack/push.go", "cmd/stack_push.go", "stack", "stack-push"},
		{"cmd/stack/pop.go", "cmd/stack_pop.go", "stack", "stack-pop"},
		{"cmd/stack/list.go", "cmd/stack_ls.go", "stack", "stack-ls"},
		{"cmd/stack/show.go", "cmd/stack_show.go", "stack", "stack-show"},
		{"cmd/stack/rename.go", "cmd/stack_rename.go", "stack", "stack-rename"},
		{"cmd/stack/rm.go", "cmd/stack_rm.go", "stack", "stack-rm"},

		// Group commands
		{"cmd/group/create.go", "cmd/group_create.go", "group", "group-create"},
		{"cmd/group/add.go", "cmd/group_addtask.go", "group", "group-addtask"},
		{"cmd/group/remove.go", "cmd/group_rmtask.go", "group", "group-rmtask"},
		{"cmd/group/list.go", "cmd/group_ls.go", "group", "group-ls"},
		{"cmd/group/show.go", "cmd/group_show.go", "group", "group-show"},
		{"cmd/group/rename.go", "cmd/group_rename.go", "group", "group-rename"},
		{"cmd/group/rm.go", "cmd/group_delete.go", "group", "group-delete"},

		// Schema commands
		{"cmd/schema/add_kind.go", "cmd/schema_add.go", "schema", "schema-add"},
		{"cmd/schema/list_kinds.go", "cmd/schema_ls.go", "schema", "schema-ls"},
		{"cmd/schema/export.go", "cmd/schema_export.go", "schema", "schema-export"},

		// Meta commands
		{"cmd/meta/set.go", "cmd/meta_set.go", "meta", "meta-set"},
		{"cmd/meta/get.go", "cmd/meta_get.go", "meta", "meta-get"},
		{"cmd/meta/list.go", "cmd/meta_ls.go", "meta", "meta-ls"},
		{"cmd/meta/claims.go", "cmd/meta_claims.go", "meta", "meta-claims"},

		// Debug commands
		{"cmd/debug/doctor.go", "cmd/debug_doctor.go", "debug", "debug-doctor"},
		{"cmd/debug/repair.go", "cmd/debug_repair.go", "debug", "debug-repair"},
		{"cmd/debug/rebuild.go", "cmd/debug_rebuild.go", "debug", "debug-rebuild"},
		{"cmd/debug/events/list.go", "cmd/debug_events_ls.go", "events", "debug-events-ls"},
		{"cmd/debug/events/show.go", "cmd/debug_events_show.go", "events", "debug-events-show"},
		{"cmd/debug/events/stats.go", "cmd/debug_events_stats.go", "events", "debug-events-stats"},
		{"cmd/debug/node/show.go", "cmd/debug_node_show.go", "node", "debug-node-show"},
		{"cmd/debug/node/regen.go", "cmd/debug_node_regen.go", "node", "debug-node-regen"},

		// Migrate commands
		{"cmd/migrate/fix_container_item_ids.go", "cmd/migrate_fix_container_item_ids.go", "migrate", "migrate-fix-container-item-ids"},
		{"cmd/migrate/fix_relocate_bug.go", "cmd/migrate_fix_relocate_bug.go", "migrate", "migrate-fix-relocate-bug"},
		{"cmd/migrate/scan_deprecated.go", "cmd/migrate_scan_deprecated.go", "migrate", "migrate-scan-deprecated"},

		// Log commands
		{"cmd/log/query.go", "cmd/log_query.go", "log", "log-query"},
		{"cmd/log/search.go", "cmd/log_search.go", "log", "log-search"},

		// Conflicts commands
		{"cmd/conflicts/numbers.go", "cmd/task_conflicts.go", "conflicts", "task-conflicts"},

		// Status commands
		{"cmd/status/sync.go", "cmd/sync_status.go", "status", "sync-status"},
	}

	for _, t := range transformations {
		if err := transformFile(t); err != nil {
			fmt.Fprintf(os.Stderr, "Error transforming %s: %v\n", t.SourcePath, err)
			os.Exit(1)
		}
		fmt.Printf("✓ %s -> %s\n", t.SourcePath, t.DestPath)
	}

	fmt.Println("\nTransformation complete!")
}

func transformFile(t Transformation) error {
	// Read source file
	content, err := os.ReadFile(t.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	text := string(content)

	// 1. Change package name
	packageRe := regexp.MustCompile(`package ` + t.OldPackage)
	text = packageRe.ReplaceAllString(text, "package cmd")

	// 2. Change command variable name
	// Extract old command name from Use field or filename
	cmdName := strings.TrimPrefix(t.NewCommand, strings.Split(t.NewCommand, "-")[0]+"-")
	if cmdName == t.NewCommand {
		cmdName = filepath.Base(t.SourcePath)
		cmdName = strings.TrimSuffix(cmdName, ".go")
	}

	// Build new variable name (e.g., remote-add -> remoteAddCmd)
	newVarName := toCamelCase(t.NewCommand) + "Cmd"

	// Find and replace var declaration
	varRe := regexp.MustCompile(`var\s+(\w+Cmd)\s*=`)
	text = varRe.ReplaceAllStringFunc(text, func(match string) string {
		return "var " + newVarName + " ="
	})

	// 3. Change Use field
	useRe := regexp.MustCompile(`Use:\s*"[^"]*"`)
	text = useRe.ReplaceAllString(text, `Use:   "`+t.NewCommand+`"`)

	// Write to destination
	if err := os.WriteFile(t.DestPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "-")
	for i := range parts {
		if len(parts[i]) > 0 {
			if i == 0 {
				// Keep first part lowercase
				parts[i] = parts[i]
			} else {
				// Capitalize subsequent parts
				parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
			}
		}
	}
	return strings.Join(parts, "")
}
