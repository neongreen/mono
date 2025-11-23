package reducer

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
)

// applyRelationAdd adds a relation edge
func (r *Reducer) applyRelationAdd(e types.Event) error {
	var payload types.RelationAddPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal relation.add payload: %w", err)
	}

	// Extract node from event ID (format: ev-<number>-<node>)
	node := ""
	if len(e.ID) > 0 {
		parts := utils.SplitEventID(e.ID)
		if len(parts) == 3 {
			node = parts[2]
		}
	}

	r.relations.AddRelation(payload.Src, payload.Type, payload.Dst, payload.Note, e.ID, node, e.TS)
	return nil
}

// applyRelationRemove removes a relation edge
func (r *Reducer) applyRelationRemove(e types.Event) error {
	var payload types.RelationRemovePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal relation.remove payload: %w", err)
	}

	// Extract node from event ID
	node := ""
	if len(e.ID) > 0 {
		parts := utils.SplitEventID(e.ID)
		if len(parts) == 3 {
			node = parts[2]
		}
	}

	r.relations.RemoveRelation(payload.Src, payload.Type, payload.Dst, e.ID, node, e.TS)
	return nil
}

// applyRelationNote sets a note on a relation
func (r *Reducer) applyRelationNote(e types.Event) error {
	var payload types.RelationNotePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal relation.note payload: %w", err)
	}

	r.relations.SetRelationNote(payload.Src, payload.Type, payload.Dst, payload.Markdown)
	return nil
}
