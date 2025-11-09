package deprecation

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanEvent_DetectsAliasFields(t *testing.T) {
	tracker := NewTracker()

	// Create an alias event with non-empty deprecated fields
	payload := types.ProjectAliasAddPayload{
		ProjectUID: "prj_123",
		Alias:      "test-alias", // deprecated:v5 - should be tracked
		Node:       "node123",    // deprecated:v5 - should be tracked
		AddedBy:    "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	event := types.Event{
		ID:        "evt_001",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindProjectAliasAdd),
		Payload:   payloadJSON,
	}

	err = ScanEvent(event, tracker)
	require.NoError(t, err)

	// Should have tracked both deprecated fields
	stats := tracker.GetStats()
	assert.Len(t, stats, 2, "should track 2 deprecated fields")

	// Check Alias field was tracked
	aliasStat, found := tracker.GetStat("ProjectAliasAddPayload.Alias")
	assert.True(t, found, "should track Alias field")
	assert.Equal(t, int64(1), aliasStat.Count)
	assert.Equal(t, string(types.EventKindProjectAliasAdd), aliasStat.EventKind)

	// Check Node field was tracked
	nodeStat, found := tracker.GetStat("ProjectAliasAddPayload.Node")
	assert.True(t, found, "should track Node field")
	assert.Equal(t, int64(1), nodeStat.Count)
}

func TestScanEvent_IgnoresNonDeprecatedEvents(t *testing.T) {
	tracker := NewTracker()

	// Create a project.created event (no deprecated fields)
	payload := types.ProjectCreatedPayload{
		ProjectUID:  "prj_123",
		Type:        "local",
		Name:        "test",
		Description: "test project",
		CreatedBy:   "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	event := types.Event{
		ID:        "evt_001",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}

	err = ScanEvent(event, tracker)
	require.NoError(t, err)

	// Should not track anything
	stats := tracker.GetStats()
	assert.Len(t, stats, 0, "should not track non-deprecated events")
}

func TestScanAllEvents_AggregatesCounts(t *testing.T) {
	// Create multiple alias events
	events := []types.Event{}

	for i := 0; i < 5; i++ {
		payload := types.ProjectAliasAddPayload{
			ProjectUID: "prj_123",
			Alias:      "test",
			Node:       "node",
			AddedBy:    "tester",
		}
		payloadJSON, _ := json.Marshal(payload)

		event := types.Event{
			ID:        string(types.NewEventID()),
			TS:        int64(i),
			CreatedAt: time.Now(),
			Actor:     "tester",
			Role:      "human",
			Kind:      string(types.EventKindProjectAliasAdd),
			Payload:   payloadJSON,
		}
		events = append(events, event)
	}

	tracker, err := ScanAllEvents(events)
	require.NoError(t, err)

	// Should aggregate counts
	aliasStat, found := tracker.GetStat("ProjectAliasAddPayload.Alias")
	assert.True(t, found)
	assert.Equal(t, int64(5), aliasStat.Count, "should count 5 alias usages")
}

func TestGetDeprecatedFieldsForEventKind(t *testing.T) {
	fields := GetDeprecatedFieldsForEventKind(string(types.EventKindProjectAliasAdd))

	assert.Len(t, fields, 2, "should have 2 deprecated fields for project.alias.add")

	// Check fields are Alias and Node
	fieldNames := make(map[string]bool)
	for _, f := range fields {
		fieldNames[f.FieldName] = true
	}

	assert.True(t, fieldNames["Alias"], "should include Alias field")
	assert.True(t, fieldNames["Node"], "should include Node field")
}
