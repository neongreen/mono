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

var ingestCmd = &cobra.Command{
	Use:   "ingest [path-or-remote]",
	Short: "Ingest events from segment files",
	Long: `Ingest events from segment files into the local database.

Examples:
  tk ingest /path/to/segment.jsonl.zst    # Ingest from a single file
  tk ingest icloud                         # Ingest from remote's segments
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("path or remote name required")
		}

		pathOrRemote := args[0]

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check if it's a file or a remote name
		var ingestErr error
		if _, err := os.Stat(pathOrRemote); err == nil {
			// It's a file
			ingestErr = ingestFile(db, pathOrRemote)
		} else {
			// Try as a remote name
			config, err := config_pkg.LoadConfig()
			if err != nil {
				return err
			}

			remoteConfig, exists := config.Remotes[pathOrRemote]
			if !exists {
				return fmt.Errorf("'%s' is neither a file nor a configured remote", pathOrRemote)
			}

			ingestErr = IngestRemote(db, pathOrRemote, remoteConfig)
		}

		return ingestErr
	},
}

// ingestFile ingests events from a single segment file
func ingestFile(db *database.DB, path string) error {
	reader := segment.NewSegmentReader(path)
	events, err := reader.ReadEvents()
	if err != nil {
		return fmt.Errorf("failed to read segment file: %w", err)
	}

	ingested := 0
	duplicates := 0

	for _, segEvent := range events {
		// Convert segment event to types.Event
		event, err := remote.SegmentEventToEvent(segEvent)
		if err != nil {
			return fmt.Errorf("failed to convert segment event %s: %w", segEvent.ID, err)
		}

		// Try to insert event (will fail if duplicate)
		err = db.InsertEvent(event)
		if err != nil {
			// Check if it's a duplicate error
			if remote.IsDuplicateError(err) {
				duplicates++
				continue
			}
			return fmt.Errorf("failed to insert event %s: %w", event.ID, err)
		}

		// Bump lamport if needed
		if err := db.BumpLamport(event.TS); err != nil {
			return fmt.Errorf("failed to bump lamport: %w", err)
		}

		// Project events into their respective tables
		if err := projectEvent(db, event); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to project event %s: %v\n", event.ID, err)
		}

		ingested++
	}

	// Ensure DB version is set to 4
	if err := db.SetDBVersion(4); err != nil {
		return fmt.Errorf("failed to set DB version to 4: %w", err)
	}

	fmt.Printf("Ingested %d events (%d duplicates skipped)\n", ingested, duplicates)
	return nil
}

// IngestRemote ingests events from all segments in a remote
func IngestRemote(db *database.DB, remoteName string, remoteConfig remote.RemoteConfig) error {
	// Use configured spaces, or default to "personal"
	spaces := remoteConfig.Spaces
	if len(spaces) == 0 {
		spaces = []string{"personal"}
	}

	for _, space := range spaces {
		if err := ingestRemoteSpace(db, remoteName, remoteConfig, space); err != nil {
			return err
		}
	}

	return nil
}

// ingestRemoteSpace ingests events from a specific space in a remote
func ingestRemoteSpace(db *database.DB, remoteName string, remoteConfig remote.RemoteConfig, space string) error {
	// Find all segment files
	segmentsDir := filepath.Join(remoteConfig.Path, space, "segments")
	if _, err := os.Stat(segmentsDir); os.IsNotExist(err) {
		fmt.Printf("No segments directory found for space '%s'\n", space)
		return nil
	}

	segmentFiles, err := remote.CollectSegmentFiles(segmentsDir)
	if err != nil {
		return fmt.Errorf("failed to scan segments directory: %w", err)
	}

	if len(segmentFiles) == 0 {
		fmt.Println("No segment files found")
		return nil
	}

	totalIngested := 0
	totalDuplicates := 0

	for _, segmentFile := range segmentFiles {
		reader := segment.NewSegmentReader(segmentFile)
		events, err := reader.ReadEvents()
		if err != nil {
			fmt.Printf("Warning: failed to read %s: %v\n", segmentFile, err)
			continue
		}

		for _, segEvent := range events {
			// Convert segment event to types.Event
			event, err := remote.SegmentEventToEvent(segEvent)
			if err != nil {
				return fmt.Errorf("failed to convert segment event %s: %w", segEvent.ID, err)
			}

			// Try to insert event (will fail if duplicate)
			err = db.InsertEvent(event)
			if err != nil {
				// Check if it's a duplicate error
				if remote.IsDuplicateError(err) {
					totalDuplicates++
					continue
				}
				return fmt.Errorf("failed to insert event %s: %w", event.ID, err)
			}

			// Bump lamport if needed
			if err := db.BumpLamport(event.TS); err != nil {
				return fmt.Errorf("failed to bump lamport: %w", err)
			}

			// Project events into their respective tables
			if err := projectEvent(db, event); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to project event %s: %v\n", event.ID, err)
			}

			totalIngested++
		}
	}

	// Ensure DB version is set to 4
	if err := db.SetDBVersion(4); err != nil {
		return fmt.Errorf("failed to set DB version to 4: %w", err)
	}

	// Save watermark
	stateDir, err := config_pkg.GetStateDir()
	if err != nil {
		return err
	}

	watermarkFile := filepath.Join(stateDir, "ingest_watermarks", remoteName, space+".json")
	watermark := remote.IngestWatermark{
		RemoteName: remoteName,
		Space:      space,
		UpdatedAt:  time.Now(),
	}
	if err := remote.SaveIngestWatermark(watermarkFile, &watermark); err != nil {
		return err
	}

	fmt.Printf("Ingested %d events from %d segments (%d duplicates skipped)\n",
		totalIngested, len(segmentFiles), totalDuplicates)
	return nil
}

// projectEvent projects an event into its respective table
func projectEvent(db *database.DB, event types.Event) error {
	switch event.Kind {
	case string(types.EventKindProjectCreated):
		return db.ProjectProjectCreatedEvent(event)
	case string(types.EventKindProjectAliasAdd):
		return db.ProjectProjectAliasAddEvent(event)
	case string(types.EventKindProjectAliasRemove):
		return db.ProjectProjectAliasRemoveEvent(event)
	case string(types.EventKindProjectDelete):
		return db.ProjectProjectDeleteEvent(event)
	case string(types.EventKindTaskCreated):
		return db.ProjectTaskCreatedEvent(event)
	case string(types.EventKindTaskNumberSet):
		return db.ProjectTaskNumberSetEvent(event)
	case string(types.EventKindTaskRelocate):
		return db.ProjectTaskRelocateEvent(event)
	case string(types.EventKindTaskTitleSet):
		return db.ProjectTaskTitleSetEvent(event)
	case string(types.EventKindTaskDelete):
		return db.ProjectTaskDeleteEvent(event)
	}
	return nil
}
