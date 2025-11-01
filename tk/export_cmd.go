package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/neongreen/mono/tk/internal/segment"
	"github.com/neongreen/mono/tk/internal/sync"
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

		db, err := OpenExistingDB()
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
		config, err := LoadConfig()
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
		remote, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found", remoteName)
		}

		// Get all events
		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		// Get or create export state
		stateDir, err := GetStateDir()
		if err != nil {
			return err
		}

		exportStateFile := filepath.Join(stateDir, "export", remoteName, space+".json")
		exportState, err := loadExportState(exportStateFile)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if exportState == nil {
			exportState = &sync.ExportState{
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
		segmentEvents := make([]sync.SegmentEvent, 0, len(eventsToExport))
		for _, e := range eventsToExport {
			segEvent, err := eventToSegmentEvent(e, space, nodeID)
			if err != nil {
				return fmt.Errorf("failed to convert event %s: %w", e.ID, err)
			}
			segmentEvents = append(segmentEvents, segEvent)
		}

		// Write segments
		segmentSeq := exportState.SegmentSeq + 1
		writer := segment.NewSegmentWriter(
			remote.Path,
			space,
			nodeID,
			segmentSeq,
			config.Sync.SegmentMaxBytes,
			config.Sync.SegmentMaxAge,
		)

		var segmentsWritten []sync.SegmentInfo
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
					remote.Path,
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

			if err := saveExportState(exportStateFile, exportState); err != nil {
				return err
			}
		}

		// Update local index mirror
		indexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
		if err := updateLocalIndex(indexPath, segmentsWritten); err != nil {
			return fmt.Errorf("failed to update local index: %w", err)
		}

		fmt.Printf("Exported %d events in %d segments\n", len(eventsToExport), len(segmentsWritten))
		return nil
	},
}

// eventToSegmentEvent converts an types.Event to a SegmentEvent
func eventToSegmentEvent(e types.Event, space, nodeID string) (sync.SegmentEvent, error) {
	// Parse payload to the right type
	var payload interface{}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return sync.SegmentEvent{}, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	ctx := &sync.SegmentContext{}
	if e.RepoUUID != "" {
		ctx.RepoUUID = &e.RepoUUID
	}
	if e.Branch != "" {
		ctx.Branch = &e.Branch
	}
	if e.Commit != "" {
		ctx.Commit = &e.Commit
	}
	if e.JJOpID != "" {
		ctx.JJOpID = &e.JJOpID
	}

	return sync.SegmentEvent{
		Schema:  "tk.event.v1",
		ID:      e.ID,
		Lamport: e.TS,
		TS:      e.CreatedAt.UTC().Format(time.RFC3339Nano),
		Node:    nodeID,
		Space:   space,
		Actor:   e.Actor,
		Role:    e.Role,
		Kind:    e.Kind,
		Payload: payload,
		Ctx:     ctx,
	}, nil
}

// loadExportState loads the export state from a file
func loadExportState(path string) (*sync.ExportState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state sync.ExportState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal export state: %w", err)
	}

	return &state, nil
}

// saveExportState saves the export state to a file
func saveExportState(path string, state *sync.ExportState) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create export state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write export state: %w", err)
	}

	return nil
}

// updateLocalIndex updates the local index mirror with new segments
func updateLocalIndex(indexPath string, newSegments []sync.SegmentInfo) error {
	// Ensure directory exists
	dir := filepath.Dir(indexPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	// Load existing index or create new one
	var index sync.IndexFile
	if data, err := os.ReadFile(indexPath); err == nil {
		if err := json.Unmarshal(data, &index); err != nil {
			return fmt.Errorf("failed to unmarshal index: %w", err)
		}
	} else {
		// Extract space from path
		parts := filepath.SplitList(indexPath)
		space := "personal" // default
		if len(parts) >= 2 {
			space = parts[len(parts)-2]
		}
		index = sync.IndexFile{
			Schema:   "tk.index.v1",
			Space:    space,
			Segments: []sync.SegmentInfo{},
		}
	}

	// Add new segments
	index.Segments = append(index.Segments, newSegments...)

	// Save index
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

func init() {
	exportCmd.Flags().String("space", "personal", "Space to export")
	exportCmd.Flags().Bool("all", false, "Export all events (not just unsent)")
}
