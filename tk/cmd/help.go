package cmd

// This file provides the "See Also" functionality for TK commands.
//
// The registry of command relationships lives in help_registry.go - if you're
// adding a new command, update that file to include cross-references.
//
// Two formatters are available:
// - SeeAlso(): Simple list of command names (current default)
// - SeeAlsoWithDescriptions(): Includes descriptions from Cobra's Short field
//
// To switch to showing descriptions, change ApplySeeAlso to call
// SeeAlsoWithDescriptions(RootCmd, related...) instead of SeeAlso(related...)

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// SeeAlso formats a "See Also" section for command help text.
// It takes a list of command paths and returns a formatted string
// that can be appended to a command's Long field.
//
// By default, this only shows command names. To include descriptions from
// Cobra commands, use SeeAlsoWithDescriptions instead.
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

// SeeAlsoWithDescriptions formats a "See Also" section with descriptions
// pulled from the Cobra command's Short field.
//
// This avoids duplication by reusing the descriptions already defined
// in each command. Commands that don't exist or have no Short description
// will be shown without a description.
func SeeAlsoWithDescriptions(root *cobra.Command, commands ...string) string {
	if len(commands) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n\nSee Also:")

	for _, cmdPath := range commands {
		b.WriteString("\n  tk ")
		b.WriteString(cmdPath)

		// Try to find the command and get its Short description
		cmd, _, err := root.Find(strings.Fields(cmdPath))
		if err == nil && cmd != nil && cmd.Short != "" {
			// Pad command name for alignment (15 chars should fit most commands)
			padding := 20 - len(cmdPath)
			if padding > 0 {
				b.WriteString(strings.Repeat(" ", padding))
			}
			b.WriteString(" - ")
			b.WriteString(cmd.Short)
		}
	}
	return b.String()
}

// ApplySeeAlso adds "See Also" sections to commands based on the registry.
// It recursively applies to all subcommands as well.
//
// This should be called once on the root command after all commands
// have been registered.
//
// By default, this shows just command names. To include descriptions from
// each command's Short field, change the function call from:
//
//	SeeAlso(related...)
//
// to:
//
//	SeeAlsoWithDescriptions(cmd.Root(), related...)
func ApplySeeAlso(cmd *cobra.Command) {
	// Get the command path without the root "tk" prefix
	cmdPath := strings.TrimPrefix(cmd.CommandPath(), "tk ")

	// Apply "See Also" if this command is in the registry
	if related, ok := seeAlsoRegistry[cmdPath]; ok {
		// Current: Simple command names only
		cmd.Long = strings.TrimSpace(cmd.Long) + SeeAlso(related...)

		// Alternative: Include descriptions from Cobra's Short field
		// cmd.Long = strings.TrimSpace(cmd.Long) + SeeAlsoWithDescriptions(cmd.Root(), related...)
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
