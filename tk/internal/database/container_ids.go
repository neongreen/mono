package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
)

// GenerateContainerID generates the next container ID for the given primitive type
// Queue: q-1, q-2, q-3, ...
// Stack: s-1, s-2, s-3, ...
// Group: g-1, g-2, g-3, ...
func (d *DB) GenerateContainerID(primitive types.ContainerPrimitive) (string, error) {
	var prefix string
	switch primitive {
	case types.PrimitiveQueue:
		prefix = "q"
	case types.PrimitiveStack:
		prefix = "s"
	case types.PrimitiveGroup:
		prefix = "g"
	default:
		return "", fmt.Errorf("unknown primitive type: %s", primitive)
	}

	// Find the highest number currently used for this primitive
	var maxNum int64
	query := fmt.Sprintf(`
		SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0)
		FROM containers
		WHERE id LIKE ?
	`, len(prefix)+2) // prefix + "-" = 2 chars

	err := d.Db.QueryRow(query, prefix+"-%").Scan(&maxNum)
	if err != nil {
		return "", fmt.Errorf("failed to get max container number: %w", err)
	}

	nextNum := maxNum + 1
	return fmt.Sprintf("%s-%d", prefix, nextNum), nil
}

// CreateDefaultKind creates a default container kind with the given name and description.
// This is used to auto-create the "general" kind when users don't specify a kind.
func (d *DB) CreateDefaultKind(primitive types.ContainerPrimitive, kindName, description string) error {
	// Get current user
	actor, err := utils.GetCurrentUser()
	if err != nil {
		return err
	}

	// Create container.kind.define event
	payload := types.DefineContainerKindPayload{
		Name:        kindName,
		Primitive:   primitive,
		Description: description,
		CreatedBy:   actor,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	eventID, err := GenerateEventID(d)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := d.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get next lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   payloadJSON,
	}

	if err := d.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	// Project the event into container_kinds table
	if err := d.ProjectContainerKindDefineEvent(event); err != nil {
		return fmt.Errorf("failed to project event: %w", err)
	}

	return nil
}
