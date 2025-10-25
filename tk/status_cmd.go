package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var statusSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Show sync status for all remotes",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig()
		if err != nil {
			return err
		}

		if len(config.Remotes) == 0 {
			fmt.Println("No remotes configured.")
			return nil
		}

		stateDir, err := GetStateDir()
		if err != nil {
			return err
		}

		for remoteName, remote := range config.Remotes {
			for _, space := range remote.Spaces {
				// Load local index mirror
				localIndexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
				localIndex, err := loadIndexFile(localIndexPath)
				localSegCount := 0
				if err == nil && localIndex != nil {
					localSegCount = len(localIndex.Segments)
				}

				// Load remote index
				remoteIndexPath := filepath.Join(remote.Path, space, "index.json")
				remoteIndex, err := loadIndexFile(remoteIndexPath)
				remoteSegCount := 0
				if err == nil && remoteIndex != nil {
					remoteSegCount = len(remoteIndex.Segments)
				}

				// Check for divergence
				diverged := "no"
				localOnly := 0
				remoteOnly := 0

				if localIndex != nil && remoteIndex != nil {
					localSegs := make(map[string]bool)
					for _, seg := range localIndex.Segments {
						localSegs[seg.Rel] = true
					}

					remoteSegs := make(map[string]bool)
					for _, seg := range remoteIndex.Segments {
						remoteSegs[seg.Rel] = true
						if !localSegs[seg.Rel] {
							remoteOnly++
						}
					}

					for _, seg := range localIndex.Segments {
						if !remoteSegs[seg.Rel] {
							localOnly++
						}
					}

					if localOnly > 0 || remoteOnly > 0 {
						diverged = "yes"
					}
				}

				// Get last sync time
				watermarkFile := filepath.Join(stateDir, "ingest_watermarks", remoteName, space+".json")
				lastSync := "never"
				if data, err := os.ReadFile(watermarkFile); err == nil {
					var watermark IngestWatermark
					if err := json.Unmarshal(data, &watermark); err == nil {
						elapsed := time.Since(watermark.UpdatedAt)
						lastSync = formatDuration(elapsed)
					}
				}

				fmt.Printf("%s/%s: local %d segs, remote %d segs, diverged: %s, last_sync: %s\n",
					remoteName, space, localSegCount, remoteSegCount, diverged, lastSync)
				if localOnly > 0 || remoteOnly > 0 {
					fmt.Printf("  local +%d seg, remote +%d seg\n", localOnly, remoteOnly)
				}
			}
		}

		return nil
	},
}

// formatDuration formats a duration in human-readable form
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
