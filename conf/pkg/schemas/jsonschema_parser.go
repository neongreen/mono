package schemas

import (
	"fmt"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// JSONSchemaParser provides universal JSON schema parsing using the jsonschema library.
// It replaces manual map[string]interface{} navigation with proper schema introspection.
type JSONSchemaParser struct {
	schema *jsonschema.Schema
}

// NewJSONSchemaParser creates a new parser from a compiled jsonschema.Schema
func NewJSONSchemaParser(schema *jsonschema.Schema) *JSONSchemaParser {
	return &JSONSchemaParser{
		schema: schema,
	}
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *JSONSchemaParser) GetCompletionOptions(path string) []CompletionOption {
	var options []CompletionOption

	if path == "" {
		// Return top-level properties
		if p.schema.Properties != nil {
			for name, propSchema := range p.schema.Properties {
				options = append(options, CompletionOption{
					Name:        name,
					Type:        getSchemaType(propSchema),
					Description: propSchema.Description,
				})
			}
		}
	} else {
		// Navigate to the nested property
		targetSchema := p.navigateToPath(path)
		if targetSchema != nil && targetSchema.Properties != nil {
			for name, propSchema := range targetSchema.Properties {
				options = append(options, CompletionOption{
					Name:        name,
					Type:        getSchemaType(propSchema),
					Description: propSchema.Description,
				})
			}
		}
	}

	// Sort options by name for consistent output
	sort.Slice(options, func(i, j int) bool {
		return options[i].Name < options[j].Name
	})

	return options
}

// GetAllPaths returns all possible dotted paths in the schema
func (p *JSONSchemaParser) GetAllPaths() []string {
	var paths []string
	p.collectPaths(p.schema, "", &paths)
	sort.Strings(paths)
	return paths
}

// collectPaths recursively collects all paths from the schema
func (p *JSONSchemaParser) collectPaths(schema *jsonschema.Schema, prefix string, paths *[]string) {
	if schema == nil || schema.Properties == nil {
		return
	}

	for name, propSchema := range schema.Properties {
		fullPath := name
		if prefix != "" {
			fullPath = prefix + "." + name
		}
		*paths = append(*paths, fullPath)

		// Recursively collect nested paths
		p.collectPaths(propSchema, fullPath, paths)
	}
}

// ValidatePath checks if a given dotted path exists in the schema
func (p *JSONSchemaParser) ValidatePath(path string) bool {
	if path == "" {
		return true
	}

	return p.navigateToPath(path) != nil
}

// navigateToPath navigates to a schema at the given dotted path
// Returns nil if the path doesn't exist
func (p *JSONSchemaParser) navigateToPath(path string) *jsonschema.Schema {
	if path == "" {
		return p.schema
	}

	parts := strings.Split(path, ".")
	current := p.schema

	for i, part := range parts {
		// Follow $ref if present
		if current.Ref != nil {
			current = current.Ref
		}

		if current.Properties == nil {
			// No explicit properties - check additionalProperties
			if additionalProps := p.getAdditionalPropertiesSchema(current); additionalProps != nil {
				// AdditionalProperties allows any nested structure
				// Continue with the additionalProperties schema for remaining parts
				current = additionalProps
				continue
			}
			return nil
		}

		// Look for the property in explicit properties
		propSchema, exists := current.Properties[part]
		if !exists {
			// Property not found in explicit properties - check additionalProperties
			if additionalProps := p.getAdditionalPropertiesSchema(current); additionalProps != nil {
				// AdditionalProperties allows this property
				current = additionalProps
				continue
			}
			return nil
		}

		// Found the property
		if i == len(parts)-1 {
			// This is the final part - follow $ref if present before returning
			if propSchema.Ref != nil {
				return propSchema.Ref
			}
			return propSchema
		}

		// Continue navigating
		current = propSchema
	}

	return current
}

// getAdditionalPropertiesSchema extracts the schema from AdditionalProperties if it exists
func (p *JSONSchemaParser) getAdditionalPropertiesSchema(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema.AdditionalProperties == nil {
		return nil
	}

	// AdditionalProperties can be bool or *Schema
	switch ap := schema.AdditionalProperties.(type) {
	case bool:
		if ap {
			// true means any additional properties are allowed
			// Return an empty schema that allows anything
			return &jsonschema.Schema{}
		}
		return nil
	case *jsonschema.Schema:
		return ap
	default:
		return nil
	}
}

// GetPropertyInfo returns detailed information about a specific property
func (p *JSONSchemaParser) GetPropertyInfo(path string) (PropertyInfo, error) {
	if path == "" {
		return PropertyInfo{}, fmt.Errorf("empty path")
	}

	schema := p.navigateToPath(path)
	if schema == nil {
		return PropertyInfo{}, fmt.Errorf("property %s not found", path)
	}

	// Extract the property name (last part of the path)
	parts := strings.Split(path, ".")
	name := parts[len(parts)-1]

	return PropertyInfo{
		Name:        name,
		Type:        getSchemaType(schema),
		Description: schema.Description,
		Default:     getSchemaDefault(schema),
		Enum:        getSchemaEnum(schema),
	}, nil
}

// GetAllSettingsWithInfo returns comprehensive information about all settings in the schema
func (p *JSONSchemaParser) GetAllSettingsWithInfo() []SettingInfo {
	var settings []SettingInfo
	p.collectSettings(p.schema, "", &settings)

	// Sort by path for consistent output
	sort.Slice(settings, func(i, j int) bool {
		return settings[i].Path < settings[j].Path
	})

	return settings
}

// collectSettings recursively collects all settings with their information
func (p *JSONSchemaParser) collectSettings(schema *jsonschema.Schema, prefix string, settings *[]SettingInfo) {
	if schema == nil || schema.Properties == nil {
		return
	}

	for name, propSchema := range schema.Properties {
		fullPath := name
		if prefix != "" {
			fullPath = prefix + "." + name
		}

		setting := SettingInfo{
			Path:        fullPath,
			Type:        getSchemaType(propSchema),
			Description: propSchema.Description,
			Default:     getSchemaDefault(propSchema),
			Enum:        getSchemaEnum(propSchema),
			// CurrentValue and IsSet will be filled in by the caller
		}
		*settings = append(*settings, setting)

		// Recursively collect nested settings
		p.collectSettings(propSchema, fullPath, settings)
	}
}

// Helper functions to extract information from jsonschema.Schema

func getSchemaType(schema *jsonschema.Schema) string {
	if schema == nil {
		return "unknown"
	}

	// Handle boolean schemas (Always field in v5)
	if schema.Always != nil {
		if *schema.Always {
			return "any"
		}
		return "never"
	}

	// Get type from Types field (slice of strings in v5)
	if len(schema.Types) > 0 {
		return schema.Types[0] // Return the first type
	}

	// Check if it's an object with properties
	if schema.Properties != nil {
		return "object"
	}

	return "unknown"
}

func getSchemaDefault(schema *jsonschema.Schema) any {
	if schema == nil {
		return nil
	}
	return schema.Default
}

func getSchemaEnum(schema *jsonschema.Schema) []string {
	if schema == nil || len(schema.Enum) == 0 {
		return nil
	}

	var enumStrings []string
	for _, value := range schema.Enum {
		if str, ok := value.(string); ok {
			enumStrings = append(enumStrings, str)
		}
	}
	return enumStrings
}
