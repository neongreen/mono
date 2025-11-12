package transformers

import "github.com/neongreen/mono/tk/internal/types"

// V4ToV5Transformer filters out deprecated v4 events
// This transformer handles the v4→v5 migration where project aliases were removed
type V4ToV5Transformer struct{}

// Transform filters out project.alias.add and project.alias.remove events
// These events are deprecated in v5 and should be ignored
func (t *V4ToV5Transformer) Transform(event types.Event) *types.Event {
	// Filter out deprecated alias events
	switch event.Kind {
	case string(types.EventKindProjectAliasAdd):
		return nil // Filter out
	case string(types.EventKindProjectAliasRemove):
		return nil // Filter out
	default:
		return &event // Pass through unchanged
	}
}
