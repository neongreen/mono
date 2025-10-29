package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

		// Open DB
		db, err := openExistingDB()
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
			config, err := LoadConfig()
			if err != nil {
				return err
			}

			remote, exists := config.Remotes[pathOrRemote]
			if !exists {
				return fmt.Errorf("'%s' is neither a file nor a configured remote", pathOrRemote)
			}

			ingestErr = ingestRemote(db, pathOrRemote, remote)
		}

		if ingestErr != nil {
			return ingestErr
		}

		return nil
	},
}

// ingestFile ingests events from a single segment file
func ingestFile(db *DB, path string) error {
	reader := NewSegmentReader(path)
	events, err := reader.ReadEvents()
	if err != nil {
		return fmt.Errorf("failed to read segment file: %w", err)
	}

	ingested := 0
	duplicates := 0

	for _, segEvent := range events {
		// Convert segment event to Event
		event, err := segmentEventToEvent(segEvent)
		if err != nil {
			return fmt.Errorf("failed to convert segment event %s: %w", segEvent.ID, err)
		}

		// Transform legacy events to v4 events BEFORE inserting
		transformedEvents, err := TransformLegacyEvent(event, func(prefix, description, createdBy string, createdAt time.Time, nodeID string) (string, error) {
			return getOrCreateProjectForPrefix(db, prefix, description, createdBy, createdAt, nodeID)
		})
		if err != nil {
			return fmt.Errorf("failed to transform legacy event %s: %w", event.ID, err)
		}

		// If event was transformed, insert transformed events instead of original
		if len(transformedEvents) > 0 {
			for _, transformedEvent := range transformedEvents {
				// Check for duplicates on transformed events
				err = db.InsertEvent(transformedEvent)
				if err != nil {
					if isDuplicateError(err) {
						duplicates++
						continue
					}
					return fmt.Errorf("failed to insert transformed event %s: %w", transformedEvent.ID, err)
				}
				// Bump lamport if needed
				if err := db.BumpLamport(transformedEvent.TS); err != nil {
					return fmt.Errorf("failed to bump lamport: %w", err)
				}
				// Project v4 events
				if err := projectV4Event(db, transformedEvent); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to project v4 event %s: %v\n", transformedEvent.ID, err)
				}
				ingested++
			}
			// Skip the original legacy event - we've inserted transformed events instead
			continue
		}

		// Event was not transformed, insert original event
		err = db.InsertEvent(event)
		if err != nil {
			// Check if it's a duplicate error
			if isDuplicateError(err) {
				duplicates++
				continue
			}
			return fmt.Errorf("failed to insert event %s: %w", event.ID, err)
		}

		// Project v4 events (always v4 now)
		if err := projectV4Event(db, event); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to project v4 event %s: %v\n", event.ID, err)
		}

		ingested++
	}

	fmt.Printf("Ingested %d events (%d duplicates skipped)\n", ingested, duplicates)
	return nil
}

// ingestRemote ingests events from all segments in a remote
func ingestRemote(db *DB, remoteName string, remote RemoteConfig) error {
	// Use configured spaces, or default to "personal"
	spaces := remote.Spaces
	if len(spaces) == 0 {
		spaces = []string{"personal"}
	}

	for _, space := range spaces {
		if err := ingestRemoteSpace(db, remoteName, remote, space); err != nil {
			return err
		}
	}

	return nil
}

// ingestRemoteSpace ingests events from a specific space in a remote
func ingestRemoteSpace(db *DB, remoteName string, remote RemoteConfig, space string) error {
	// Find all segment files
	segmentsDir := filepath.Join(remote.Path, space, "segments")
	if _, err := os.Stat(segmentsDir); os.IsNotExist(err) {
		fmt.Printf("No segments directory found for space '%s'\n", space)
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

			// Transform legacy events to v4 events BEFORE inserting
			transformedEvents, err := TransformLegacyEvent(event, func(prefix, description, createdBy string, createdAt time.Time, nodeID string) (string, error) {
				return getOrCreateProjectForPrefix(db, prefix, description, createdBy, createdAt, nodeID)
			})
			if err != nil {
				return fmt.Errorf("failed to transform legacy event %s: %w", event.ID, err)
			}

			// If event was transformed, insert transformed events instead of original
			if len(transformedEvents) > 0 {
				for _, transformedEvent := range transformedEvents {
					// Check for duplicates on transformed events
					err = db.InsertEvent(transformedEvent)
					if err != nil {
						if isDuplicateError(err) {
							totalDuplicates++
							continue
						}
						return fmt.Errorf("failed to insert transformed event %s: %w", transformedEvent.ID, err)
					}
					// Bump lamport if needed
					if err := db.BumpLamport(transformedEvent.TS); err != nil {
						return fmt.Errorf("failed to bump lamport: %w", err)
					}
					// Project v4 events
					if err := projectV4Event(db, transformedEvent); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to project v4 event %s: %v\n", transformedEvent.ID, err)
					}
					totalIngested++
				}
				// Skip the original legacy event - we've inserted transformed events instead
				continue
			}

			// Event was not transformed, insert original event
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

			// Project v4 events (always v4 now)
			if err := projectV4Event(db, event); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to project v4 event %s: %v\n", event.ID, err)
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

// transformLegacyEvent transforms legacy v1/v2/v3 events into v4 events
// Returns empty slice if event is not a legacy event that needs transformation
func transformLegacyEvent(db *DB, event Event) ([]Event, error) {
	switch event.Kind {
	case "prefix.created":
		return transformPrefixCreated(db, event)
	case "task.created":
		// Check if it's legacy format (has task_id field, not task_uid)
		var payload TaskCreatedPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return nil, nil // Not a legacy task.created, return empty
		}
		// Legacy format has task_id field with format prefix-number-node
		if payload.TaskID != "" && payload.TaskUUID == "" {
			return transformLegacyTaskCreated(db, event)
		}
		return nil, nil // Already v4 format
	case "task.reprefix":
		return transformTaskReprefix(db, event)
	default:
		return nil, nil // Not a legacy event that needs transformation
	}
}

// transformPrefixCreated transforms prefix.created → project.created + project.alias.add
func transformPrefixCreated(db *DB, event Event) ([]Event, error) {
	var payload PrefixCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse prefix.created payload: %w", err)
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return nil, fmt.Errorf("failed to get node ID: %w", err)
	}

	// Extract node from event ID if possible (format: ev-<number>-<node>)
	parts := strings.Split(event.ID, "-")
	if len(parts) >= 3 {
		nodeID = parts[2]
	}

	// Get or create project UID for this prefix
	projectUID, err := getOrCreateProjectForPrefix(db, payload.Prefix, payload.Description, payload.CreatedBy, event.CreatedAt, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create project for prefix %s: %w", payload.Prefix, err)
	}

	// Create project.created event
	projectCreatedPayload := ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        payload.Prefix,
		Description: payload.Description,
		CreatedBy:   payload.CreatedBy,
	}
	projectCreatedJSON, err := json.Marshal(projectCreatedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project.created payload: %w", err)
	}

	projectCreatedEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindProjectCreated),
		Payload:   projectCreatedJSON,
	}

	// Create project.alias.add event
	aliasPayload := ProjectAliasAddPayload{
		ProjectUID: projectUID,
		Alias:      payload.Prefix,
		Node:       nodeID,
		AddedBy:    payload.CreatedBy,
	}
	aliasJSON, err := json.Marshal(aliasPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project.alias.add payload: %w", err)
	}

	aliasEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS + 1, // Slightly after project.created
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindProjectAliasAdd),
		Payload:   aliasJSON,
	}

	return []Event{projectCreatedEvent, aliasEvent}, nil
}

// transformLegacyTaskCreated transforms legacy task.created → v4 task.created + task.number.set
func transformLegacyTaskCreated(db *DB, event Event) ([]Event, error) {
	var payload TaskCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse legacy task.created payload: %w", err)
	}

	// Parse task_id to extract prefix, number, node
	prefix, number, node, err := ParseTaskIDLegacy(payload.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task_id %s: %w", payload.TaskID, err)
	}

	// Get or create project UID for this prefix
	projectUID, err := getOrCreateProjectForPrefix(db, prefix, "", payload.CreatedBy, event.CreatedAt, node)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create project for prefix %s: %w", prefix, err)
	}

	// Generate new task UID
	taskUID := string(NewTaskUID())

	// Create v4 task.created event
	taskCreatedPayload := TaskCreatedV4Payload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: number,
		CreatedNode:    node,
		Title:          payload.Title,
		CreatedBy:      payload.CreatedBy,
	}
	taskCreatedJSON, err := json.Marshal(taskCreatedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task.created payload: %w", err)
	}

	taskCreatedEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskCreated),
		Payload:   taskCreatedJSON,
	}

	// Create task.number.set event
	numberPayload := TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     number,
		Reason:     "migrated from legacy",
	}
	numberJSON, err := json.Marshal(numberPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task.number.set payload: %w", err)
	}

	numberEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS + 1, // Slightly after task.created
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskNumberSet),
		Payload:   numberJSON,
	}

	return []Event{taskCreatedEvent, numberEvent}, nil
}

// transformTaskReprefix transforms task.reprefix → task.relocate
func transformTaskReprefix(db *DB, event Event) ([]Event, error) {
	var payload TaskReprefixPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse task.reprefix payload: %w", err)
	}

	// Get project UIDs for old and new prefixes
	fromProjectUID, err := getOrCreateProjectForPrefix(db, payload.OldPrefix, "", event.Actor, event.CreatedAt, payload.OldNode)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create project for prefix %s: %w", payload.OldPrefix, err)
	}

	toProjectUID, err := getOrCreateProjectForPrefix(db, payload.NewPrefix, "", event.Actor, event.CreatedAt, payload.OldNode)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create project for prefix %s: %w", payload.NewPrefix, err)
	}

	// Get task UID - need to resolve from legacy task
	// For now, we'll need to resolve it from the task_id or task_uuid
	// This is a simplified version - full implementation would need to track UUID mappings
	taskUID := string(NewTaskUID())

	// Create task.relocate event
	relocatePayload := TaskRelocatePayload{
		TaskUID:        taskUID,
		FromProjectUID: fromProjectUID,
		ToProjectUID:   toProjectUID,
		NumberPolicy: NumberPolicyPayload{
			Mode:   "force",
			Number: payload.NewNumber,
		},
	}
	relocateJSON, err := json.Marshal(relocatePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task.relocate payload: %w", err)
	}

	relocateEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskRelocate),
		Payload:   relocateJSON,
	}

	return []Event{relocateEvent}, nil
}

// getOrCreateProjectForPrefix gets or creates a project UID for a given prefix
func getOrCreateProjectForPrefix(db *DB, prefix, description, createdBy string, createdAt time.Time, nodeID string) (string, error) {
	// Check if project already exists for this prefix alias
	var projectUID string
	err := db.db.QueryRow(`
		SELECT project_uid FROM project_aliases 
		WHERE alias = ? AND node = ?
		LIMIT 1
	`, prefix, nodeID).Scan(&projectUID)

	if err == nil {
		// Project exists, return it
		return projectUID, nil
	}

	// Project doesn't exist, create it
	projectUID = string(NewProjectUID())

	// Create project.created event
	projectPayload := ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        prefix,
		Description: description,
		CreatedBy:   createdBy,
	}
	projectJSON, err := json.Marshal(projectPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal project.created payload: %w", err)
	}

	projectEvent := Event{
		ID:        string(NewEventID()),
		TS:        0, // Will be set during insert
		CreatedAt: createdAt,
		Actor:     createdBy,
		Role:      "human",
		Kind:      string(EventKindProjectCreated),
		Payload:   projectJSON,
	}

	// Insert and project the event
	if err := db.InsertEvent(projectEvent); err != nil {
		if !isDuplicateError(err) {
			return "", fmt.Errorf("failed to insert project.created event: %w", err)
		}
		// If duplicate, try to read the existing project
		err = db.db.QueryRow(`
			SELECT project_uid FROM project_aliases 
			WHERE alias = ? AND node = ?
			LIMIT 1
		`, prefix, nodeID).Scan(&projectUID)
		if err != nil {
			return "", fmt.Errorf("failed to read existing project: %w", err)
		}
		return projectUID, nil
	}

	if err := db.ProjectProjectCreatedEvent(projectEvent); err != nil {
		return "", fmt.Errorf("failed to project project.created event: %w", err)
	}

	// Create project.alias.add event
	aliasPayload := ProjectAliasAddPayload{
		ProjectUID: projectUID,
		Alias:      prefix,
		Node:       nodeID,
		AddedBy:    createdBy,
	}
	aliasJSON, err := json.Marshal(aliasPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal project.alias.add payload: %w", err)
	}

	aliasEvent := Event{
		ID:        string(NewEventID()),
		TS:        0, // Will be set during insert
		CreatedAt: createdAt,
		Actor:     createdBy,
		Role:      "human",
		Kind:      string(EventKindProjectAliasAdd),
		Payload:   aliasJSON,
	}

	if err := db.InsertEvent(aliasEvent); err != nil {
		if !isDuplicateError(err) {
			return "", fmt.Errorf("failed to insert project.alias.add event: %w", err)
		}
	}

	if err := db.ProjectProjectAliasAddEvent(aliasEvent); err != nil {
		return "", fmt.Errorf("failed to project project.alias.add event: %w", err)
	}

	return projectUID, nil
}

// projectV4Event projects a v4 event into its respective table
func projectV4Event(db *DB, event Event) error {
	switch event.Kind {
	case string(EventKindProjectCreated):
		return db.ProjectProjectCreatedEvent(event)
	case string(EventKindProjectAliasAdd):
		return db.ProjectProjectAliasAddEvent(event)
	case string(EventKindProjectAliasRemove):
		return db.ProjectProjectAliasRemoveEvent(event)
	case string(EventKindTaskCreated):
		return db.ProjectTaskCreatedV4Event(event)
	case string(EventKindTaskNumberSet):
		return db.ProjectTaskNumberSetEvent(event)
	case string(EventKindTaskRelocate):
		return db.ProjectTaskRelocateEvent(event)
	case string(EventKindTaskTitleSet):
		return db.ProjectTaskTitleSetEvent(event)
	default:
		return nil // Not a v4 event that needs projection
	}
}
