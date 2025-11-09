package deprecation

import (
	"github.com/neongreen/mono/tk/internal/types"
)

// FieldSpec defines a deprecated field to track
type FieldSpec struct {
	EventKind string // e.g., "project.alias.add"
	TypeName  string // e.g., "ProjectAliasAddPayload"
	FieldName string // e.g., "Alias"
	Version   string // e.g., "v5"
	Reason    string // why deprecated
}

// FieldPath returns the full field path for display
func (f FieldSpec) FieldPath() string {
	return f.TypeName + "." + f.FieldName
}

// v5DeprecatedFields lists all fields deprecated in v5
var v5DeprecatedFields = []FieldSpec{
	{
		EventKind: string(types.EventKindProjectAliasAdd),
		TypeName:  "ProjectAliasAddPayload",
		FieldName: "Alias",
		Version:   "v5",
		Reason:    "Aliases removed in favor of project names",
	},
	{
		EventKind: string(types.EventKindProjectAliasAdd),
		TypeName:  "ProjectAliasAddPayload",
		FieldName: "Node",
		Version:   "v5",
		Reason:    "Aliases removed in favor of project names",
	},
	{
		EventKind: string(types.EventKindProjectAliasRemove),
		TypeName:  "ProjectAliasRemovePayload",
		FieldName: "Alias",
		Version:   "v5",
		Reason:    "Aliases removed in favor of project names",
	},
	{
		EventKind: string(types.EventKindProjectAliasRemove),
		TypeName:  "ProjectAliasRemovePayload",
		FieldName: "Node",
		Version:   "v5",
		Reason:    "Aliases removed in favor of project names",
	},
}

// GetDeprecatedFieldsForVersion returns all fields deprecated in a specific version
func GetDeprecatedFieldsForVersion(version string) []FieldSpec {
	switch version {
	case "v5":
		return v5DeprecatedFields
	default:
		return nil
	}
}

// GetAllDeprecatedFields returns all registered deprecated fields
func GetAllDeprecatedFields() []FieldSpec {
	return v5DeprecatedFields
}

// GetDeprecatedFieldsForEventKind returns deprecated fields for a specific event kind
func GetDeprecatedFieldsForEventKind(eventKind string) []FieldSpec {
	var fields []FieldSpec
	for _, field := range v5DeprecatedFields {
		if field.EventKind == eventKind {
			fields = append(fields, field)
		}
	}
	return fields
}
