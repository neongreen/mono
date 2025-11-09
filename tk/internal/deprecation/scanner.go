package deprecation

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/neongreen/mono/tk/internal/types"
)

// ScanEvent scans an event for deprecated field usage and records it in the tracker
func ScanEvent(event types.Event, tracker *Tracker) error {
	// Get deprecated fields for this event kind
	deprecatedFields := GetDeprecatedFieldsForEventKind(event.Kind)
	if len(deprecatedFields) == 0 {
		return nil // No deprecated fields to track for this event kind
	}

	// Unmarshal payload into the appropriate type
	payload, err := unmarshalPayloadByKind(event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload for %s: %w", event.Kind, err)
	}

	// Check each deprecated field
	for _, fieldSpec := range deprecatedFields {
		if isFieldNonEmpty(payload, fieldSpec.FieldName) {
			tracker.RecordUsage(fieldSpec.FieldPath(), event.Kind)
		}
	}

	return nil
}

// unmarshalPayloadByKind unmarshals event payload into the correct type based on event kind
func unmarshalPayloadByKind(event types.Event) (interface{}, error) {
	switch event.Kind {
	case string(types.EventKindProjectAliasAdd):
		var payload types.ProjectAliasAddPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return nil, err
		}
		return payload, nil

	case string(types.EventKindProjectAliasRemove):
		var payload types.ProjectAliasRemovePayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return nil, err
		}
		return payload, nil

	default:
		// Event kind has no deprecated fields to track
		return nil, nil
	}
}

// isFieldNonEmpty checks if a field in a struct is non-empty using reflection
func isFieldNonEmpty(payload interface{}, fieldName string) bool {
	if payload == nil {
		return false
	}

	v := reflect.ValueOf(payload)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return false
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return false
	}

	return isNonEmpty(field)
}

// isNonEmpty checks if a reflect.Value is non-empty
func isNonEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Bool:
		return v.Bool()
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Ptr, reflect.Interface:
		return !v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() > 0
	default:
		// For unknown types, consider non-empty if not zero value
		return !v.IsZero()
	}
}

// ScanAllEvents scans all events in a database for deprecated field usage
func ScanAllEvents(events []types.Event) (*Tracker, error) {
	tracker := NewTracker()

	for _, event := range events {
		if err := ScanEvent(event, tracker); err != nil {
			// Log error but continue scanning
			// Some events might fail to unmarshal if schema changed
			continue
		}
	}

	return tracker, nil
}
