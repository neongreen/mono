package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var (
	// Version information - set via -ldflags at build time
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

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
