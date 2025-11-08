package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/segment"
	"github.com/neongreen/mono/tk/internal/types"
)

// ExportParams contains parameters for the Export function
type ExportParams struct {
	RemoteName   string
	RemoteConfig config.RemoteConfig
	Space        string
	ExportAll    bool
	StateDir     string
	SyncConfig   config.SyncConfig
}

// ExportResult contains the results of an export operation
type ExportResult struct {
	EventsExported  int
	SegmentsWritten int
	SegmentFiles    []SegmentInfo
}

// Export exports local events to segment files for a remote
func Export(db *database.DB, params ExportParams) (*ExportResult, error) {
	// Get node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return nil, fmt.Errorf("failed to get node ID: %w", err)
	}

	// Get all events
	events, err := db.GetEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// Get or create export state
	exportStateFile := filepath.Join(params.StateDir, "export", params.RemoteName, params.Space+".json")
	exportState, err := LoadExportState(exportStateFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load export state: %w", err)
	}
	if exportState == nil {
		exportState = &ExportState{
			RemoteName:          params.RemoteName,
			Space:               params.Space,
			LastExportedEventID: "",
			SegmentSeq:          0,
		}
	}

	// Filter events to export
	eventsToExport := filterEventsToExport(events, exportState.LastExportedEventID, params.ExportAll)
	if len(eventsToExport) == 0 {
		return &ExportResult{}, nil
	}

	// Convert to segment events
	segmentEvents := make([]SegmentEvent, 0, len(eventsToExport))
	for _, e := range eventsToExport {
		segEvent, err := EventToSegmentEvent(e, params.Space, nodeID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event %s: %w", e.ID, err)
		}
		segmentEvents = append(segmentEvents, segEvent)
	}

	// Write segments
	segmentsWritten, err := writeSegments(
		params.RemoteConfig.Path,
		params.Space,
		nodeID,
		exportState.SegmentSeq,
		params.SyncConfig,
		segmentEvents,
	)
	if err != nil {
		return nil, err
	}

	// Update export state
	if len(eventsToExport) > 0 {
		exportState.LastExportedEventID = eventsToExport[len(eventsToExport)-1].ID
		exportState.SegmentSeq = exportState.SegmentSeq + int64(len(segmentsWritten))
		exportState.UpdatedAt = time.Now()

		if err := SaveExportState(exportStateFile, exportState); err != nil {
			return nil, fmt.Errorf("failed to save export state: %w", err)
		}
	}

	// Update local index mirror
	indexPath := filepath.Join(params.StateDir, "remotes", params.RemoteName, params.Space, "index.json")
	if err := UpdateLocalIndex(indexPath, segmentsWritten); err != nil {
		return nil, fmt.Errorf("failed to update local index: %w", err)
	}

	return &ExportResult{
		EventsExported:  len(eventsToExport),
		SegmentsWritten: len(segmentsWritten),
		SegmentFiles:    segmentsWritten,
	}, nil
}

// filterEventsToExport filters events based on the last exported event ID
func filterEventsToExport(events []types.Event, lastExportedEventID string, exportAll bool) []types.Event {
	if exportAll {
		return events
	}

	var eventsToExport []types.Event
	foundLast := lastExportedEventID == ""
	for _, e := range events {
		if foundLast {
			eventsToExport = append(eventsToExport, e)
		} else if e.ID == lastExportedEventID {
			foundLast = true
		}
	}
	return eventsToExport
}

// writeSegments writes events to segment files with rotation
func writeSegments(
	remotePath string,
	space string,
	nodeID string,
	startSegmentSeq int64,
	syncConfig config.SyncConfig,
	segmentEvents []SegmentEvent,
) ([]SegmentInfo, error) {
	segmentSeq := startSegmentSeq + 1
	writer := segment.NewSegmentWriter(
		remotePath,
		space,
		nodeID,
		segmentSeq,
		syncConfig.SegmentMaxBytes,
		syncConfig.SegmentMaxAge,
	)

	var segmentsWritten []SegmentInfo
	for _, segEvent := range segmentEvents {
		writer.AddEvent(segEvent)

		if writer.ShouldRotate() {
			segInfo, err := writer.WriteSegment()
			if err != nil {
				return nil, fmt.Errorf("failed to write segment: %w", err)
			}
			if segInfo != nil {
				segmentsWritten = append(segmentsWritten, *segInfo)
			}

			// Start new segment
			segmentSeq++
			writer = segment.NewSegmentWriter(
				remotePath,
				space,
				nodeID,
				segmentSeq,
				syncConfig.SegmentMaxBytes,
				syncConfig.SegmentMaxAge,
			)
		}
	}

	// Write any remaining events
	if writer.HasPendingEvents() {
		segInfo, err := writer.WriteSegment()
		if err != nil {
			return nil, fmt.Errorf("failed to write final segment: %w", err)
		}
		if segInfo != nil {
			segmentsWritten = append(segmentsWritten, *segInfo)
		}
	}

	return segmentsWritten, nil
}

// LoadExportState loads the export state from a file
func LoadExportState(path string) (*ExportState, error) {
	return LoadJSON[ExportState](path)
}

// SaveExportState saves the export state to a file
func SaveExportState(path string, state *ExportState) error {
	return SaveJSON(path, state)
}
