package cobralint

import (
	"golang.org/x/tools/go/analysis"
)

// RequireJSONFlagChecker ensures all commands have a --json flag.
type RequireJSONFlagChecker struct{}

func (c *RequireJSONFlagChecker) Name() string {
	return "require-json-flag"
}

func (c *RequireJSONFlagChecker) Check(pass *analysis.Pass, cmd *CommandInfo) {
	// Skip the root command as it typically doesn't have its own flags
	if cmd.IsRoot {
		return
	}

	// Skip if command has an exemption
	if cmd.ExemptFromJSONFlag != nil {
		return
	}

	// Check if the command has a --json flag
	hasJSONFlag := false
	for _, flag := range cmd.Flags {
		if flag.Name == "json" {
			hasJSONFlag = true
			break
		}
	}

	if !hasJSONFlag {
		pass.Reportf(cmd.Pos, "command %q (use: %q) missing required --json flag", cmd.Name, cmd.Use)
	}
}
