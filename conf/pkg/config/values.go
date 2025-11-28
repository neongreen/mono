package config

import (
	"strings"
	"unicode"

	"github.com/creachadair/tomledit/parser"
)

// FlattenValues converts nested configuration maps to dotted-path representation.
// Keys are quoted when necessary to produce TOML-compliant paths (e.g., aliases.".")
func FlattenValues(values map[string]any) map[string]any {
	result := make(map[string]any)
	normalized := normalizeValues(values)
	flattenRecursive(normalized, "", result)
	return result
}

// ExpandValues converts a map of dotted paths back into a nested map structure.
func ExpandValues(flat map[string]any) map[string]any {
	if flat == nil {
		return nil
	}

	result := make(map[string]any)
	for path, val := range flat {
		key, err := parser.ParseKey(path)
		if err != nil || len(key) == 0 {
			result[path] = normalizeValue(val)
			continue
		}
		setNestedValue(result, key, val)
	}
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

func ensureMap(m map[string]any) map[string]any {
	if m == nil {
		return make(map[string]any)
	}
	return m
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

func mergeNestedValues(dst map[string]any, src map[string]any) map[string]any {
	dst = normalizeValues(dst)
	if dst == nil {
		dst = make(map[string]any)
	}

	srcNormalized := normalizeValues(src)
	if srcNormalized == nil {
		return dst
	}

	for key, val := range srcNormalized {
		if existing, ok := dst[key]; ok {
			existingMap, existingIsMap := existing.(map[string]any)
			newMap, newIsMap := val.(map[string]any)
			if existingIsMap && newIsMap {
				dst[key] = mergeNestedValues(existingMap, newMap)
				continue
			}
		}
		dst[key] = val
	}

	return dst
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

func getNestedValue(m map[string]any, key parser.Key) (any, bool) {
	if len(key) == 0 {
		return nil, false
	}

	current := m
	for i := range key {
		part := key[i]
		val, ok := current[part]
		if !ok {
			return nil, false
		}

		if i == len(key)-1 {
			return val, true
		}

		nextMap, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		current = nextMap
	}

	return nil, false
}

func unsetNestedValue(m map[string]any, key parser.Key) bool {
	if len(key) == 0 {
		return false
	}

	current := m
	stack := make([]map[string]any, 0, len(key))
	stack = append(stack, current)

	for i := 0; i < len(key)-1; i++ {
		part := key[i]
		next, ok := current[part].(map[string]any)
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
		child, ok := parent[childKey].(map[string]any)
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
