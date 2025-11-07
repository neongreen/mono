package schemas

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/lib/configschema"
)

// JJSchemaParser handles parsing of jj JSON schema for completion data
// It wraps either JSONSchemaParser or LegacyJSONSchemaParser
type JJSchemaParser struct {
	parser       *configschema.JSONSchemaParser
	legacyParser *configschema.LegacyJSONSchemaParser
}

// NewJJSchemaParser creates a new jj schema parser using the jsonschema library
func NewJJSchemaParser() (*JJSchemaParser, error) {
	// Try to use the new jsonschema-based parser
	parser, err := CompileJJSchema()
	if err != nil {
		// Fall back to legacy parsing if compilation fails
		var schema map[string]any
		if unmarshalErr := json.Unmarshal([]byte(JJSchema), &schema); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse jj schema JSON: %w (jsonschema error: %w)", unmarshalErr, err)
		}
		return &JJSchemaParser{
			parser:       nil,
			legacyParser: configschema.NewLegacyJSONSchemaParser(schema),
		}, nil
	}

	return &JJSchemaParser{
		parser:       parser,
		legacyParser: nil,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *JJSchemaParser) GetCompletionOptions(path string) []configschema.CompletionOption {
	if p.parser != nil {
		return p.parser.GetCompletionOptions(path)
	}
	return p.legacyParser.GetCompletionOptions(path)
}

// GetAllPaths returns all possible dotted paths in the schema
func (p *JJSchemaParser) GetAllPaths() []string {
	if p.parser != nil {
		return p.parser.GetAllPaths()
	}
	return p.legacyParser.GetAllPaths()
}

// ValidatePath checks if a given dotted path exists in the schema
func (p *JJSchemaParser) ValidatePath(path string) bool {
	if p.parser != nil {
		return p.parser.ValidatePath(path)
	}
	return p.legacyParser.ValidatePath(path)
}

// GetPropertyInfo returns detailed information about a specific property
func (p *JJSchemaParser) GetPropertyInfo(path string) (configschema.PropertyInfo, error) {
	if p.parser != nil {
		return p.parser.GetPropertyInfo(path)
	}
	return p.legacyParser.GetPropertyInfo(path)
}

// GetAllSettingsWithInfo returns comprehensive information about all settings in the schema
func (p *JJSchemaParser) GetAllSettingsWithInfo() []configschema.SettingInfo {
	if p.parser != nil {
		return p.parser.GetAllSettingsWithInfo()
	}
	return p.legacyParser.GetAllSettingsWithInfo()
}
