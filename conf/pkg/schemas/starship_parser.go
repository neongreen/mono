package schemas

import (
	"fmt"

	"github.com/neongreen/mono/lib/configschema"
)

// StarshipSchemaParser handles parsing of starship JSON schema for completion data
type StarshipSchemaParser struct {
	parser *configschema.JSONSchemaParser
}

// NewStarshipSchemaParser creates a new starship schema parser using the jsonschema library
func NewStarshipSchemaParser() (*StarshipSchemaParser, error) {
	parser, err := CompileStarshipSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to compile starship schema: %w", err)
	}

	return &StarshipSchemaParser{
		parser: parser,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *StarshipSchemaParser) GetCompletionOptions(path string) []configschema.CompletionOption {
	return p.parser.GetCompletionOptions(path)
}

// GetAllPaths returns all possible dotted paths in the schema
func (p *StarshipSchemaParser) GetAllPaths() []string {
	return p.parser.GetAllPaths()
}

// ValidatePath checks if a given dotted path exists in the schema
func (p *StarshipSchemaParser) ValidatePath(path string) bool {
	return p.parser.ValidatePath(path)
}

// GetPropertyInfo returns detailed information about a specific property
func (p *StarshipSchemaParser) GetPropertyInfo(path string) (configschema.PropertyInfo, error) {
	return p.parser.GetPropertyInfo(path)
}

// GetAllSettingsWithInfo returns comprehensive information about all settings in the schema
func (p *StarshipSchemaParser) GetAllSettingsWithInfo() []configschema.SettingInfo {
	return p.parser.GetAllSettingsWithInfo()
}
