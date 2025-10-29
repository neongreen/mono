package editors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/creachadair/tomledit/parser"
)

// JSONEditor provides surgical editing of JSON files
type JSONEditor struct {
	filePath string
	dryRun   bool
}

// NewJSONEditor creates a new JSON editor for the specified file
func NewJSONEditor(filePath string) *JSONEditor {
	return &JSONEditor{
		filePath: filePath,
		dryRun:   false,
	}
}

// NewJSONEditorWithDryRun creates a new JSON editor with dry-run mode
func NewJSONEditorWithDryRun(filePath string, dryRun bool) *JSONEditor {
	return &JSONEditor{
		filePath: filePath,
		dryRun:   dryRun,
	}
}

// SetDryRun enables or disables dry-run mode
func (e *JSONEditor) SetDryRun(dryRun bool) {
	e.dryRun = dryRun
}

// SetValue sets a value at the specified dotted path
func (e *JSONEditor) SetValue(path string, value interface{}) error {
	if e.dryRun {
		fmt.Printf("DRY RUN: Would set %s = %v in %s\n", path, value, e.filePath)
		return nil
	}

	// Read existing file content if it exists
	var data map[string]interface{}

	if _, err := os.Stat(e.filePath); err == nil {
		content, err := os.ReadFile(e.filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", e.filePath, err)
		}

		if err := json.Unmarshal(content, &data); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		data = make(map[string]interface{})
	}

	// Parse the path and set the value
	key, err := parser.ParseKey(path)
	if err != nil {
		return fmt.Errorf("failed to parse path %s: %w", path, err)
	}

	setNestedValue(data, key, value)

	// Ensure directory exists
	dir := filepath.Dir(e.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write back with pretty printing
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(e.filePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetValue retrieves a value at the specified dotted path
func (e *JSONEditor) GetValue(path string) (interface{}, error) {
	content, err := os.ReadFile(e.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			if e.dryRun {
				fmt.Printf("DRY RUN: File %s does not exist\n", e.filePath)
			}
			return nil, fmt.Errorf("file does not exist: %s", e.filePath)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Parse the path and get the value
	key, err := parser.ParseKey(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %s: %w", path, err)
	}

	return getNestedValue(data, key)
}

// UnsetValue removes a value at the specified dotted path
func (e *JSONEditor) UnsetValue(path string) error {
	if e.dryRun {
		fmt.Printf("DRY RUN: Would unset %s in %s\n", path, e.filePath)
		return nil
	}

	content, err := os.ReadFile(e.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Parse the path and unset the value
	key, err := parser.ParseKey(path)
	if err != nil {
		return fmt.Errorf("failed to parse path %s: %w", path, err)
	}

	if !unsetNestedValue(data, key) {
		return fmt.Errorf("path not found: %s", path)
	}

	// Write back with pretty printing
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(e.filePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// PreviewSetValue shows what setting a value would do without doing it
func (e *JSONEditor) PreviewSetValue(path string, value interface{}) (string, error) {
	// Read existing file content if it exists
	var data map[string]interface{}

	if _, err := os.Stat(e.filePath); err == nil {
		content, err := os.ReadFile(e.filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", e.filePath, err)
		}

		if err := json.Unmarshal(content, &data); err != nil {
			return "", fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		data = make(map[string]interface{})
	}

	// Parse the path and set the value
	key, err := parser.ParseKey(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse path %s: %w", path, err)
	}

	setNestedValue(data, key, value)

	// Return pretty-printed JSON
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(output), nil
}

// PreviewUnsetValue shows what unsetting a value would do without doing it
func (e *JSONEditor) PreviewUnsetValue(path string) (string, error) {
	content, err := os.ReadFile(e.filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Parse the path and unset the value
	key, err := parser.ParseKey(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse path %s: %w", path, err)
	}

	if !unsetNestedValue(data, key) {
		return "", fmt.Errorf("path not found: %s", path)
	}

	// Return pretty-printed JSON
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(output), nil
}

// GetAllValues returns all configuration values as a nested map
func (e *JSONEditor) GetAllValues() (map[string]interface{}, error) {
	content, err := os.ReadFile(e.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return data, nil
}

// SetAllValues sets multiple configuration values from a nested map structure
func (e *JSONEditor) SetAllValues(values map[string]interface{}) error {
	if e.dryRun {
		fmt.Println("DRY RUN: Would set all values")
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(e.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write with pretty printing
	output, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(e.filePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Helper functions for nested map operations

func setNestedValue(m map[string]interface{}, key parser.Key, value interface{}) {
	if len(key) == 0 {
		return
	}

	current := m
	for i := 0; i < len(key)-1; i++ {
		part := key[i]
		next, ok := current[part]
		if !ok {
			nextMap := make(map[string]interface{})
			current[part] = nextMap
			current = nextMap
			continue
		}

		nextMap, ok := next.(map[string]interface{})
		if !ok {
			nextMap = make(map[string]interface{})
			current[part] = nextMap
		}
		current = nextMap
	}

	current[key[len(key)-1]] = value
}

func getNestedValue(m map[string]interface{}, key parser.Key) (interface{}, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("empty key")
	}

	current := m
	for i := 0; i < len(key); i++ {
		part := key[i]
		val, ok := current[part]
		if !ok {
			return nil, fmt.Errorf("key not found: %s", strings.Join(key[:i+1], "."))
		}

		if i == len(key)-1 {
			return val, nil
		}

		nextMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("path is not a map at: %s", strings.Join(key[:i+1], "."))
		}
		current = nextMap
	}

	return nil, fmt.Errorf("unexpected end of path")
}

func unsetNestedValue(m map[string]interface{}, key parser.Key) bool {
	if len(key) == 0 {
		return false
	}

	current := m
	stack := make([]map[string]interface{}, 0, len(key))
	stack = append(stack, current)

	for i := 0; i < len(key)-1; i++ {
		part := key[i]
		next, ok := current[part].(map[string]interface{})
		if !ok {
			return false
		}
		stack = append(stack, next)
		current = next
	}

	lastPart := key[len(key)-1]
	if _, exists := current[lastPart]; !exists {
		return false
	}
	delete(current, lastPart)

	// Cleanup empty maps from the bottom up
	for i := len(stack) - 1; i > 0; i-- {
		parent := stack[i-1]
		childKey := key[i-1]
		child, ok := parent[childKey].(map[string]interface{})
		if !ok {
			break
		}
		if len(child) == 0 {
			delete(parent, childKey)
		} else {
			break
		}
	}

	return true
}
