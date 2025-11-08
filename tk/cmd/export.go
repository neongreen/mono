package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/remote"
	"github.com/neongreen/mono/tk/internal/segment"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [remote-name]",
	Short: "Export local events to segment files",
	Long: `Export unsent local events to segment files.

Examples:
  tk export icloud               # Export to icloud remote
  tk export icloud --space personal --all  # Export all events
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space, _ := cmd.Flags().GetString("space")
		exportAll, _ := cmd.Flags().GetBool("all")

		// If no remote name provided, use default from config
		var remoteName string
		if len(args) > 0 {
			remoteName = args[0]
		}

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Get node ID
		nodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// If no remote name and no default, error
		if remoteName == "" {
			// Try to find a default remote
			if len(config.Remotes) == 0 {
				return fmt.Errorf("no remotes configured; add one with 'tk remote add'")
			}
			if len(config.Remotes) == 1 {
				// Use the only remote
				for name := range config.Remotes {
					remoteName = name
				}
			} else {
				return fmt.Errorf("multiple remotes configured; please specify which one to use")
			}
		}

		// Get remote config
		remoteConfig, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found", remoteName)
		}

		// Get all events
		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		// Get or create export state
		stateDir, err := config_pkg.GetStateDir()
		if err != nil {
			return err
		}

		exportStateFile := filepath.Join(stateDir, "export", remoteName, space+".json")
		exportState, err := remote.LoadExportState(exportStateFile)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if exportState == nil {
			exportState = &remote.ExportState{
				RemoteName:          remoteName,
				Space:               space,
				LastExportedEventID: "",
				SegmentSeq:          0,
			}
		}

		// Filter events to export
		var eventsToExport []types.Event
		if exportAll {
			eventsToExport = events
		} else {
			// Export only events after the last exported event
			foundLast := exportState.LastExportedEventID == ""
			for _, e := range events {
				if foundLast {
					eventsToExport = append(eventsToExport, e)
				} else if e.ID == exportState.LastExportedEventID {
					foundLast = true
				}
			}
		}

		if len(eventsToExport) == 0 {
			fmt.Println("No new events to export")
			return nil
		}

		// Convert to segment events
		segmentEvents := make([]remote.SegmentEvent, 0, len(eventsToExport))
		for _, e := range eventsToExport {
			segEvent, err := remote.EventToSegmentEvent(e, space, nodeID)
			if err != nil {
				return fmt.Errorf("failed to convert event %s: %w", e.ID, err)
			}
			segmentEvents = append(segmentEvents, segEvent)
		}

		// Write segments
		segmentSeq := exportState.SegmentSeq + 1
		writer := segment.NewSegmentWriter(
			remoteConfig.Path,
			space,
			nodeID,
			segmentSeq,
			config.Sync.SegmentMaxBytes,
			config.Sync.SegmentMaxAge,
		)

		var segmentsWritten []remote.SegmentInfo
		for _, segEvent := range segmentEvents {
			writer.AddEvent(segEvent)

			if writer.ShouldRotate() {
				segInfo, err := writer.WriteSegment()
				if err != nil {
					return fmt.Errorf("failed to write segment: %w", err)
				}
				if segInfo != nil {
					segmentsWritten = append(segmentsWritten, *segInfo)
					fmt.Printf("Wrote segment: %s\n", segInfo.Rel)
				}

				// Start new segment
				segmentSeq++
				writer = segment.NewSegmentWriter(
					remoteConfig.Path,
					space,
					nodeID,
					segmentSeq,
					config.Sync.SegmentMaxBytes,
					config.Sync.SegmentMaxAge,
				)
			}
		}

		// Write any remaining events
		if writer.HasPendingEvents() {
			segInfo, err := writer.WriteSegment()
			if err != nil {
				return fmt.Errorf("failed to write final segment: %w", err)
			}
			if segInfo != nil {
				segmentsWritten = append(segmentsWritten, *segInfo)
				fmt.Printf("Wrote segment: %s\n", segInfo.Rel)
			}
		}

		// Update export state
		if len(eventsToExport) > 0 {
			exportState.LastExportedEventID = eventsToExport[len(eventsToExport)-1].ID
			exportState.SegmentSeq = segmentSeq
			exportState.UpdatedAt = time.Now()

			if err := remote.SaveExportState(exportStateFile, exportState); err != nil {
				return err
			}
		}

		// Update local index mirror
		indexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
		if err := remote.UpdateLocalIndex(indexPath, segmentsWritten); err != nil {
			return fmt.Errorf("failed to update local index: %w", err)
		}

		fmt.Printf("Exported %d events in %d segments\n", len(eventsToExport), len(segmentsWritten))
		return nil
	},
}

func init() {
	exportCmd.Flags().String("space", "personal", "Space to export")
	exportCmd.Flags().Bool("all", false, "Export all events (not just unsent)")
}
