package remote

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// EventToSegmentEvent converts a types.Event to a SegmentEvent for serialization
func EventToSegmentEvent(e types.Event, space, nodeID string) (SegmentEvent, error) {
	// Parse payload to the right type
	var payload any
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return SegmentEvent{}, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	ctx := &SegmentContext{}
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

	return SegmentEvent{
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

// SegmentEventToEvent converts a SegmentEvent to a types.Event for database insertion
func SegmentEventToEvent(se SegmentEvent) (types.Event, error) {
	// Parse timestamp
	createdAt, err := time.Parse(time.RFC3339Nano, se.TS)
	if err != nil {
		// Try RFC3339 without nano
		createdAt, err = time.Parse(time.RFC3339, se.TS)
		if err != nil {
			return types.Event{}, fmt.Errorf("failed to parse timestamp: %w", err)
		}
	}

	// Marshal payload back to JSON
	payloadJSON, err := json.Marshal(se.Payload)
	if err != nil {
		return types.Event{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := types.Event{
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
