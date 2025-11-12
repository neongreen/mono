package transformers

import "github.com/neongreen/mono/tk/internal/types"

// EventTransformer transforms or filters events during ingestion
type EventTransformer interface {
	// Transform takes an event and returns a transformed version or nil to filter it out
	Transform(event types.Event) *types.Event
}

// TransformEvent applies all registered transformers to an event
// Returns nil if the event should be filtered out
func TransformEvent(event types.Event) *types.Event {
	result := &event
	for _, transformer := range transformers {
		result = transformer.Transform(*result)
		if result == nil {
			return nil // Event filtered out
		}
	}
	return result
}

// transformers is the registry of all transformers, applied in order
var transformers = []EventTransformer{
	&V4ToV5Transformer{},
}
