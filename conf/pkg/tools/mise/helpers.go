package mise

import "strings"

// lookupValueByPath looks up a value in a nested map using a dotted path
func lookupValueByPath(data map[string]any, path string) any {
	if path == "" {
		return nil
	}

	parts := splitPath(path)
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return nil
		}

		// If this is the last part, return the value
		if i == len(parts)-1 {
			return value
		}

		// Otherwise, expect a nested map and continue traversing
		nestedMap, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		current = nestedMap
	}

	return nil
}

// splitPath splits a dotted path into its component parts
func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, ".")
}
