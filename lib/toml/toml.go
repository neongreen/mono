// Package tomlcp provides comment-preserving TOML parsing and serialization.
//
// This package wraps github.com/creachadair/tomledit to provide a stable,
// user-friendly API for reading, modifying, and writing TOML files while
// preserving all comments, formatting, and declaration order.
//
// The package supports TOML-compliant path parsing, including quoted keys
// for special characters. For example: aliases."." or section."key with spaces".
// Invalid paths like "aliases." (trailing dot) are rejected.
//
// Example usage:
//
//	doc, err := tomlcp.Parse([]byte(`
//	  # This is a comment
//	  [server]
//	  host = "localhost"
//	  port = 8080
//	`))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Modify a value
//	doc.Set("server.port", 9090)
//
//	// Use quoted keys for special characters
//	doc.Set(`aliases."."`, "status")
//
//	// Write back with comments preserved
//	output := doc.String()
package toml

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
)

// Document represents a parsed TOML document with preserved structure,
// including comments, formatting, and declaration order.
type Document struct {
	doc *tomledit.Document
}

// Parse parses a TOML document from bytes and returns a Document that
// preserves all structural information including comments.
func Parse(input []byte) (*Document, error) {
	doc, err := tomledit.Parse(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}
	return &Document{doc: doc}, nil
}

// ParseString is a convenience wrapper around Parse that accepts a string.
func ParseString(input string) (*Document, error) {
	return Parse([]byte(input))
}

// Get retrieves a value at the given dotted path (e.g., "server.host").
// Returns nil if the path doesn't exist.
func (d *Document) Get(path string) (interface{}, error) {
	keys, err := parseKeyPath(path)
	if err != nil {
		return nil, err
	}

	entry := d.doc.First(keys...)
	if entry == nil {
		return nil, nil // Path doesn't exist
	}

	if entry.KeyValue == nil {
		return nil, nil // Entry is a section, not a value
	}

	return parseValue(entry.KeyValue.Value)
}

// Set sets a value at the given dotted path, creating intermediate sections
// if necessary. The value can be a string, int, float64, bool, or []interface{}.
// This method preserves the original style (dotted keys vs sections, quote styles).
func (d *Document) Set(path string, value interface{}) error {
	keys, err := parseKeyPath(path)
	if err != nil {
		return err
	}

	// Check if the key already exists
	existingEntry := d.doc.First(keys...)

	if existingEntry != nil && existingEntry.KeyValue != nil {
		// Key exists - update it in place, preserving style and comments
		return d.updateExistingKey(existingEntry, value)
	}

	// Key doesn't exist - add it in an appropriate style
	return d.addNewKey(keys, value)
}

// updateExistingKey updates an existing key while preserving its formatting
func (d *Document) updateExistingKey(entry *tomledit.Entry, value interface{}) error {
	oldValue := entry.KeyValue.Value

	// Try to format the new value in the same style as the old value
	newValue, err := d.formatValuePreservingStyle(value, oldValue)
	if err != nil {
		return fmt.Errorf("failed to format value: %w", err)
	}

	// Update the value while preserving comments
	entry.KeyValue.Value = newValue
	entry.KeyValue.Value.Trailer = oldValue.Trailer // Preserve trailing comment

	return nil
}

// formatValuePreservingStyle formats a value trying to match the style of an existing value
func (d *Document) formatValuePreservingStyle(value interface{}, existingValue parser.Value) (parser.Value, error) {
	// For strings, try to preserve quote style
	if strValue, ok := value.(string); ok {
		if token, ok := existingValue.X.(parser.Token); ok {
			// Preserve the quote style from the original
			originalStr := token.String()

			var formatted string
			if strings.HasPrefix(originalStr, "'''") {
				// Multiline literal string
				formatted = "'''" + strValue + "'''"
			} else if strings.HasPrefix(originalStr, `"""`) {
				// Multiline basic string
				formatted = `"""` + strValue + `"""`
			} else if strings.HasPrefix(originalStr, "'") {
				// Single-quoted literal string
				formatted = "'" + strValue + "'"
			} else {
				// Double-quoted basic string (default)
				formatted = quoteString(strValue)
			}

			return parser.ParseValue(formatted)
		}
	}

	// For non-strings or when we can't preserve style, use default formatting
	valueStr, err := formatValueToString(value)
	if err != nil {
		return parser.Value{}, err
	}

	return parser.ParseValue(valueStr)
}

// addNewKey adds a new key in an appropriate style (dotted key vs section)
func (d *Document) addNewKey(keys parser.Key, value interface{}) error {
	// Format the value
	valueStr, err := formatValueToString(value)
	if err != nil {
		return fmt.Errorf("failed to format value: %w", err)
	}

	parsedValue, err := parser.ParseValue(valueStr)
	if err != nil {
		return fmt.Errorf("failed to parse value: %w", err)
	}

	// Determine where and how to add the new key
	if len(keys) == 1 {
		// Top-level key - add to global section
		return d.addToGlobalSection(keys, parsedValue)
	}

	// Multi-part key - determine if we should use dotted key or section style
	// Check if there are existing dotted keys with the same prefix
	if d.hasDottedKeysWithPrefix(keys[:len(keys)-1]) {
		// Use dotted key style to match existing style
		return d.addDottedKey(keys, parsedValue)
	}

	// Check if there's an existing section for this prefix
	tableName := keys[:len(keys)-1]
	entry := transform.FindTable(d.doc, tableName...)

	if entry != nil {
		// Section exists - add to it
		return d.addToSection(entry.Section, parser.Key{keys[len(keys)-1]}, parsedValue)
	}

	// No existing pattern. Prefer creating table sections for nested keys.
	if len(keys) > 1 {
		section, err := d.ensureSection(tableName)
		if err != nil {
			return err
		}
		return d.addToSection(section, parser.Key{keys[len(keys)-1]}, parsedValue)
	}

	// Fallback to global section for single-part keys.
	return d.addToGlobalSection(keys, parsedValue)
}

// hasDottedKeysWithPrefix checks if there are any dotted keys with the given prefix
func (d *Document) hasDottedKeysWithPrefix(prefix parser.Key) bool {
	if d.doc.Global == nil {
		return false
	}

	prefixStr := strings.Join(prefix, ".") + "."

	for _, item := range d.doc.Global.Items {
		if kv, ok := item.(*parser.KeyValue); ok {
			keyStr := kv.Name.String()
			if strings.HasPrefix(keyStr, prefixStr) {
				return true
			}
		}
	}

	return false
}

// addToGlobalSection adds a key-value pair to the global section
func (d *Document) addToGlobalSection(keys parser.Key, value parser.Value) error {
	if d.doc.Global == nil {
		d.doc.Global = &tomledit.Section{}
	}

	kv := &parser.KeyValue{
		Name:  keys,
		Value: value,
	}

	transform.InsertMapping(d.doc.Global, kv, true)
	return nil
}

// addDottedKey adds a key as a dotted key in the global section
func (d *Document) addDottedKey(keys parser.Key, value parser.Value) error {
	if d.doc.Global == nil {
		d.doc.Global = &tomledit.Section{}
	}

	kv := &parser.KeyValue{
		Name:  keys, // Full dotted path
		Value: value,
	}

	transform.InsertMapping(d.doc.Global, kv, true)
	return nil
}

// addToSection adds a key-value pair to an existing section
func (d *Document) addToSection(section *tomledit.Section, key parser.Key, value parser.Value) error {
	kv := &parser.KeyValue{
		Name:  key,
		Value: value,
	}

	transform.InsertMapping(section, kv, true)
	return nil
}

// ensureSection finds or creates a table section for the given key.
func (d *Document) ensureSection(name parser.Key) (*tomledit.Section, error) {
	if len(name) == 0 {
		if d.doc.Global == nil {
			d.doc.Global = &tomledit.Section{}
		}
		return d.doc.Global, nil
	}

	if entry := transform.FindTable(d.doc, name...); entry != nil {
		return entry.Section, nil
	}

	// Ensure parent section exists to maintain hierarchy.
	if len(name) > 1 {
		if _, err := d.ensureSection(name[:len(name)-1]); err != nil {
			return nil, err
		}
	}

	section := &tomledit.Section{
		Heading: &parser.Heading{
			Name: copyKey(name),
		},
	}

	d.doc.Sections = append(d.doc.Sections, section)

	return section, nil
}

func copyKey(key parser.Key) parser.Key {
	if key == nil {
		return nil
	}
	out := make(parser.Key, len(key))
	copy(out, key)
	return out
}

// Delete removes a key at the given dotted path.
// Returns nil if the path doesn't exist.
func (d *Document) Delete(path string) error {
	keys, err := parseKeyPath(path)
	if err != nil {
		return err
	}

	entry := d.doc.First(keys...)
	if entry == nil {
		return nil // Path doesn't exist, nothing to delete
	}

	entry.Remove()
	return nil
}

// Has returns true if the given path exists in the document.
func (d *Document) Has(path string) bool {
	val, _ := d.Get(path)
	return val != nil
}

// String serializes the document back to TOML format, preserving all
// comments, formatting, and declaration order.
func (d *Document) String() string {
	return string(d.Bytes())
}

// Bytes serializes the document back to TOML format as a byte slice.
func (d *Document) Bytes() []byte {
	var buf bytes.Buffer
	var formatter tomledit.Formatter
	formatter.Format(&buf, d.doc)
	return buf.Bytes()
}

// parseKeyPath parses a dotted path into a parser.Key using the TOML-compliant
// parser from tomledit. This properly handles quoted keys like aliases."." and
// rejects invalid paths like "aliases." (trailing dot).
//
// Examples:
//   - "simple" -> ["simple"]
//   - "dotted.key" -> ["dotted", "key"]
//   - "aliases.\".\"" or `aliases."."` -> ["aliases", "."]
//   - "aliases." -> error (trailing dot not allowed)
//   - "" -> error (empty path not allowed)
func parseKeyPath(path string) (parser.Key, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}

	// Use tomledit's parser.ParseKey which handles quoted keys and validates
	// according to TOML specification
	key, err := parser.ParseKey(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path %q: %w", path, err)
	}

	return key, nil
}

// parseValue converts a parser.Value into a Go value.
func parseValue(v parser.Value) (interface{}, error) {
	switch datum := v.X.(type) {
	case parser.Token:
		// Get the string representation and parse it
		text := datum.String()

		// Try to determine the type and parse accordingly
		// For strings, they'll be quoted, so unquote them
		if strings.HasPrefix(text, `"`) || strings.HasPrefix(text, `'`) {
			return unquoteString(text), nil
		}

		// Try boolean
		if text == "true" {
			return true, nil
		}
		if text == "false" {
			return false, nil
		}

		// Try integer
		if i, err := strconv.ParseInt(text, 10, 64); err == nil {
			return i, nil
		}

		// Try float
		if f, err := strconv.ParseFloat(text, 64); err == nil {
			return f, nil
		}

		// Return as string if all else fails
		return text, nil

	case parser.Array:
		result := make([]interface{}, 0, len(datum))
		for _, item := range datum {
			if arrayItem, ok := item.(parser.Value); ok {
				val, err := parseValue(arrayItem)
				if err != nil {
					return nil, err
				}
				result = append(result, val)
			}
		}
		return result, nil

	case parser.Inline:
		result := make(map[string]interface{})
		for _, kv := range datum {
			val, err := parseValue(kv.Value)
			if err != nil {
				return nil, err
			}
			result[kv.Name.String()] = val
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported value type: %T", v.X)
	}
}

// FormatValueToString converts a Go value into a TOML value string.
// This is useful for displaying values in a TOML-compatible format.
func FormatValueToString(v interface{}) (string, error) {
	return formatValueToString(v)
}

// formatValueToString converts a Go value into a TOML value string.
func formatValueToString(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		// Quote and escape the string
		return quoteString(val), nil

	case int:
		return strconv.FormatInt(int64(val), 10), nil

	case int64:
		return strconv.FormatInt(val, 10), nil

	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil

	case bool:
		if val {
			return "true", nil
		}
		return "false", nil

	case []interface{}:
		parts := make([]string, len(val))
		for i, item := range val {
			s, err := formatValueToString(item)
			if err != nil {
				return "", err
			}
			parts[i] = s
		}
		return "[" + strings.Join(parts, ", ") + "]", nil

	case []string:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = quoteString(item)
		}
		return "[" + strings.Join(parts, ", ") + "]", nil

	case []int:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = strconv.Itoa(item)
		}
		return "[" + strings.Join(parts, ", ") + "]", nil

	case map[string]interface{}:
		parts := make([]string, 0, len(val))
		for k, v := range val {
			vs, err := formatValueToString(v)
			if err != nil {
				return "", err
			}
			parts = append(parts, k+" = "+vs)
		}
		return "{" + strings.Join(parts, ", ") + "}", nil

	default:
		return "", fmt.Errorf("unsupported value type: %T", v)
	}
}

// quoteString adds quotes around a string value and escapes special characters.
func quoteString(s string) string {
	// Use basic string (double quotes) and escape special characters
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")
	return `"` + escaped + `"`
}

// unquoteString removes quotes and unescapes a string value.
func unquoteString(s string) string {
	// Remove surrounding quotes
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			s = s[1 : len(s)-1]
		}
	}

	// Unescape common sequences (order matters - do \\ last to avoid double unescaping)
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result = append(result, '\n')
				i++
			case 'r':
				result = append(result, '\r')
				i++
			case 't':
				result = append(result, '\t')
				i++
			case '"':
				result = append(result, '"')
				i++
			case '\'':
				result = append(result, '\'')
				i++
			case '\\':
				result = append(result, '\\')
				i++
			default:
				result = append(result, s[i])
			}
		} else {
			result = append(result, s[i])
		}
	}

	return string(result)
}
