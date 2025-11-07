package schemas

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/lib/configschema"
)

// ClaudeSchemaParser handles parsing of Claude JSON schema for completion data
// It wraps either JSONSchemaParser or LegacyJSONSchemaParser
type ClaudeSchemaParser struct {
	parser       *configschema.JSONSchemaParser
	legacyParser *configschema.LegacyJSONSchemaParser
}

// NewClaudeSchemaParser creates a new Claude schema parser using the jsonschema library
func NewClaudeSchemaParser() (*ClaudeSchemaParser, error) {
	// Try to use the new jsonschema-based parser
	parser, err := CompileClaudeSchema()
	if err != nil {
		// Fall back to legacy parsing if compilation fails
		var schema map[string]any
		if unmarshalErr := json.Unmarshal([]byte(ClaudeSchema), &schema); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse Claude schema JSON: %w (jsonschema error: %w)", unmarshalErr, err)
		}
		return &ClaudeSchemaParser{
			parser:       nil,
			legacyParser: configschema.NewLegacyJSONSchemaParser(schema),
		}, nil
	}

	return &ClaudeSchemaParser{
		parser:       parser,
		legacyParser: nil,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *ClaudeSchemaParser) GetCompletionOptions(path string) []configschema.CompletionOption {
	if p.parser != nil {
		return p.parser.GetCompletionOptions(path)
	}
	return p.legacyParser.GetCompletionOptions(path)
}

// ValidatePath checks if a configuration path exists in the schema
func (p *ClaudeSchemaParser) ValidatePath(path string) bool {
	if p.parser != nil {
		return p.parser.ValidatePath(path)
	}
	return p.legacyParser.ValidatePath(path)
}

// GetAllPaths returns all valid configuration paths from the schema
func (p *ClaudeSchemaParser) GetAllPaths() []string {
	if p.parser != nil {
		return p.parser.GetAllPaths()
	}
	return p.legacyParser.GetAllPaths()
}

// GetAllSettingsWithInfo returns all settings with their schema information
func (p *ClaudeSchemaParser) GetAllSettingsWithInfo() []configschema.SettingInfo {
	if p.parser != nil {
		return p.parser.GetAllSettingsWithInfo()
	}
	return p.legacyParser.GetAllSettingsWithInfo()
}
