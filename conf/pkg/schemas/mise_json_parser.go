package schemas

import (
	"fmt"

	"github.com/neongreen/mono/lib/configschema"
)

// MiseSchemaParser handles parsing of Mise JSON schema for completion data
type MiseSchemaParser struct {
	parser *configschema.JSONSchemaParser
}

// NewMiseSchemaParser creates a new Mise schema parser using the jsonschema library
func NewMiseSchemaParser() (*MiseSchemaParser, error) {
	parser, err := CompileMiseSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to compile Mise schema: %w", err)
	}

	return &MiseSchemaParser{
		parser: parser,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *MiseSchemaParser) GetCompletionOptions(path string) []configschema.CompletionOption {
	return p.parser.GetCompletionOptions(path)
}

// ValidatePath checks if a configuration path exists in the schema
func (p *MiseSchemaParser) ValidatePath(path string) bool {
	return p.parser.ValidatePath(path)
}

// GetAllPaths returns all valid configuration paths from the schema
func (p *MiseSchemaParser) GetAllPaths() []string {
	return p.parser.GetAllPaths()
}

// GetAllSettingsWithInfo returns all settings with their schema information
func (p *MiseSchemaParser) GetAllSettingsWithInfo() []configschema.SettingInfo {
	return p.parser.GetAllSettingsWithInfo()
}

// ValidateValue checks if a value conforms to the mise schema at the given path
func (p *MiseSchemaParser) ValidateValue(path string, value any) error {
	return p.parser.ValidateValue(path, value)
}

// ValidateDocument validates a full mise configuration map against the schema
func (p *MiseSchemaParser) ValidateDocument(values map[string]any) error {
	return p.parser.ValidateDocument(values)
}
