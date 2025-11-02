package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/segment"
	"github.com/neongreen/mono/tk/internal/sync"
	"github.com/spf13/cobra"
)

var debugEventsCmd = &cobra.Command{
	Use:   "debug-events [task-id]",
	Short: "Dump all raw events for a task ID (for debugging)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// Get the first remote
		var remote sync.RemoteConfig
		var remoteName string
		for name, r := range config.Remotes {
			remote = r
			remoteName = name
			break
		}

		if remoteName == "" {
			return fmt.Errorf("no remotes configured")
		}

		fmt.Printf("Searching remote '%s' for events related to task '%s'\n\n", remoteName, taskID)

		// Use configured spaces, or default to "personal"
		spaces := remote.Spaces
		if len(spaces) == 0 {
			spaces = []string{"personal"}
		}

		foundCount := 0
		for _, space := range spaces {
			// Find all segment files for this space
			segmentsDir := filepath.Join(remote.Path, space, "segments")
			if _, err := os.Stat(segmentsDir); os.IsNotExist(err) {
				fmt.Printf("No segments directory found for space '%s'\n", space)
				continue
			}

			segmentFiles, err := collectSegmentFiles(segmentsDir)
			if err != nil {
				return fmt.Errorf("failed to scan segments directory for space '%s': %w", space, err)
			}

			if len(segmentFiles) == 0 {
				fmt.Printf("No segment files found in space '%s'\n", space)
				continue
			}

			fmt.Printf("Scanning %d segment files in space '%s'...\n", len(segmentFiles), space)

			for _, segmentFile := range segmentFiles {
				reader := segment.NewSegmentReader(segmentFile)
				events, err := reader.ReadEvents()
				if err != nil {
					fmt.Printf("Warning: failed to read %s: %v\n", segmentFile, err)
					continue
				}

				for _, event := range events {
					// Convert payload to JSON bytes for searching
					payloadBytes, err := json.Marshal(event.Payload)
					if err != nil {
						continue
					}
					payloadStr := string(payloadBytes)

					// Check if this event references our task
					if containsTaskRef(payloadStr, taskID) {
						foundCount++
						fmt.Printf("=== types.Event #%d ===\n", foundCount)
						fmt.Printf("ID:       %s\n", event.ID)
						fmt.Printf("Lamport:  %d\n", event.Lamport)
						fmt.Printf("Time:     %s\n", event.TS)
						fmt.Printf("Actor:    %s\n", event.Actor)
						fmt.Printf("Role:     %s\n", event.Role)
						fmt.Printf("Kind:     %s\n", event.Kind)

						// Pretty-print the payload
						prettyJSON, _ := json.MarshalIndent(event.Payload, "  ", "  ")
						fmt.Printf("Payload:  %s\n", string(prettyJSON))

						if event.Ctx != nil {
							fmt.Printf("Context:\n")
							if event.Ctx.RepoUUID != nil {
								fmt.Printf("  RepoUUID: %s\n", *event.Ctx.RepoUUID)
							}
							if event.Ctx.Branch != nil {
								fmt.Printf("  Branch:   %s\n", *event.Ctx.Branch)
							}
							if event.Ctx.Commit != nil {
								fmt.Printf("  Commit:   %s\n", *event.Ctx.Commit)
							}
						}
						fmt.Println()
					}
				}
			}
		}

		fmt.Printf("\nTotal events found: %d\n", foundCount)
		return nil
	},
}

func containsTaskRef(payloadStr string, taskID string) bool {
	// Simple string search - just check if the task ID appears anywhere
	// This is intentionally naive for debugging purposes
	return len(payloadStr) > 0 && (strings.Contains(payloadStr, `"`+taskID+`"`) ||
		strings.Contains(payloadStr, taskID))
}

func init() {
	debugCmd.AddCommand(debugEventsCmd)
}
