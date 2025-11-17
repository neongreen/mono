package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"
)

var (
	// Version information - automatically populated from VCS
	Version   = "dev"
	GitCommit = "unknown" // Short commit hash (12 chars) for display
	BuildTime = "unknown"

	// Additional version metadata
	GitCommitFull = "unknown" // Full commit hash
	GitModified   = false     // Whether there are uncommitted changes
	CommitTime    = "unknown" // Commit timestamp (vcs.time)
)

func init() {
	// Try to read VCS information from build info
	// This works automatically when building in a git repository with plain `go build`
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	var revision, vcsTime, modified string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.time":
			vcsTime = setting.Value
		case "vcs.modified":
			modified = setting.Value
		}
	}

	// Populate GitCommit if we have revision info
	if revision != "" {
		GitCommitFull = revision // Store full commit hash
		// Use short commit hash (first 12 chars) for display
		if len(revision) > 12 {
			GitCommit = revision[:12]
		} else {
			GitCommit = revision
		}
	}

	// Populate BuildTime and CommitTime if we have VCS time info
	if vcsTime != "" {
		BuildTime = vcsTime
		CommitTime = vcsTime
	}

	// Store modified flag
	if modified == "true" {
		GitModified = true
	}

	// Append "-dirty" to version if there are uncommitted changes
	if GitModified && Version == "dev" {
		Version = "dev-dirty"
	}
}

// PrintVersion prints version information for the given tool name.
func PrintVersion(toolName string) {
	// Format build time as human-readable in local time
	buildTimeStr := BuildTime
	if t, err := time.Parse(time.RFC3339, BuildTime); err == nil {
		buildTimeStr = t.Local().Format("Jan 2, 2006 15:04 MST")
	}

	fmt.Printf("%s version %s\n", toolName, Version)
	fmt.Printf("  commit: %s\n", GitCommit)
	fmt.Printf("  built:  %s\n", buildTimeStr)
	fmt.Printf("  go:     %s\n", runtime.Version())
}

// NewVersionCommand creates a new version command for the given tool name.
// The tool name is used to customize the output (e.g., "tk version 1.0.0").
func NewVersionCommand(toolName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOutput, _ := cmd.Flags().GetBool("json")

			if jsonOutput {
				// Output as JSON with ISO timestamp
				versionInfo := map[string]string{
					"version":    Version,
					"commit":     GitCommit,
					"build_time": BuildTime,
					"go_version": runtime.Version(),
				}
				output, err := json.MarshalIndent(versionInfo, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(output))
			} else {
				// Format build time as human-readable in local time
				buildTimeStr := BuildTime
				if t, err := time.Parse(time.RFC3339, BuildTime); err == nil {
					buildTimeStr = t.Local().Format("Jan 2, 2006 15:04 MST")
				}

				fmt.Printf("%s version %s\n", toolName, Version)
				fmt.Printf("  commit: %s\n", GitCommit)
				fmt.Printf("  built:  %s\n", buildTimeStr)
				fmt.Printf("  go:     %s\n", runtime.Version())
			}
			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output version information as JSON")
	return cmd
}
