package reducer

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/types"
)

// Apply applies an event to update the task state
func (r *Reducer) Apply(e types.Event) error {
	// Try project events first
	handled, err := r.ApplyProjectEvent(e)
	if err != nil {
		return err
	}
	// If project handler processed this event, don't run legacy handlers
	if handled {
		return nil
	}

	// Handle shared events (status, notes, relations, delete, metadata)
	switch e.Kind {
	case "task.status.set":
		return r.applyTaskStatusSet(e)
	case "task.note.add":
		return r.applyTaskNoteAdd(e)
	case "task.delete":
		return r.applyTaskDelete(e)
	case "task.meta.set":
		return r.applyTaskMetaSet(e)
	case "project.delete":
		return r.applyProjectDelete(e)
	case "relation.add":
		return r.applyRelationAdd(e)
	case "relation.remove":
		return r.applyRelationRemove(e)
	case "relation.note":
		return r.applyRelationNote(e)
	case "task.attachment.add":
		return r.applyTaskAttachmentAdd(e)
	case "task.attachment.remove":
		return r.applyTaskAttachmentRemove(e)
	default:
		// Unknown events are ignored for forward compatibility
		return nil
	}
}

// BuildFromEvents builds the current state from a list of events
func BuildFromEvents(events []types.Event) (*Reducer, error) {
	reducer := NewReducer()
	for _, e := range events {
		if err := reducer.Apply(e); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.ID, err)
		}
	}
	return reducer, nil
}

// BuildFromEventsWithConfig builds the current state from events and finalizes relations
func BuildFromEventsWithConfig(events []types.Event, config *config.Config) (*Reducer, error) {
	reducer := NewReducer()
	for _, e := range events {
		if err := reducer.Apply(e); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.ID, err)
		}
	}
	reducer.FinalizeRelations(config)
	return reducer, nil
}
