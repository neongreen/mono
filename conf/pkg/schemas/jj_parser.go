package schemas

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// JJSchemaParser handles parsing of jj JSON schema for completion data
// It now wraps JSONSchemaParser which uses the jsonschema library properly
type JJSchemaParser struct {
	// Legacy field kept for compatibility, but not used
	schema map[string]interface{}
	// New field using proper jsonschema library
	parser *JSONSchemaParser
}

// NewJJSchemaParser creates a new jj schema parser using the jsonschema library
func NewJJSchemaParser() (*JJSchemaParser, error) {
	// Use the new jsonschema-based parser
	parser, err := CompileJJSchema()
	if err != nil {
		// Fall back to manual parsing if compilation fails
		var schema map[string]interface{}
		if unmarshalErr := json.Unmarshal([]byte(JJSchema), &schema); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse jj schema JSON: %w (jsonschema error: %v)", unmarshalErr, err)
		}
		return &JJSchemaParser{
			schema: schema,
			parser: nil,
		}, nil
	}

	return &JJSchemaParser{
		schema: nil,
		parser: parser,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *JJSchemaParser) GetCompletionOptions(path string) []CompletionOption {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.GetCompletionOptions(path)
	}

	// Fallback to legacy implementation
	var options []CompletionOption

	// Get the properties from the schema
	properties, ok := p.schema["properties"].(map[string]interface{})
	if !ok {
		return options
	}

	if path == "" {
		// Return top-level properties
		for name, prop := range properties {
			if propMap, ok := prop.(map[string]interface{}); ok {
				option := CompletionOption{
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
func (p *JJSchemaParser) getNestedCompletionOptions(properties map[string]interface{}, path string) []CompletionOption {
	var options []CompletionOption

	parts := strings.Split(path, ".")
	current := properties

	// Navigate through the path
	for i, part := range parts {
		if prop, exists := current[part]; exists {
			if propMap, ok := prop.(map[string]interface{}); ok {
				if i == len(parts)-1 {
					// This is the final part - return its sub-properties
					if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
						for name, nestedProp := range nestedProps {
							if nestedPropMap, ok := nestedProp.(map[string]interface{}); ok {
								option := CompletionOption{
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
					if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
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
func (p *JJSchemaParser) GetAllPaths() []string {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.GetAllPaths()
	}

	// Fallback to legacy implementation
	var paths []string

	properties, ok := p.schema["properties"].(map[string]interface{})
	if !ok {
		return paths
	}

	paths = p.collectAllPaths(properties, "")
	sort.Strings(paths)
	return paths
}

// collectAllPaths recursively collects all possible paths
func (p *JJSchemaParser) collectAllPaths(properties map[string]interface{}, prefix string) []string {
	var paths []string

	for name, prop := range properties {
		var currentPath string
		if prefix == "" {
			currentPath = name
		} else {
			currentPath = prefix + "." + name
		}

		paths = append(paths, currentPath)

		if propMap, ok := prop.(map[string]interface{}); ok {
			if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
				nestedPaths := p.collectAllPaths(nestedProps, currentPath)
				paths = append(paths, nestedPaths...)
			}
		}
	}

	return paths
}

// ValidatePath checks if a given dotted path exists in the schema
func (p *JJSchemaParser) ValidatePath(path string) bool {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.ValidatePath(path)
	}

	// Fallback to legacy implementation
	if path == "" {
		return true
	}

	properties, ok := p.schema["properties"].(map[string]interface{})
	if !ok {
		return false
	}

	parts := strings.Split(path, ".")
	current := properties
	var currentPropSchema map[string]interface{}

	for i, part := range parts {
		// First, try to find the property in the current properties map
		if prop, exists := current[part]; exists {
			if propMap, ok := prop.(map[string]interface{}); ok {
				currentPropSchema = propMap
				if i == len(parts)-1 {
					// This is the final part - it's valid
					return true
				}
				// Continue navigating if there are nested properties
				if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
					current = nestedProps
					continue
				}
				// No explicit nested properties - check if this property allows additionalProperties
				if additionalProps, ok := propMap["additionalProperties"].(map[string]interface{}); ok {
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
			if additionalProps, ok := currentPropSchema["additionalProperties"].(map[string]interface{}); ok {
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
func (p *JJSchemaParser) validateAgainstSchema(schema map[string]interface{}, parts []string) bool {
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
	if nestedProps, ok := schema["properties"].(map[string]interface{}); ok {
		// The schema has explicit properties - check if parts[1] is one of them
		if prop, exists := nestedProps[parts[1]]; exists {
			if propMap, ok := prop.(map[string]interface{}); ok {
				// Found the property - continue validation with remaining parts
				return p.validateAgainstSchema(propMap, parts[2:])
			}
			// Property exists but not a map - can't nest further
			return false
		}
		// parts[1] not in explicit properties - check if schema has additionalProperties
		if additionalProps, ok := schema["additionalProperties"].(map[string]interface{}); ok {
			return p.validateAgainstSchema(additionalProps, parts[1:])
		}
		// No additionalProperties and parts[1] not in properties - invalid
		return false
	}

	// No explicit properties - check if schema has additionalProperties
	if additionalProps, ok := schema["additionalProperties"].(map[string]interface{}); ok {
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
func (p *JJSchemaParser) GetPropertyInfo(path string) (PropertyInfo, error) {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.GetPropertyInfo(path)
	}

	// Fallback to legacy implementation
	if path == "" {
		return PropertyInfo{}, fmt.Errorf("empty path")
	}

	properties, ok := p.schema["properties"].(map[string]interface{})
	if !ok {
		return PropertyInfo{}, fmt.Errorf("no properties found in schema")
	}

	parts := strings.Split(path, ".")
	current := properties

	for i, part := range parts {
		if prop, exists := current[part]; exists {
			if propMap, ok := prop.(map[string]interface{}); ok {
				if i == len(parts)-1 {
					// This is the target property
					return PropertyInfo{
						Name:        part,
						Type:        getTypeFromProperty(propMap),
						Description: getDescriptionFromProperty(propMap),
						Default:     getDefaultFromProperty(propMap),
						Enum:        getEnumFromProperty(propMap),
					}, nil
				} else {
					// Continue navigating
					if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
						current = nestedProps
					} else {
						return PropertyInfo{}, fmt.Errorf("cannot navigate to %s: no nested properties", part)
					}
				}
			} else {
				return PropertyInfo{}, fmt.Errorf("invalid property structure at %s", part)
			}
		} else {
			return PropertyInfo{}, fmt.Errorf("property %s not found", part)
		}
	}

	return PropertyInfo{}, fmt.Errorf("unexpected end of path navigation")
}

// PropertyInfo contains detailed information about a schema property
type PropertyInfo struct {
	Name        string
	Type        string
	Description string
	Default     interface{}
	Enum        []string
}

// SettingInfo contains comprehensive information about a setting including current value
type SettingInfo struct {
	Path         string
	Type         string
	Description  string
	Default      interface{}
	Enum         []string
	CurrentValue interface{}
	IsSet        bool
}

// GetAllSettingsWithInfo returns comprehensive information about all settings in the schema
func (p *JJSchemaParser) GetAllSettingsWithInfo() []SettingInfo {
	// Use the new parser if available
	if p.parser != nil {
		return p.parser.GetAllSettingsWithInfo()
	}

	// Fallback to legacy implementation
	var settings []SettingInfo

	properties, ok := p.schema["properties"].(map[string]interface{})
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
func (p *JJSchemaParser) collectAllSettingsWithInfo(properties map[string]interface{}, prefix string) []SettingInfo {
	var settings []SettingInfo

	for name, prop := range properties {
		var currentPath string
		if prefix == "" {
			currentPath = name
		} else {
			currentPath = prefix + "." + name
		}

		if propMap, ok := prop.(map[string]interface{}); ok {
			// Add this setting to the list
			setting := SettingInfo{
				Path:        currentPath,
				Type:        getTypeFromProperty(propMap),
				Description: getDescriptionFromProperty(propMap),
				Default:     getDefaultFromProperty(propMap),
				Enum:        getEnumFromProperty(propMap),
				// CurrentValue and IsSet will be filled in by the caller
			}
			settings = append(settings, setting)

			// Recursively collect nested settings
			if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
				nestedSettings := p.collectAllSettingsWithInfo(nestedProps, currentPath)
				settings = append(settings, nestedSettings...)
			}
		}
	}

	return settings
}

// Helper functions to extract information from property maps
func getTypeFromProperty(prop map[string]interface{}) string {
	if t, ok := prop["type"].(string); ok {
		return t
	}
	return "unknown"
}

func getDescriptionFromProperty(prop map[string]interface{}) string {
	if desc, ok := prop["description"].(string); ok {
		return desc
	}
	return ""
}

func getDefaultFromProperty(prop map[string]interface{}) interface{} {
	if def, ok := prop["default"]; ok {
		return def
	}
	return nil
}

func getEnumFromProperty(prop map[string]interface{}) []string {
	if enumInterface, ok := prop["enum"]; ok {
		if enumSlice, ok := enumInterface.([]interface{}); ok {
			var enumStrings []string
			for _, item := range enumSlice {
				if str, ok := item.(string); ok {
					enumStrings = append(enumStrings, str)
				}
			}
			return enumStrings
		}
	}
	return nil
}
