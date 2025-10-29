package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check if it's a file or a remote name
		if _, err := os.Stat(pathOrRemote); err == nil {
			// It's a file
			return ingestFile(db, pathOrRemote)
		}

		// Try as a remote name
		config, err := LoadConfig()
		if err != nil {
			return err
		}

		remote, exists := config.Remotes[pathOrRemote]
		if !exists {
			return fmt.Errorf("'%s' is neither a file nor a configured remote", pathOrRemote)
		}

		return ingestRemote(db, pathOrRemote, remote)
	},
}

// ingestFile ingests events from a single segment file
func ingestFile(db *DB, path string) error {
	reader := NewSegmentReader(path)
	events, err := reader.ReadEvents()
	if err != nil {
		return fmt.Errorf("failed to read segment file: %w", err)
	}

	// Check database version to determine if we should project v4 events
	dbVersion, err := db.GetDBVersion()
	if err != nil {
		return fmt.Errorf("failed to get database version: %w", err)
	}

	ingested := 0
	duplicates := 0

	for _, segEvent := range events {
		// Convert segment event to Event
		event, err := segmentEventToEvent(segEvent)
		if err != nil {
			return fmt.Errorf("failed to convert segment event %s: %w", segEvent.ID, err)
		}

		// Try to insert event (will fail if duplicate)
		err = db.InsertEvent(event)
		if err != nil {
			// Check if it's a duplicate error
			if isDuplicateError(err) {
				duplicates++
				continue
			}
			return fmt.Errorf("failed to insert event %s: %w", event.ID, err)
		}

		// Bump lamport if needed
		if err := db.BumpLamport(event.TS); err != nil {
			return fmt.Errorf("failed to bump lamport: %w", err)
		}

		// Project prefix events into prefixes table (v3 compatible)
		if event.Kind == "prefix.created" {
			if err := db.ProjectPrefixCreatedEvent(event); err != nil {
				// Log but don't fail - projection errors are not critical
				fmt.Fprintf(os.Stderr, "Warning: failed to project prefix.created event %s: %v\n", event.ID, err)
			}
		}
		if event.Kind == "prefix.removed" {
			if err := db.ProjectPrefixRemovedEvent(event); err != nil {
				// Log but don't fail - projection errors are not critical
				fmt.Fprintf(os.Stderr, "Warning: failed to project prefix.removed event %s: %v\n", event.ID, err)
			}
		}

		// Only project v4 events if database is v4 or higher
		if dbVersion >= 4 {
			// Project v4 events into their respective tables
			switch event.Kind {
			case string(EventKindProjectCreated):
				if err := db.ProjectProjectCreatedEvent(event); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to project project.created event %s: %v\n", event.ID, err)
				}
			case string(EventKindProjectAliasAdd):
				if err := db.ProjectProjectAliasAddEvent(event); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to project project.alias.add event %s: %v\n", event.ID, err)
				}
			case string(EventKindProjectAliasRemove):
				if err := db.ProjectProjectAliasRemoveEvent(event); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to project project.alias.remove event %s: %v\n", event.ID, err)
				}
			case string(EventKindTaskCreated):
				if err := db.ProjectTaskCreatedV4Event(event); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to project task.created event %s: %v\n", event.ID, err)
				}
			case string(EventKindTaskNumberSet):
				if err := db.ProjectTaskNumberSetEvent(event); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to project task.number.set event %s: %v\n", event.ID, err)
				}
			case string(EventKindTaskRelocate):
				if err := db.ProjectTaskRelocateEvent(event); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to project task.relocate event %s: %v\n", event.ID, err)
				}
			case string(EventKindTaskTitleSet):
				if err := db.ProjectTaskTitleSetEvent(event); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to project task.title.set event %s: %v\n", event.ID, err)
				}
			}
		}

		ingested++
	}

	fmt.Printf("Ingested %d events (%d duplicates skipped)\n", ingested, duplicates)
	return nil
}

// ingestRemote ingests events from all segments in a remote
func ingestRemote(db *DB, remoteName string, remote RemoteConfig) error {
	space := "personal" // TODO: make this configurable

	// Find all segment files
	segmentsDir := filepath.Join(remote.Path, space, "segments")
	if _, err := os.Stat(segmentsDir); os.IsNotExist(err) {
		fmt.Println("No segments directory found")
		return nil
	}

	var segmentFiles []string
	err := filepath.Walk(segmentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".zst" {
			segmentFiles = append(segmentFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan segments directory: %w", err)
	}

	if len(segmentFiles) == 0 {
		fmt.Println("No segment files found")
		return nil
	}

	// Check database version to determine if we should project v4 events
	dbVersion, err := db.GetDBVersion()
	if err != nil {
		return fmt.Errorf("failed to get database version: %w", err)
	}

	totalIngested := 0
	totalDuplicates := 0

	for _, segmentFile := range segmentFiles {
		reader := NewSegmentReader(segmentFile)
		events, err := reader.ReadEvents()
		if err != nil {
			fmt.Printf("Warning: failed to read %s: %v\n", segmentFile, err)
			continue
		}

		for _, segEvent := range events {
			// Convert segment event to Event
			event, err := segmentEventToEvent(segEvent)
			if err != nil {
				return fmt.Errorf("failed to convert segment event %s: %w", segEvent.ID, err)
			}

			// Try to insert event (will fail if duplicate)
			err = db.InsertEvent(event)
			if err != nil {
				// Check if it's a duplicate error
				if isDuplicateError(err) {
					totalDuplicates++
					continue
				}
				return fmt.Errorf("failed to insert event %s: %w", event.ID, err)
			}

			// Bump lamport if needed
			if err := db.BumpLamport(event.TS); err != nil {
				return fmt.Errorf("failed to bump lamport: %w", err)
			}

			// Project prefix events into prefixes table
			if event.Kind == "prefix.created" {
				if err := db.ProjectPrefixCreatedEvent(event); err != nil {
					// Log but don't fail - projection errors are not critical
					fmt.Fprintf(os.Stderr, "Warning: failed to project prefix.created event %s: %v\n", event.ID, err)
				}
			}
			if event.Kind == "prefix.removed" {
				if err := db.ProjectPrefixRemovedEvent(event); err != nil {
					// Log but don't fail - projection errors are not critical
					fmt.Fprintf(os.Stderr, "Warning: failed to project prefix.removed event %s: %v\n", event.ID, err)
				}
			}

			// Only project v4 events if database is v4 or higher
			if dbVersion >= 4 {
				// Project v4 events into their respective tables
				switch event.Kind {
				case string(EventKindProjectCreated):
					if err := db.ProjectProjectCreatedEvent(event); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to project project.created event %s: %v\n", event.ID, err)
					}
				case string(EventKindProjectAliasAdd):
					if err := db.ProjectProjectAliasAddEvent(event); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to project project.alias.add event %s: %v\n", event.ID, err)
					}
				case string(EventKindProjectAliasRemove):
					if err := db.ProjectProjectAliasRemoveEvent(event); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to project project.alias.remove event %s: %v\n", event.ID, err)
					}
				case string(EventKindTaskCreated):
					if err := db.ProjectTaskCreatedV4Event(event); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to project task.created event %s: %v\n", event.ID, err)
					}
				case string(EventKindTaskNumberSet):
					if err := db.ProjectTaskNumberSetEvent(event); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to project task.number.set event %s: %v\n", event.ID, err)
					}
				case string(EventKindTaskRelocate):
					if err := db.ProjectTaskRelocateEvent(event); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to project task.relocate event %s: %v\n", event.ID, err)
					}
				case string(EventKindTaskTitleSet):
					if err := db.ProjectTaskTitleSetEvent(event); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to project task.title.set event %s: %v\n", event.ID, err)
					}
				}
			}

			totalIngested++
		}
	}

	// Save watermark
	stateDir, err := GetStateDir()
	if err != nil {
		return err
	}

	watermarkFile := filepath.Join(stateDir, "ingest_watermarks", remoteName, space+".json")
	watermark := IngestWatermark{
		RemoteName: remoteName,
		Space:      space,
		UpdatedAt:  time.Now(),
	}
	if err := saveIngestWatermark(watermarkFile, &watermark); err != nil {
		return err
	}

	fmt.Printf("Ingested %d events from %d segments (%d duplicates skipped)\n",
		totalIngested, len(segmentFiles), totalDuplicates)
	return nil
}

// segmentEventToEvent converts a SegmentEvent to an Event
func segmentEventToEvent(se SegmentEvent) (Event, error) {
	// Parse timestamp
	createdAt, err := time.Parse(time.RFC3339Nano, se.TS)
	if err != nil {
		// Try RFC3339 without nano
		createdAt, err = time.Parse(time.RFC3339, se.TS)
		if err != nil {
			return Event{}, fmt.Errorf("failed to parse timestamp: %w", err)
		}
	}

	// Marshal payload back to JSON
	payloadJSON, err := json.Marshal(se.Payload)
	if err != nil {
		return Event{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := Event{
		ID:        se.ID,
		TS:        se.Lamport,
		CreatedAt: createdAt,
		Actor:     se.Actor,
		Role:      se.Role,
		Kind:      se.Kind,
		Payload:   payloadJSON,
	}

	if se.Ctx != nil {
		if se.Ctx.RepoUUID != nil {
			event.RepoUUID = *se.Ctx.RepoUUID
		}
		if se.Ctx.Branch != nil {
			event.Branch = *se.Ctx.Branch
		}
		if se.Ctx.Commit != nil {
			event.Commit = *se.Ctx.Commit
		}
		if se.Ctx.JJOpID != nil {
			event.JJOpID = *se.Ctx.JJOpID
		}
	}

	return event, nil
}

// isDuplicateError checks if an error is a duplicate key error
func isDuplicateError(err error) bool {
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

// saveIngestWatermark saves an ingest watermark to a file
func saveIngestWatermark(path string, watermark *IngestWatermark) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create watermark directory: %w", err)
	}

	data, err := json.MarshalIndent(watermark, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal watermark: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write watermark: %w", err)
	}

	return nil
}
