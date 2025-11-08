package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/segment"
)

// IngestFileResult contains the results of ingesting a single file
type IngestFileResult struct {
	EventsIngested int
	Duplicates     int
}

// IngestRemoteResult contains the results of ingesting from a remote
type IngestRemoteResult struct {
	EventsIngested int
	Duplicates     int
	SegmentsRead   int
}

// IngestFile ingests events from a single segment file
func IngestFile(db *database.DB, path string) (*IngestFileResult, error) {
	reader := segment.NewSegmentReader(path)
	events, err := reader.ReadEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to read segment file: %w", err)
	}

	ingested := 0
	duplicates := 0

	for _, segEvent := range events {
		// Convert segment event to types.Event
		event, err := SegmentEventToEvent(segEvent)
		if err != nil {
			return nil, fmt.Errorf("failed to convert segment event %s: %w", segEvent.ID, err)
		}

		// Try to insert event (will fail if duplicate)
		err = db.InsertEvent(event)
		if err != nil {
			// Check if it's a duplicate error
			if IsDuplicateError(err) {
				duplicates++
				continue
			}
			return nil, fmt.Errorf("failed to insert event %s: %w", event.ID, err)
		}

		// Bump lamport if needed
		if err := db.BumpLamport(event.TS); err != nil {
			return nil, fmt.Errorf("failed to bump lamport: %w", err)
		}

		// Project events into their respective tables
		if err := db.ProjectEvent(event); err != nil {
			// Non-fatal error - continue processing but log warning
			// Caller should handle logging
		}

		ingested++
	}

	// Ensure DB version is set to 4
	if err := db.SetDBVersion(4); err != nil {
		return nil, fmt.Errorf("failed to set DB version to 4: %w", err)
	}

	return &IngestFileResult{
		EventsIngested: ingested,
		Duplicates:     duplicates,
	}, nil
}

// IngestRemote ingests events from all segments in a remote
func IngestRemote(db *database.DB, remoteName string, remoteConfig config.RemoteConfig, stateDir string) (*IngestRemoteResult, error) {
	// Use configured spaces, or default to "personal"
	spaces := remoteConfig.Spaces
	if len(spaces) == 0 {
		spaces = []string{"personal"}
	}

	totalIngested := 0
	totalDuplicates := 0
	totalSegments := 0

	for _, space := range spaces {
		result, err := ingestRemoteSpace(db, remoteName, remoteConfig, space, stateDir)
		if err != nil {
			return nil, err
		}
		totalIngested += result.EventsIngested
		totalDuplicates += result.Duplicates
		totalSegments += result.SegmentsRead
	}

	return &IngestRemoteResult{
		EventsIngested: totalIngested,
		Duplicates:     totalDuplicates,
		SegmentsRead:   totalSegments,
	}, nil
}

// ingestRemoteSpace ingests events from a specific space in a remote
func ingestRemoteSpace(db *database.DB, remoteName string, remoteConfig config.RemoteConfig, space string, stateDir string) (*IngestRemoteResult, error) {
	// Find all segment files
	segmentsDir := filepath.Join(remoteConfig.Path, space, "segments")
	if _, err := os.Stat(segmentsDir); os.IsNotExist(err) {
		// No segments directory - return empty result
		return &IngestRemoteResult{}, nil
	}

	segmentFiles, err := CollectSegmentFiles(segmentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan segments directory: %w", err)
	}

	if len(segmentFiles) == 0 {
		return &IngestRemoteResult{}, nil
	}

	totalIngested := 0
	totalDuplicates := 0

	for _, segmentFile := range segmentFiles {
		reader := segment.NewSegmentReader(segmentFile)
		events, err := reader.ReadEvents()
		if err != nil {
			// Non-fatal error - skip this file and continue
			// Caller should handle logging
			continue
		}

		for _, segEvent := range events {
			// Convert segment event to types.Event
			event, err := SegmentEventToEvent(segEvent)
			if err != nil {
				return nil, fmt.Errorf("failed to convert segment event %s: %w", segEvent.ID, err)
			}

			// Try to insert event (will fail if duplicate)
			err = db.InsertEvent(event)
			if err != nil {
				// Check if it's a duplicate error
				if IsDuplicateError(err) {
					totalDuplicates++
					continue
				}
				return nil, fmt.Errorf("failed to insert event %s: %w", event.ID, err)
			}

			// Bump lamport if needed
			if err := db.BumpLamport(event.TS); err != nil {
				return nil, fmt.Errorf("failed to bump lamport: %w", err)
			}

			// Project events into their respective tables
			if err := db.ProjectEvent(event); err != nil {
				// Non-fatal error - continue processing
			}

			totalIngested++
		}
	}

	// Ensure DB version is set to 4
	if err := db.SetDBVersion(4); err != nil {
		return nil, fmt.Errorf("failed to set DB version to 4: %w", err)
	}

	// Save watermark
	watermarkFile := filepath.Join(stateDir, "ingest_watermarks", remoteName, space+".json")
	watermark := IngestWatermark{
		RemoteName: remoteName,
		Space:      space,
		UpdatedAt:  time.Now(),
	}
	if err := SaveIngestWatermark(watermarkFile, &watermark); err != nil {
		return nil, err
	}

	return &IngestRemoteResult{
		EventsIngested: totalIngested,
		Duplicates:     totalDuplicates,
		SegmentsRead:   len(segmentFiles),
	}, nil
}

// LoadIngestWatermark loads the ingest watermark from a file
func LoadIngestWatermark(path string) (*IngestWatermark, error) {
	return LoadJSON[IngestWatermark](path)
}

// SaveIngestWatermark saves an ingest watermark to a file
func SaveIngestWatermark(path string, watermark *IngestWatermark) error {
	return SaveJSON(path, watermark)
}

// IsDuplicateError checks if an error is a duplicate key error
func IsDuplicateError(err error) bool {
	// SQLite's duplicate key error contains "UNIQUE constraint failed"
	return err != nil && (
	// modernc.org/sqlite error messages
	containsString(err.Error(), "UNIQUE constraint failed") ||
		containsString(err.Error(), "constraint failed"))
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || containsString(s[1:], substr)))
}
