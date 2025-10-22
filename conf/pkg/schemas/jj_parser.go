package schemas

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// JJSchemaParser handles parsing of jj JSON schema for completion data
type JJSchemaParser struct {
	schema map[string]interface{}
}

// NewJJSchemaParser creates a new jj schema parser
func NewJJSchemaParser() (*JJSchemaParser, error) {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(JJSchema), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse jj schema JSON: %w", err)
	}

	return &JJSchemaParser{
		schema: schema,
	}, nil
}

// GetCompletionOptions returns completion options for a given dotted path
func (p *JJSchemaParser) GetCompletionOptions(path string) []CompletionOption {
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
		if prop, exists := current[part]; exists {
			if propMap, ok := prop.(map[string]interface{}); ok {
				if i == len(parts)-1 {
					// This is the final part - it's valid regardless of whether it has nested properties
					return true
				} else {
					// Continue navigating if there are nested properties
					if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
						current = nestedProps
					} else {
						// No nested properties but more parts in path - invalid
						return false
					}
				}
			} else {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// GetPropertyInfo returns detailed information about a specific property
func (p *JJSchemaParser) GetPropertyInfo(path string) (PropertyInfo, error) {
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
