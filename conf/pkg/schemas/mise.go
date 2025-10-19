package schemas

import (
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// MiseSchemaField represents a field in the mise configuration schema
type MiseSchemaField struct {
	Type        string                    `toml:"type"`
	Description string                    `toml:"description"`
	Properties  []MiseSchemaFieldProperty `toml:"properties,omitempty"`
}

// MiseSchemaFieldProperty represents a property within a schema field
type MiseSchemaFieldProperty struct {
	Name        string `toml:"name"`
	Type        string `toml:"type"`
	Description string `toml:"description"`
}

// MiseSchema represents the complete mise configuration schema
type MiseSchema struct {
	Schema struct {
		Name        string `toml:"name"`
		Description string `toml:"description"`
		Version     string `toml:"version"`
	} `toml:"schema"`
	Fields map[string]MiseSchemaField `toml:"fields"`
}

// LoadMiseSchema loads and parses the mise schema
func LoadMiseSchema() (*MiseSchema, error) {
	var schema MiseSchema
	if err := toml.Unmarshal([]byte(MiseSchemaData), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse mise schema: %w", err)
	}
	return &schema, nil
}

// GetCompletionOptions returns completion options for a given path
func (s *MiseSchema) GetCompletionOptions(path string) []CompletionOption {
	var options []CompletionOption

	// If path is empty, return top-level fields
	if path == "" {
		for name, field := range s.Fields {
			options = append(options, CompletionOption{
				Name:        name,
				Type:        field.Type,
				Description: field.Description,
			})
		}
		return options
	}

	// Handle nested paths
	parts := strings.Split(path, ".")
	topLevel := parts[0]
	field, exists := s.Fields[topLevel]
	if !exists {
		return options
	}

	if len(parts) == 1 {
		// Return properties of this field
		for _, prop := range field.Properties {
			options = append(options, CompletionOption{
				Name:        prop.Name,
				Type:        prop.Type,
				Description: prop.Description,
			})
		}
	}

	return options
}

// CompletionOption represents a completion option
type CompletionOption struct {
	Name        string
	Type        string
	Description string
}
