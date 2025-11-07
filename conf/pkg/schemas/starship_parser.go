package schemas

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/neongreen/mono/lib/configschema"
)

// StarshipSchemaParser handles parsing of starship JSON schema for completion data
// It wraps JSONSchemaParser which uses the jsonschema library properly
type StarshipSchemaParser struct {
	// Legacy field kept for compatibility, but not used
	schema map[string]any
	// New field using proper jsonschema library
	parser *configschema.JSONSchemaParser
}

// NewStarshipSchemaParser creates a new starship schema parser using the jsonschema library
func NewStarshipSchemaParser() (*StarshipSchemaParser, error) {
	// Use the new jsonschema-based parser
	parser, err := CompileStarshipSchema()
	if err != nil {
		// Fall back to manual parsing if compilation fails
		var schema map[string]any
		if unmarshalErr := json.Unmarshal([]byte(StarshipSchema), &schema); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse starship schema JSON: %w (jsonschema error: %w)", unmarshalErr, err)
		}
		return &StarshipSchemaParser{
			schema: schema,
			parser: nil,
		}, nil
	}

	return &StarshipSchemaParser{
		schema: nil,
		parser: parser,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *StarshipSchemaParser) GetCompletionOptions(path string) []configschema.CompletionOption {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.GetCompletionOptions(path)
	}

	// Fallback to legacy implementation
	var options []configschema.CompletionOption

	// Get the properties from the schema
	properties, ok := p.schema["properties"].(map[string]any)
	if !ok {
		return options
	}

	if path == "" {
		// Return top-level properties
		for name, prop := range properties {
			if propMap, ok := prop.(map[string]any); ok {
				option := configschema.CompletionOption{
					Name:        name,
					Type:        getTypeFromProperty(propMap),
					Description: getDescriptionFromProperty(propMap),
				}
				options = append(options, option)
			}
		}
	} else {
		// Navigate to the nested property
		options = p.getNestedCompletionOptions(properties, path)
	}

	// Sort options by name for consistent output
	sort.Slice(options, func(i, j int) bool {
		return options[i].Name < options[j].Name
	})

	return options
}

// getNestedCompletionOptions navigates through nested properties to find completion options
func (p *StarshipSchemaParser) getNestedCompletionOptions(properties map[string]any, path string) []configschema.CompletionOption {
	var options []configschema.CompletionOption

	parts := strings.Split(path, ".")
	current := properties

	// Navigate through the path
	for i, part := range parts {
		if prop, exists := current[part]; exists {
			if propMap, ok := prop.(map[string]any); ok {
				if i == len(parts)-1 {
					// This is the final part - return its sub-properties
					if nestedProps, ok := propMap["properties"].(map[string]any); ok {
						for name, nestedProp := range nestedProps {
							if nestedPropMap, ok := nestedProp.(map[string]any); ok {
								option := configschema.CompletionOption{
									Name:        name,
									Type:        getTypeFromProperty(nestedPropMap),
									Description: getDescriptionFromProperty(nestedPropMap),
								}
								options = append(options, option)
							}
						}
					}
					return options
				} else {
					// Continue navigating
					if nestedProps, ok := propMap["properties"].(map[string]any); ok {
						current = nestedProps
					} else {
						return options // Can't navigate further
					}
				}
			} else {
				return options // Invalid property structure
			}
		} else {
			return options // Property doesn't exist
		}
	}

	return options
}

// GetAllPaths returns all possible dotted paths in the schema
func (p *StarshipSchemaParser) GetAllPaths() []string {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.GetAllPaths()
	}

	// Fallback to legacy implementation
	var paths []string

	properties, ok := p.schema["properties"].(map[string]any)
	if !ok {
		return paths
	}

	paths = p.collectAllPaths(properties, "")
	sort.Strings(paths)
	return paths
}

// collectAllPaths recursively collects all possible paths
func (p *StarshipSchemaParser) collectAllPaths(properties map[string]any, prefix string) []string {
	var paths []string

	for name, prop := range properties {
		var currentPath string
		if prefix == "" {
			currentPath = name
		} else {
			currentPath = prefix + "." + name
		}

		paths = append(paths, currentPath)

		if propMap, ok := prop.(map[string]any); ok {
			if nestedProps, ok := propMap["properties"].(map[string]any); ok {
				nestedPaths := p.collectAllPaths(nestedProps, currentPath)
				paths = append(paths, nestedPaths...)
			}
		}
	}

	return paths
}

// ValidatePath checks if a given dotted path exists in the schema
func (p *StarshipSchemaParser) ValidatePath(path string) bool {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.ValidatePath(path)
	}

	// Fallback to legacy implementation
	if path == "" {
		return true
	}

	properties, ok := p.schema["properties"].(map[string]any)
	if !ok {
		return false
	}

	parts := strings.Split(path, ".")
	current := properties
	var currentPropSchema map[string]any

	for i, part := range parts {
		// First, try to find the property in the current properties map
		if prop, exists := current[part]; exists {
			if propMap, ok := prop.(map[string]any); ok {
				currentPropSchema = propMap
				if i == len(parts)-1 {
					// This is the final part - it's valid
					return true
				}
				// Continue navigating if there are nested properties
				if nestedProps, ok := propMap["properties"].(map[string]any); ok {
					current = nestedProps
					continue
				}
				// No explicit nested properties - check if this property allows additionalProperties
				if additionalProps, ok := propMap["additionalProperties"].(map[string]any); ok {
					// The property has additionalProperties - check if the remaining path is valid
					// under the additionalProperties schema
					return p.validateAgainstSchema(additionalProps, parts[i+1:])
				}
				// No nested properties and no additionalProperties - invalid
				return false
			}
			return false
		}

		// Property doesn't exist in explicit properties - check if we're inside a property with additionalProperties
		if currentPropSchema != nil {
			if additionalProps, ok := currentPropSchema["additionalProperties"].(map[string]any); ok {
				// Current property allows additionalProperties - validate remaining path against that schema
				return p.validateAgainstSchema(additionalProps, parts[i:])
			}
		}

		// Not found and no additionalProperties - invalid
		return false
	}

	return true
}

// validateAgainstSchema validates a path against a schema definition
// This is used to validate paths under additionalProperties
// parts[0] is the current property name to validate, parts[1:] are the remaining path segments
func (p *StarshipSchemaParser) validateAgainstSchema(schema map[string]any, parts []string) bool {
	if len(parts) == 0 {
		return true
	}

	// We're in an additionalProperties context, so parts[0] (the current property name) is always valid
	// We just need to check if there are more parts and if they can be validated against this schema

	if len(parts) == 1 {
		// This is the last part - it's valid
		return true
	}

	// More parts remaining - check if the schema defines a structure that allows deeper nesting
	// First, check if the schema has explicit properties
	if nestedProps, ok := schema["properties"].(map[string]any); ok {
		// The schema has explicit properties - check if parts[1] is one of them
		if prop, exists := nestedProps[parts[1]]; exists {
			if propMap, ok := prop.(map[string]any); ok {
				// Found the property - continue validation with remaining parts
				return p.validateAgainstSchema(propMap, parts[2:])
			}
			// Property exists but not a map - can't nest further
			return false
		}
		// parts[1] not in explicit properties - check if schema has additionalProperties
		if additionalProps, ok := schema["additionalProperties"].(map[string]any); ok {
			return p.validateAgainstSchema(additionalProps, parts[1:])
		}
		// No additionalProperties and parts[1] not in properties - invalid
		return false
	}

	// No explicit properties - check if schema has additionalProperties
	if additionalProps, ok := schema["additionalProperties"].(map[string]any); ok {
		// Schema allows any nested property - validate remaining path against additionalProperties
		return p.validateAgainstSchema(additionalProps, parts[1:])
	}

	// Check the schema type
	if schemaType, ok := schema["type"].(string); ok {
		// Non-object types don't support nested properties
		if schemaType != "object" {
			return false
		}
		// Object type with no properties or additionalProperties - no nesting allowed
		return false
	}

	// Unknown schema structure - reject nesting
	return false
}

// GetPropertyInfo returns detailed information about a specific property
func (p *StarshipSchemaParser) GetPropertyInfo(path string) (configschema.PropertyInfo, error) {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.GetPropertyInfo(path)
	}

	// Fallback to legacy implementation
	if path == "" {
		return configschema.PropertyInfo{}, fmt.Errorf("empty path")
	}

	properties, ok := p.schema["properties"].(map[string]any)
	if !ok {
		return configschema.PropertyInfo{}, fmt.Errorf("no properties found in schema")
	}

	parts := strings.Split(path, ".")
	current := properties

	for i, part := range parts {
		if prop, exists := current[part]; exists {
			if propMap, ok := prop.(map[string]any); ok {
				if i == len(parts)-1 {
					// This is the target property
					return configschema.PropertyInfo{
						Name:        part,
						Type:        getTypeFromProperty(propMap),
						Description: getDescriptionFromProperty(propMap),
						Default:     getDefaultFromProperty(propMap),
						Enum:        getEnumFromProperty(propMap),
					}, nil
				} else {
					// Continue navigating
					if nestedProps, ok := propMap["properties"].(map[string]any); ok {
						current = nestedProps
					} else {
						return configschema.PropertyInfo{}, fmt.Errorf("cannot navigate to %s: no nested properties", part)
					}
				}
			} else {
				return configschema.PropertyInfo{}, fmt.Errorf("invalid property structure at %s", part)
			}
		} else {
			return configschema.PropertyInfo{}, fmt.Errorf("property %s not found", part)
		}
	}

	return configschema.PropertyInfo{}, fmt.Errorf("unexpected end of path navigation")
}

// GetAllSettingsWithInfo returns comprehensive information about all settings in the schema
func (p *StarshipSchemaParser) GetAllSettingsWithInfo() []configschema.SettingInfo {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.GetAllSettingsWithInfo()
	}

	// Fallback to legacy implementation
	var settings []configschema.SettingInfo

	properties, ok := p.schema["properties"].(map[string]any)
	if !ok {
		return settings
	}

	settings = p.collectAllSettingsWithInfo(properties, "")

	// Sort by path for consistent output
	sort.Slice(settings, func(i, j int) bool {
		return settings[i].Path < settings[j].Path
	})

	return settings
}

// collectAllSettingsWithInfo recursively collects all settings with their information
func (p *StarshipSchemaParser) collectAllSettingsWithInfo(properties map[string]any, prefix string) []configschema.SettingInfo {
	var settings []configschema.SettingInfo

	for name, prop := range properties {
		var currentPath string
		if prefix == "" {
			currentPath = name
		} else {
			currentPath = prefix + "." + name
		}

		if propMap, ok := prop.(map[string]any); ok {
			// Add this setting to the list
			setting := configschema.SettingInfo{
				Path:        currentPath,
				Type:        getTypeFromProperty(propMap),
				Description: getDescriptionFromProperty(propMap),
				Default:     getDefaultFromProperty(propMap),
				Enum:        getEnumFromProperty(propMap),
				// CurrentValue and IsSet will be filled in by the caller
			}
			settings = append(settings, setting)

			// Recursively collect nested settings
			if nestedProps, ok := propMap["properties"].(map[string]any); ok {
				nestedSettings := p.collectAllSettingsWithInfo(nestedProps, currentPath)
				settings = append(settings, nestedSettings...)
			}
		}
	}

	return settings
}
