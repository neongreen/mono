package types

import (
	"encoding/json"
	"testing"
)

func TestQueuePushPayload_JSONMarshaling(t *testing.T) {
	// Create a payload with typed TaskUID
	taskUID := NewTaskUID()
	payload := QueuePushPayload{
		ContainerID: "q-1",
		ItemID:      taskUID,
	}

	// Marshal to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Marshaled JSON: %s", string(data))

	// Unmarshal back
	var unmarshaled QueuePushPayload
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if unmarshaled.ItemID != taskUID {
		t.Errorf("Round-trip failed: got %v, want %v", unmarshaled.ItemID, taskUID)
	}

	t.Log("✓ TaskUID type works with JSON marshaling (no custom marshal needed)")
}
