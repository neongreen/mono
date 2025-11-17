package transformers

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestV4ToV5Transformer(t *testing.T) {
	transformer := &V4ToV5Transformer{}

	tests := []struct {
		name     string
		event    types.Event
		filtered bool // true if event should be filtered out
	}{
		{
			name: "project.alias.add is filtered",
			event: types.Event{
				Kind: string(types.EventKindProjectAliasAdd),
			},
			filtered: true,
		},
		{
			name: "project.alias.remove is filtered",
			event: types.Event{
				Kind: string(types.EventKindProjectAliasRemove),
			},
			filtered: true,
		},
		{
			name: "project.created passes through",
			event: types.Event{
				Kind: string(types.EventKindProjectCreated),
			},
			filtered: false,
		},
		{
			name: "task.created passes through",
			event: types.Event{
				Kind: string(types.EventKindTaskCreated),
			},
			filtered: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformer.Transform(tt.event)
			if tt.filtered && result != nil {
				t.Errorf("Expected event to be filtered out but got %v", result)
			}
			if !tt.filtered && result == nil {
				t.Errorf("Expected event to pass through but got nil")
			}
			if !tt.filtered && result != nil && result.Kind != tt.event.Kind {
				t.Errorf("Expected kind %s but got %s", tt.event.Kind, result.Kind)
			}
		})
	}
}

func TestTransformEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    types.Event
		filtered bool
	}{
		{
			name: "alias events filtered by v4-to-v5 transformer",
			event: types.Event{
				ID:   "ev-1",
				Kind: string(types.EventKindProjectAliasAdd),
			},
			filtered: true,
		},
		{
			name: "normal events pass through all transformers",
			event: types.Event{
				ID:   "ev-2",
				Kind: string(types.EventKindTaskCreated),
			},
			filtered: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformEvent(tt.event)
			if tt.filtered && result != nil {
				t.Errorf("Expected event to be filtered but got %v", result)
			}
			if !tt.filtered && result == nil {
				t.Errorf("Expected event to pass through but got nil")
			}
		})
	}
}
