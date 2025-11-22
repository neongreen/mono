package schemas

import (
	"fmt"

	"github.com/neongreen/mono/lib/configschema"
)

// JJSchemaParser handles parsing of jj JSON schema for completion data
type JJSchemaParser struct {
	parser *configschema.JSONSchemaParser
}

// NewJJSchemaParser creates a new jj schema parser using the jsonschema library
func NewJJSchemaParser() (*JJSchemaParser, error) {
	parser, err := CompileJJSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to compile jj schema: %w", err)
	}

	return &JJSchemaParser{
		parser: parser,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *JJSchemaParser) GetCompletionOptions(path string) []configschema.CompletionOption {
	return p.parser.GetCompletionOptions(path)
}

// GetAllPaths returns all possible dotted paths in the schema
func (p *JJSchemaParser) GetAllPaths() []string {
	return p.parser.GetAllPaths()
}

// ValidatePath checks if a given dotted path exists in the schema
func (p *JJSchemaParser) ValidatePath(path string) bool {
	return p.parser.ValidatePath(path)
}

// GetPropertyInfo returns detailed information about a specific property
func (p *JJSchemaParser) GetPropertyInfo(path string) (configschema.PropertyInfo, error) {
	return p.parser.GetPropertyInfo(path)
}

// GetAllSettingsWithInfo returns comprehensive information about all settings in the schema
func (p *JJSchemaParser) GetAllSettingsWithInfo() []configschema.SettingInfo {
	return p.parser.GetAllSettingsWithInfo()
}

// ValidateValue checks if a value conforms to the jj schema at the given path
func (p *JJSchemaParser) ValidateValue(path string, value any) error {
	return p.parser.ValidateValue(path, value)
}

// ValidateDocument validates a full jj configuration map against the schema
func (p *JJSchemaParser) ValidateDocument(values map[string]any) error {
	return p.parser.ValidateDocument(values)
}
