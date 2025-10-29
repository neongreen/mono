package schemas

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// MiseSchemaParser handles parsing of Mise JSON schema for completion data
// It now wraps JSONSchemaParser which uses the jsonschema library properly
type MiseSchemaParser struct {
	// Legacy field kept for compatibility, but not used
	schema map[string]interface{}
	// New field using proper jsonschema library
	parser *JSONSchemaParser
}

// NewMiseSchemaParser creates a new Mise schema parser using the jsonschema library
func NewMiseSchemaParser() (*MiseSchemaParser, error) {
	// Use the new jsonschema-based parser
	parser, err := CompileMiseSchema()
	if err != nil {
		// Fall back to manual parsing if compilation fails
		var schema map[string]interface{}
		if unmarshalErr := json.Unmarshal([]byte(MiseJSONSchema), &schema); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse Mise schema JSON: %w (jsonschema error: %v)", unmarshalErr, err)
		}
		return &MiseSchemaParser{
			schema: schema,
			parser: nil,
		}, nil
	}

	return &MiseSchemaParser{
		schema: nil,
		parser: parser,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *MiseSchemaParser) GetCompletionOptions(path string) []CompletionOption {
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
func (p *MiseSchemaParser) getNestedCompletionOptions(properties map[string]interface{}, path string) []CompletionOption {
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

// ValidatePath checks if a configuration path exists in the schema
func (p *MiseSchemaParser) ValidatePath(path string) bool {
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

	for i, part := range parts {
		prop, exists := current[part]
		if !exists {
			return false
		}

		if i == len(parts)-1 {
			// We've reached the final part and it exists
			return true
		}

		// Need to navigate deeper
		if propMap, ok := prop.(map[string]interface{}); ok {
			if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
				current = nestedProps
			} else {
				// Can't navigate further but the path continues
				return false
			}
		} else {
			return false
		}
	}

	return false
}

// GetAllPaths returns all valid configuration paths from the schema
func (p *MiseSchemaParser) GetAllPaths() []string {
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

	p.collectPaths(properties, "", &paths)

	// Sort for consistent output
	sort.Strings(paths)

	return paths
}

// collectPaths recursively collects all paths from the schema
func (p *MiseSchemaParser) collectPaths(properties map[string]interface{}, prefix string, paths *[]string) {
	for name, prop := range properties {
		fullPath := name
		if prefix != "" {
			fullPath = prefix + "." + name
		}

		*paths = append(*paths, fullPath)

		// Check if this property has nested properties
		if propMap, ok := prop.(map[string]interface{}); ok {
			if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
				p.collectPaths(nestedProps, fullPath, paths)
			}
		}
	}
}

// GetAllSettingsWithInfo returns all settings with their schema information
func (p *MiseSchemaParser) GetAllSettingsWithInfo() []SettingInfo {
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

	p.collectSettingsInfo(properties, "", &settings)

	// Sort by path
	sort.Slice(settings, func(i, j int) bool {
		return settings[i].Path < settings[j].Path
	})

	return settings
}

// collectSettingsInfo recursively collects setting information from the schema
func (p *MiseSchemaParser) collectSettingsInfo(properties map[string]interface{}, prefix string, settings *[]SettingInfo) {
	for name, prop := range properties {
		fullPath := name
		if prefix != "" {
			fullPath = prefix + "." + name
		}

		if propMap, ok := prop.(map[string]interface{}); ok {
			setting := SettingInfo{
				Path:        fullPath,
				Type:        getTypeFromProperty(propMap),
				Description: getDescriptionFromProperty(propMap),
			}

			// Get default value if present
			if defaultVal, ok := propMap["default"]; ok {
				setting.Default = defaultVal
			}

			*settings = append(*settings, setting)

			// Check if this property has nested properties
			if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
				p.collectSettingsInfo(nestedProps, fullPath, settings)
			}
		}
	}
}
