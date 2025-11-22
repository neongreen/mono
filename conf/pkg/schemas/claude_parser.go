package schemas

import (
	"fmt"

	"github.com/neongreen/mono/lib/configschema"
)

// ClaudeSchemaParser handles parsing of Claude JSON schema for completion data
type ClaudeSchemaParser struct {
	parser *configschema.JSONSchemaParser
}

// NewClaudeSchemaParser creates a new Claude schema parser using the jsonschema library
func NewClaudeSchemaParser() (*ClaudeSchemaParser, error) {
	parser, err := CompileClaudeSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to compile Claude schema: %w", err)
	}

	return &ClaudeSchemaParser{
		parser: parser,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *ClaudeSchemaParser) GetCompletionOptions(path string) []configschema.CompletionOption {
	return p.parser.GetCompletionOptions(path)
}

// ValidatePath checks if a configuration path exists in the schema
func (p *ClaudeSchemaParser) ValidatePath(path string) bool {
	return p.parser.ValidatePath(path)
}

// GetAllPaths returns all valid configuration paths from the schema
func (p *ClaudeSchemaParser) GetAllPaths() []string {
	return p.parser.GetAllPaths()
}

// GetAllSettingsWithInfo returns all settings with their schema information
func (p *ClaudeSchemaParser) GetAllSettingsWithInfo() []configschema.SettingInfo {
	return p.parser.GetAllSettingsWithInfo()
}

// ValidateValue checks if a value conforms to the Claude schema at the given path
func (p *ClaudeSchemaParser) ValidateValue(path string, value any) error {
	return p.parser.ValidateValue(path, value)
}

// ValidateDocument validates a full Claude configuration map against the schema
func (p *ClaudeSchemaParser) ValidateDocument(values map[string]any) error {
	return p.parser.ValidateDocument(values)
}
