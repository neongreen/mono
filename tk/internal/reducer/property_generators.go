package reducer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
	"pgregory.net/rapid"
)

// Event generators for property-based testing
//
// These generators create valid random events for testing CRDT properties.

// GenEventID generates a random event ID
func GenEventID() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		// Generate a simple event ID with random number
		num := rapid.IntRange(1, 999999).Draw(t, "event_num")
		node := rapid.StringN(8, 8, -1).Draw(t, "node")
		return fmt.Sprintf("ev-%d-%s", num, node)
	})
}

// GenTaskUID generates a random task UID
func GenTaskUID() *rapid.Generator[string] {
	return rapid.StringMatching(`task_[0-9A-HJKMNP-TV-Z]{26}`).
		Filter(func(s string) bool { return len(s) == 31 })
}

// GenProjectUID generates a random project UID
func GenProjectUID() *rapid.Generator[string] {
	return rapid.StringMatching(`proj_[0-9A-HJKMNP-TV-Z]{26}`).
		Filter(func(s string) bool { return len(s) == 31 })
}

// GenNodeID generates a random node ID
func GenNodeID() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-z0-9]{8}`)
}

// GenActor generates a random actor name
func GenActor() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"alice", "bob", "charlie", "dave"})
}

// GenRole generates a random role
func GenRole() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"human", "agent", "bot"})
}

// GenTaskCreatedEvent generates a random task.created event
func GenTaskCreatedEvent(ts int64) *rapid.Generator[types.Event] {
	return rapid.Custom(func(t *rapid.T) types.Event {
		taskUID := GenTaskUID().Draw(t, "task_uid")
		projectUID := GenProjectUID().Draw(t, "project_uid")
		nodeID := GenNodeID().Draw(t, "node_id")
		actor := GenActor().Draw(t, "actor")

		payload := types.TaskCreatedPayload{
			TaskUID:        taskUID,
			ProjectUID:     projectUID,
			ProposedNumber: rapid.Int64Range(1, 1000).Draw(t, "number"),
			CreatedNode:    nodeID,
			Title:          rapid.StringN(5, 50, -1).Draw(t, "title"),
			CreatedBy:      actor,
		}

		payloadJSON, _ := json.Marshal(payload)

		return types.Event{
			ID:        GenEventID().Draw(t, "event_id"),
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      GenRole().Draw(t, "role"),
			Kind:      string(types.EventKindTaskCreated),
			Payload:   payloadJSON,
		}
	})
}

// GenTaskStatusSetEvent generates a random task.status.set event
func GenTaskStatusSetEvent(taskUID string, ts int64) *rapid.Generator[types.Event] {
	return rapid.Custom(func(t *rapid.T) types.Event {
		actor := GenActor().Draw(t, "actor")
		role := GenRole().Draw(t, "role")

		payload := types.TaskStatusSetPayload{
			TaskUUID: taskUID,
			Axis:     "generic",
			State:    rapid.SampledFrom([]string{"todo", "wip", "done", "blocked"}).Draw(t, "state"),
			Role:     role,
		}

		payloadJSON, _ := json.Marshal(payload)

		return types.Event{
			ID:        GenEventID().Draw(t, "event_id"),
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      role,
			Kind:      "task.status.set",
			Payload:   payloadJSON,
		}
	})
}

// GenRelationAddEvent generates a random relation.add event
func GenRelationAddEvent(srcUID, dstUID string, ts int64) *rapid.Generator[types.Event] {
	return rapid.Custom(func(t *rapid.T) types.Event {
		actor := GenActor().Draw(t, "actor")

		payload := types.RelationAddPayload{
			Src:  srcUID,
			Type: rapid.SampledFrom([]string{"blocks", "subtask", "related"}).Draw(t, "relation_type"),
			Dst:  dstUID,
			Note: "",
		}

		payloadJSON, _ := json.Marshal(payload)

		return types.Event{
			ID:        GenEventID().Draw(t, "event_id"),
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      GenRole().Draw(t, "role"),
			Kind:      "relation.add",
			Payload:   payloadJSON,
		}
	})
}

// GenEventSequence generates a valid sequence of events with consistent Lamport timestamps
func GenEventSequence(minEvents, maxEvents int) *rapid.Generator[[]types.Event] {
	return rapid.Custom(func(t *rapid.T) []types.Event {
		numEvents := rapid.IntRange(minEvents, maxEvents).Draw(t, "num_events")
		events := make([]types.Event, 0, numEvents)

		// Create a few tasks first
		numTasks := rapid.IntRange(1, 5).Draw(t, "num_tasks")
		taskUIDs := make([]string, numTasks)

		for i := 0; i < numTasks; i++ {
			event := GenTaskCreatedEvent(int64(i)).Draw(t, "task_created")
			events = append(events, event)

			// Extract task UID from payload
			var payload types.TaskCreatedPayload
			json.Unmarshal(event.Payload, &payload)
			taskUIDs[i] = payload.TaskUID
		}

		// Generate remaining events
		ts := int64(numTasks)
		for len(events) < numEvents {
			eventType := rapid.IntRange(0, 2).Draw(t, "event_type")
			taskUID := rapid.SampledFrom(taskUIDs).Draw(t, "target_task")

			var event types.Event
			switch eventType {
			case 0:
				// Status change
				event = GenTaskStatusSetEvent(taskUID, ts).Draw(t, "status_event")
			case 1:
				// Relation
				otherTask := rapid.SampledFrom(taskUIDs).Draw(t, "other_task")
				if otherTask != taskUID {
					event = GenRelationAddEvent(taskUID, otherTask, ts).Draw(t, "relation_event")
				} else {
					continue // Skip self-relations
				}
			default:
				// Create another task
				event = GenTaskCreatedEvent(ts).Draw(t, "new_task")
				var payload types.TaskCreatedPayload
				json.Unmarshal(event.Payload, &payload)
				taskUIDs = append(taskUIDs, payload.TaskUID)
			}

			events = append(events, event)
			ts++
		}

		return events
	})
}
