package beads

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// ImportBeadsRelationships imports relationships for a beads issue
func ImportBeadsRelationships(db *database.DB, issue BeadsIssue, taskUID string, issueMap map[string]string) (int, error) {
	count := 0

	if issue.Dependencies == nil {
		return 0, nil
	}

	// Process dependency array
	for _, dep := range issue.Dependencies {
		// Only process dependencies where this issue is the source
		if dep.IssueID != issue.ID {
			continue
		}

		// Look up the target task UID
		targetUID, ok := issueMap[dep.DependsOnID]
		if !ok {
			// Target issue not imported, skip
			continue
		}

		// Map beads relationship type to tk type
		tkRelType, err := mapRelationType(dep.Type)
		if err != nil {
			// Unknown type, skip
			continue
		}

		// Create the relationship
		if err := CreateRelation(db, taskUID, targetUID, tkRelType); err != nil {
			return count, fmt.Errorf("failed to create relation: %w", err)
		}
		count++
	}

	return count, nil
}

// mapRelationType maps beads relationship type to tk type
func mapRelationType(beadsType string) (string, error) {
	switch beadsType {
	case "parent-child":
		return "parent", nil
	case "blocks":
		return "blocks", nil
	case "related":
		return "related", nil
	case "discovered-from":
		return "related", nil // Map discovered-from to related
	default:
		return "", fmt.Errorf("unknown relationship type: %s", beadsType)
	}
}

// CreateRelation creates a relationship between two tasks
func CreateRelation(db *database.DB, fromUID, toUID, relType string) error {
	payload := types.RelationAddPayload{
		Src:  fromUID,
		Type: relType,
		Dst:  toUID,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal relation payload: %w", err)
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "importer",
		Role:      "human",
		Kind:      string(types.EventKindRelationAdd),
		Payload:   payloadJSON,
	}

	return db.InsertEvent(event)
}
