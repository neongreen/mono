package remote

import (
	"fmt"
	"log/slog"
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
	log := slog.With("remote", params.RemoteName, "space", params.Space)
	log.Debug("export: starting", "export_all", params.ExportAll)

	// Get node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return nil, fmt.Errorf("failed to get node ID: %w", err)
	}
	log.Debug("export: got node ID", "node_id", nodeID)

	// Get all events
	events, err := db.GetEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	log.Debug("export: got events from database", "total_events", len(events))

	// Get or create export state
	exportStateFile := filepath.Join(params.StateDir, "export", params.RemoteName, params.Space+".json")
	log.Debug("export: loading export state", "path", exportStateFile)
	exportState, err := LoadExportState(exportStateFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load export state: %w", err)
	}
	if exportState == nil {
		log.Debug("export: no export state found, creating new state")
		exportState = &ExportState{
			RemoteName:          params.RemoteName,
			Space:               params.Space,
			LastExportedEventID: "",
			SegmentSeq:          0,
		}
	} else {
		log.Debug("export: loaded export state", "last_exported_id", exportState.LastExportedEventID, "segment_seq", exportState.SegmentSeq)
	}

	// Filter events to export
	eventsToExport := filterEventsToExport(events, exportState.LastExportedEventID, params.ExportAll)
	log.Info("export: filtered events to export", "total_events", len(events), "to_export", len(eventsToExport))
	if len(eventsToExport) == 0 {
		log.Debug("export: no events to export")
		return &ExportResult{}, nil
	}

	// Convert to segment events
	log.Debug("export: converting events to segment format")
	segmentEvents := make([]SegmentEvent, 0, len(eventsToExport))
	for _, e := range eventsToExport {
		segEvent, err := EventToSegmentEvent(e, params.Space, nodeID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event %s: %w", e.ID, err)
		}
		segmentEvents = append(segmentEvents, segEvent)
	}

	// Write segments
	log.Debug("export: writing segments", "remote_path", params.RemoteConfig.Path, "start_segment_seq", exportState.SegmentSeq)
	segmentsWritten, err := writeSegments(
		log,
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
	log.Info("export: segments written", "count", len(segmentsWritten))

	// Update export state
	if len(eventsToExport) > 0 {
		exportState.LastExportedEventID = eventsToExport[len(eventsToExport)-1].ID
		exportState.SegmentSeq = exportState.SegmentSeq + int64(len(segmentsWritten))
		exportState.UpdatedAt = time.Now()

		log.Debug("export: saving export state", "path", exportStateFile, "last_exported_id", exportState.LastExportedEventID, "segment_seq", exportState.SegmentSeq)
		if err := SaveExportState(exportStateFile, exportState); err != nil {
			return nil, fmt.Errorf("failed to save export state: %w", err)
		}
	}

	// Update local index mirror
	indexPath := filepath.Join(params.StateDir, "remotes", params.RemoteName, params.Space, "index.json")
	log.Debug("export: updating local index", "path", indexPath, "new_segments", len(segmentsWritten))
	if err := UpdateLocalIndex(indexPath, segmentsWritten); err != nil {
		return nil, fmt.Errorf("failed to update local index: %w", err)
	}

	// Cache segment files under state directory for restoration
	for _, seg := range segmentsWritten {
		if err := CacheSegmentFile(params.StateDir, params.RemoteName, params.RemoteConfig.Path, seg); err != nil {
			return nil, err
		}
	}

	log.Info("export: completed", "events_exported", len(eventsToExport), "segments_written", len(segmentsWritten))
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
	log *slog.Logger,
	remotePath string,
	space string,
	nodeID string,
	startSegmentSeq int64,
	syncConfig config.SyncConfig,
	segmentEvents []SegmentEvent,
) ([]SegmentInfo, error) {
	segmentSeq := startSegmentSeq + 1
	log.Debug("export: writeSegments starting", "remote_path", remotePath, "node_id", nodeID, "segment_seq", segmentSeq, "event_count", len(segmentEvents))
	writer := segment.NewSegmentWriter(
		remotePath,
		space,
		nodeID,
		segmentSeq,
		syncConfig.SegmentMaxBytes,
		syncConfig.SegmentMaxAge,
	)

	var segmentsWritten []SegmentInfo
	for i, segEvent := range segmentEvents {
		writer.AddEvent(segEvent)

		if writer.ShouldRotate() {
			log.Debug("export: rotating segment", "event_index", i, "segment_seq", segmentSeq)
			segInfo, err := writer.WriteSegment()
			if err != nil {
				return nil, fmt.Errorf("failed to write segment: %w", err)
			}
			if segInfo != nil {
				log.Debug("export: segment written", "path", segInfo.Rel, "size", segInfo.Size, "sha256", segInfo.SHA256)
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
		log.Debug("export: writing final segment", "segment_seq", segmentSeq)
		segInfo, err := writer.WriteSegment()
		if err != nil {
			return nil, fmt.Errorf("failed to write final segment: %w", err)
		}
		if segInfo != nil {
			log.Debug("export: final segment written", "path", segInfo.Rel, "size", segInfo.Size, "sha256", segInfo.SHA256)
			segmentsWritten = append(segmentsWritten, *segInfo)
		}
	}

	log.Debug("export: writeSegments completed", "segments_written", len(segmentsWritten))
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
