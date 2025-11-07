package toml

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/creachadair/tomledit/parser"
	tomlv2 "github.com/pelletier/go-toml/v2"
)

// WriteFile writes the provided values to the given TOML file while preserving
// all existing comments and formatting. Missing parent directories are created
// automatically.
func WriteFile(path string, values map[string]any) error {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var doc *Document
	if err == nil {
		doc, err = Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse existing TOML %s: %w", path, err)
		}
	} else {
		doc, err = ParseString("")
		if err != nil {
			return fmt.Errorf("failed to create TOML document for %s: %w", path, err)
		}
	}

	if err := doc.ApplyMap(values); err != nil {
		return fmt.Errorf("failed to apply values to %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	if err := os.WriteFile(path, doc.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// ApplyMap merges the provided nested map into the document, preserving
// any existing keys that are not present in values. This allows partial
// updates without affecting unmanaged configuration.
func (d *Document) ApplyMap(values map[string]any) error {
	if d == nil {
		return fmt.Errorf("cannot apply values to a nil document")
	}

	desired := flattenValues(values)

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(desired))
	for path := range desired {
		keys = append(keys, path)
	}
	sort.Strings(keys)

	// Set all desired values, preserving any existing keys not in desired
	for _, path := range keys {
		if err := d.Set(path, desired[path]); err != nil {
			return fmt.Errorf("failed to set %s: %w", path, err)
		}
	}

	return nil
}

// ReplaceMap updates the document with the provided nested map and removes any
// keys that are not present in values. Use ApplyMap instead if you want to
// preserve unmanaged keys.
func (d *Document) ReplaceMap(values map[string]any) error {
	if d == nil {
		return fmt.Errorf("cannot apply values to a nil document")
	}

	desired := flattenValues(values)

	var existing map[string]any
	if data := d.Bytes(); len(data) > 0 {
		if err := tomlv2.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("failed to parse existing document: %w", err)
		}
	}

	current := flattenValues(existing)

	// Delete keys that are not in desired values
	for path := range current {
		if _, ok := desired[path]; !ok {
			if err := d.Delete(path); err != nil {
				return fmt.Errorf("failed to delete %s: %w", path, err)
			}
		}
	}

	keys := make([]string, 0, len(desired))
	for path := range desired {
		keys = append(keys, path)
	}
	sort.Strings(keys)

	for _, path := range keys {
		if err := d.Set(path, desired[path]); err != nil {
			return fmt.Errorf("failed to set %s: %w", path, err)
		}
	}

	return nil
}

func flattenValues(values map[string]any) map[string]any {
	if values == nil {
		return map[string]any{}
	}

	result := make(map[string]any)
	normalized := normalizeValues(values)
	flattenRecursive(normalized, "", result)
	return result
}

func flattenRecursive(values map[string]any, prefix string, result map[string]any) {
	if values == nil {
		return
	}

	for key, val := range values {
		formattedKey := formatKeySegment(key)

		fullKey := formattedKey
		if prefix != "" {
			fullKey = prefix + "." + formattedKey
		}

		if nested, ok := val.(map[string]any); ok {
			flattenRecursive(nested, fullKey, result)
			continue
		}

		result[fullKey] = val
	}
}

func formatKeySegment(key string) string {
	if isBareKey(key) {
		return key
	}
	escaped := strings.ReplaceAll(key, `"`, `\"`)
	return `"` + escaped + `"`
}

func isBareKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	for _, r := range key {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func normalizeValues(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}

	normalized := make(map[string]any, len(values))
	for rawKey, rawVal := range values {
		normalizedVal := normalizeValue(rawVal)
		key, err := parser.ParseKey(rawKey)
		if err == nil && len(key) > 1 {
			setNestedValue(normalized, key, normalizedVal)
			continue
		}
		if err == nil && len(key) == 1 {
			normalized[key[0]] = normalizedVal
			continue
		}
		normalized[rawKey] = normalizedVal
	}
	return normalized
}

func normalizeValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return normalizeValues(v)
	case map[any]any:
		converted := make(map[string]any, len(v))
		for k, elem := range v {
			strKey, ok := k.(string)
			if !ok {
				continue
			}
			converted[strKey] = normalizeValue(elem)
		}
		return normalizeValues(converted)
	case []any:
		out := make([]any, len(v))
		for i, elem := range v {
			out[i] = normalizeValue(elem)
		}
		return out
	default:
		return value
	}
}

func setNestedValue(m map[string]any, key parser.Key, value any) {
	if len(key) == 0 {
		return
	}

	current := ensureMap(m)
	for i := 0; i < len(key)-1; i++ {
		part := key[i]
		next, ok := current[part]
		if !ok {
			nextMap := make(map[string]any)
			current[part] = nextMap
			current = nextMap
			continue
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			nextMap = make(map[string]any)
			current[part] = nextMap
		}
		current = nextMap
	}

	current[key[len(key)-1]] = normalizeValue(value)
}

func ensureMap(m map[string]any) map[string]any {
	if m == nil {
		return make(map[string]any)
	}
	return m
}
