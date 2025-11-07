package schemas

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/lib/configschema"
)

// MiseSchemaParser handles parsing of Mise JSON schema for completion data
// It wraps either JSONSchemaParser or LegacyJSONSchemaParser
type MiseSchemaParser struct {
	parser       *configschema.JSONSchemaParser
	legacyParser *configschema.LegacyJSONSchemaParser
}

// NewMiseSchemaParser creates a new Mise schema parser using the jsonschema library
func NewMiseSchemaParser() (*MiseSchemaParser, error) {
	// Try to use the new jsonschema-based parser
	parser, err := CompileMiseSchema()
	if err != nil {
		// Fall back to legacy parsing if compilation fails
		var schema map[string]any
		if unmarshalErr := json.Unmarshal([]byte(MiseJSONSchema), &schema); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse Mise schema JSON: %w (jsonschema error: %w)", unmarshalErr, err)
		}
		return &MiseSchemaParser{
			parser:       nil,
			legacyParser: configschema.NewLegacyJSONSchemaParser(schema),
		}, nil
	}

	return &MiseSchemaParser{
		parser:       parser,
		legacyParser: nil,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *MiseSchemaParser) GetCompletionOptions(path string) []configschema.CompletionOption {
	if p.parser != nil {
		return p.parser.GetCompletionOptions(path)
	}
	return p.legacyParser.GetCompletionOptions(path)
}

// ValidatePath checks if a configuration path exists in the schema
func (p *MiseSchemaParser) ValidatePath(path string) bool {
	if p.parser != nil {
		return p.parser.ValidatePath(path)
	}
	return p.legacyParser.ValidatePath(path)
}

// GetAllPaths returns all valid configuration paths from the schema
func (p *MiseSchemaParser) GetAllPaths() []string {
	if p.parser != nil {
		return p.parser.GetAllPaths()
	}
	return p.legacyParser.GetAllPaths()
}

// GetAllSettingsWithInfo returns all settings with their schema information
func (p *MiseSchemaParser) GetAllSettingsWithInfo() []configschema.SettingInfo {
	if p.parser != nil {
		return p.parser.GetAllSettingsWithInfo()
	}
	return p.legacyParser.GetAllSettingsWithInfo()
}
