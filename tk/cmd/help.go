package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// seeAlsoRegistry maps command paths to related commands that should appear
// in the "See Also" section of their help text.
//
// This registry improves command discoverability, especially for AI agents
// who might not guess at less-obvious command names.
var seeAlsoRegistry = map[string][]string{
	// Core task management
	"new":      {"show", "mark", "edit", "describe", "ls"},
	"show":     {"edit", "describe", "note", "history", "relate ls", "attach"},
	"ls":       {"show", "mark", "new", "project ls", "blocked"},
	"mark":     {"show", "ls", "blockers"},
	"edit":     {"describe", "show", "note", "mv"},
	"describe": {"edit", "show", "note"},
	"note":     {"show", "describe", "history"},
	"attach":   {"show", "note"},
	"mv":       {"edit", "project ls", "show"},
	"rm":       {"show", "ls"},

	// Relations & dependencies
	"relate":        {"relate add", "relate ls", "dup", "blockers", "graph"},
	"relate add":    {"relate ls", "relate remove", "dup", "graph"},
	"relate remove": {"relate ls", "relate add"},
	"relate ls":     {"relate add", "graph", "show", "blockers"},
	"dup":           {"relate add", "relate ls"},
	"blockers":      {"blocked", "relate add", "mark", "graph"},
	"blocked":       {"blockers", "mark", "relate ls"},
	"graph":         {"relate ls", "blockers", "show"},

	// Project management
	"project":        {"project create", "project ls", "project rename"},
	"project create": {"new", "mv", "project ls"},
	"project ls":     {"project create", "ls", "mv"},
	"project rm":     {"project ls", "mv"},
	"project rename": {"project ls"},

	// Container management - queues
	"queue":        {"queue create", "queue push", "queue pop", "queue list", "stack", "group"},
	"queue create": {"queue push", "queue list", "new"},
	"queue push":   {"queue pop", "queue show", "ls"},
	"queue pop":    {"queue push", "queue show", "mark"},
	"queue list":   {"queue show", "queue create", "stack list", "group list"},
	"queue show":   {"queue push", "queue pop", "ls"},
	"queue rename": {"queue list"},
	"queue rm":     {"queue list"},

	// Container management - stacks
	"stack":        {"stack create", "stack push", "stack pop", "stack list", "queue", "group"},
	"stack create": {"stack push", "stack list", "new"},
	"stack push":   {"stack pop", "stack show", "ls"},
	"stack pop":    {"stack push", "stack show", "mark"},
	"stack list":   {"stack show", "stack create", "queue list", "group list"},
	"stack show":   {"stack push", "stack pop", "ls"},
	"stack rename": {"stack list"},
	"stack rm":     {"stack list"},

	// Container management - groups
	"group":        {"group create", "group add", "group list", "queue", "stack"},
	"group create": {"group add", "group list"},
	"group add":    {"group show", "group remove", "ls"},
	"group remove": {"group add", "group show"},
	"group list":   {"group show", "group create", "queue list", "stack list"},
	"group show":   {"group add", "ls"},
	"group rename": {"group list"},
	"group rm":     {"group list"},

	// Schema & metadata
	"schema":        {"schema add", "schema list", "schema export", "meta"},
	"schema add":    {"schema list", "meta set"},
	"schema list":   {"schema add", "schema export", "meta list"},
	"schema export": {"schema list", "schema add"},
	"meta":          {"meta set", "meta get", "meta list", "schema"},
	"meta set":      {"meta get", "meta list", "show"},
	"meta get":      {"meta set", "meta list", "show"},
	"meta list":     {"meta get", "meta claims", "schema list"},
	"meta claims":   {"meta list", "meta get"},

	// Sync & remote
	"remote":      {"remote add", "remote ls", "push", "pull", "sync"},
	"remote add":  {"remote ls", "sync", "push"},
	"remote ls":   {"remote add", "remote rm", "status sync"},
	"remote rm":   {"remote ls"},
	"push":        {"pull", "sync", "remote ls", "status sync"},
	"pull":        {"push", "sync", "ingest"},
	"sync":        {"push", "pull", "status sync", "remote ls"},
	"ingest":      {"pull", "sync"},
	"status sync": {"sync", "remote ls", "push", "pull"},

	// Debugging & maintenance
	"debug":              {"debug doctor", "debug repair", "debug rebuild"},
	"debug doctor":       {"debug repair", "conflicts"},
	"debug repair":       {"debug doctor", "debug rebuild"},
	"debug rebuild":      {"debug repair", "ingest"},
	"debug events":       {"debug events list", "debug events show", "debug events stats", "log query"},
	"debug events list":  {"debug events show", "debug events stats"},
	"debug events show":  {"debug events list", "history"},
	"debug events stats": {"debug events list"},
	"debug node":         {"debug node show", "debug node regen", "remote ls"},
	"debug node show":    {"debug node regen"},
	"debug node regen":   {"debug node show"},
	"conflicts":          {"conflicts numbers", "debug doctor", "edit"},
	"conflicts numbers":  {"edit"},
	"history":            {"show", "log query", "debug events show"},
	"log":                {"log query", "log search"},
	"log query":          {"log search", "debug events list", "history"},
	"log search":         {"log query"},

	// Migration & setup
	"init":                    {"new", "project create", "remote add"},
	"migrate":                 {"migrate scan-deprecated", "debug doctor"},
	"migrate scan-deprecated": {"debug doctor"},
	"import-beads":            {"new", "ls", "ingest"},
	"version":                 {"debug doctor"},

	// Database
	"db":      {"db path"},
	"db path": {"init"},

	// Status
	"status":     {"status sync"},
	"statusline": {"status sync"},
}

// SeeAlso formats a "See Also" section for command help text.
// It takes a list of command paths and returns a formatted string
// that can be appended to a command's Long field.
func SeeAlso(commands ...string) string {
	if len(commands) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n\nSee Also:")
	for _, cmd := range commands {
		b.WriteString("\n  tk ")
		b.WriteString(cmd)
	}
	return b.String()
}

// ApplySeeAlso adds "See Also" sections to commands based on the registry.
// It recursively applies to all subcommands as well.
//
// This should be called once on the root command after all commands
// have been registered.
func ApplySeeAlso(cmd *cobra.Command) {
	// Get the command path without the root "tk" prefix
	cmdPath := cmd.CommandPath()
	if strings.HasPrefix(cmdPath, "tk ") {
		cmdPath = cmdPath[3:] // Remove "tk " prefix
	}

	// Apply "See Also" if this command is in the registry
	if related, ok := seeAlsoRegistry[cmdPath]; ok {
		cmd.Long = strings.TrimSpace(cmd.Long) + SeeAlso(related...)
	}

	// Recursively apply to subcommands
	for _, subcmd := range cmd.Commands() {
		ApplySeeAlso(subcmd)
	}
}

// ValidateSeeAlso checks that all commands referenced in the registry
// actually exist in the command tree.
//
// This function is intended for use in tests only, not at runtime.
// It uses Cobra's Find() method to verify command existence.
//
// Returns an error if any referenced command doesn't exist, or if
// a source command in the registry doesn't exist.
func ValidateSeeAlso(root *cobra.Command) error {
	var errors []string

	for cmdPath, relatedCmds := range seeAlsoRegistry {
		// Verify the source command exists
		sourceCmd, _, err := root.Find(strings.Fields(cmdPath))
		if err != nil || sourceCmd == nil {
			errors = append(errors, fmt.Sprintf("source command %q not found in command tree", cmdPath))
			continue
		}

		// Verify each related command exists
		for _, relPath := range relatedCmds {
			targetCmd, _, err := root.Find(strings.Fields(relPath))
			if err != nil || targetCmd == nil {
				errors = append(errors,
					fmt.Sprintf("command %q references non-existent command %q", cmdPath, relPath))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("See Also validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}
	return nil
}
