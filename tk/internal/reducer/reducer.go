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
		//lion:events section="task.status.set"
		// Applies a status update to the reducer task view, keeping history consistent with projection tables while ignoring unknown statuses for forward compatibility.
		return r.applyTaskStatusSet(e)
	case "task.note.add":
		//lion:events section="task.note.add"
		// Appends an immutable note entry to the reducer, keeping timestamps from the event payload and leaving existing notes untouched.
		return r.applyTaskNoteAdd(e)
	case "task.delete":
		//lion:events section="task.delete"
		// Marks tasks as deleted in-memory without removing historical data so later events can no-op safely when rebuilding from the log.
		return r.applyTaskDelete(e)
	case "task.meta.set":
		//lion:events section="task.meta.set"
		// Updates reducer metadata claims and preserves competing values so resolution logic stays deterministic with the projection layer.
		return r.applyTaskMetaSet(e)
	case "relation.add":
		//lion:events section="relation.add"
		// Adds a relation edge between tasks and recomputes the relations graph so blockers and dependents stay in sync with later FinalizeRelations calls.
		return r.applyRelationAdd(e)
	case "relation.remove":
		//lion:events section="relation.remove"
		// Removes a relation edge if present; missing edges are ignored to keep replay idempotent when events arrive out of order.
		return r.applyRelationRemove(e)
	case "relation.note":
		//lion:events section="relation.note"
		// Attaches free-form notes to relations without altering the graph structure, allowing multiple notes per relation across the log.
		return r.applyRelationNote(e)
	case "task.attachment.add":
		//lion:events section="task.attachment.add"
		// Tracks attachment metadata on the reducer task so CLI commands can surface linked artifacts without querying projection tables.
		return r.applyTaskAttachmentAdd(e)
	case "task.attachment.remove":
		//lion:events section="task.attachment.remove"
		// Removes attachment references if they exist; missing attachments are ignored to keep log replay idempotent.
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
