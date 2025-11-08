package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/sync"
	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Show sync status for all remotes",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		if len(config.Remotes) == 0 {
			if jsonOutput {
				fmt.Println("[]")
			} else {
				fmt.Println("No remotes configured.")
			}
			return nil
		}

		stateDir, err := config_pkg.GetStateDir()
		if err != nil {
			return err
		}

		type SyncStatusOutput struct {
			Remote     string `json:"remote"`
			Space      string `json:"space"`
			LocalSegs  int    `json:"local_segments"`
			RemoteSegs int    `json:"remote_segments"`
			Diverged   bool   `json:"diverged"`
			LocalOnly  int    `json:"local_only_segments"`
			RemoteOnly int    `json:"remote_only_segments"`
			LastSync   string `json:"last_sync"`
		}

		var statuses []SyncStatusOutput

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
				}

				// Get last sync time
				watermarkFile := filepath.Join(stateDir, "ingest_watermarks", remoteName, space+".json")
				lastSync := "never"
				if data, err := os.ReadFile(watermarkFile); err == nil {
					var watermark sync.IngestWatermark
					if err := json.Unmarshal(data, &watermark); err == nil {
						elapsed := time.Since(watermark.UpdatedAt)
						lastSync = formatDuration(elapsed)
					}
				}

				statuses = append(statuses, SyncStatusOutput{
					Remote:     remoteName,
					Space:      space,
					LocalSegs:  localSegCount,
					RemoteSegs: remoteSegCount,
					Diverged:   localOnly > 0 || remoteOnly > 0,
					LocalOnly:  localOnly,
					RemoteOnly: remoteOnly,
					LastSync:   lastSync,
				})
			}
		}

		if jsonOutput {
			output, err := json.MarshalIndent(statuses, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal status: %w", err)
			}
			fmt.Println(string(output))
		} else {
			for _, status := range statuses {
				diverged := "no"
				if status.Diverged {
					diverged = "yes"
				}
				fmt.Printf("%s/%s: local %d segs, remote %d segs, diverged: %s, last_sync: %s\n",
					status.Remote, status.Space, status.LocalSegs, status.RemoteSegs, diverged, status.LastSync)
				if status.LocalOnly > 0 || status.RemoteOnly > 0 {
					fmt.Printf("  local +%d seg, remote +%d seg\n", status.LocalOnly, status.RemoteOnly)
				}
			}
		}

		return nil
	},
}

func init() {
	SyncCmd.Flags().Bool("json", false, "Output as JSON")
}

// loadIndexFile loads an index.json file
func loadIndexFile(path string) (*sync.IndexFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var index sync.IndexFile
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	return &index, nil
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
