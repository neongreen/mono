// Package tomlcp provides comment-preserving TOML parsing and serialization.
//
// This package wraps github.com/creachadair/tomledit to provide a stable,
// user-friendly API for reading, modifying, and writing TOML files while
// preserving all comments, formatting, and declaration order.
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
//	// Write back with comments preserved
//	output := doc.String()
package tomlcp

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
func (d *Document) Set(path string, value interface{}) error {
	keys, err := parseKeyPath(path)
	if err != nil {
		return err
	}

	// Format the value to a TOML string
	valueStr, err := formatValueToString(value)
	if err != nil {
		return fmt.Errorf("failed to format value: %w", err)
	}

	// Parse the value string into a parser.Value
	parsedValue, err := parser.ParseValue(valueStr)
	if err != nil {
		return fmt.Errorf("failed to parse value: %w", err)
	}

	// Create the key-value pair
	kv := &parser.KeyValue{
		Name:  keys,
		Value: parsedValue,
	}

	// Check if the key already exists to preserve its comments
	existingEntry := d.doc.First(keys...)
	if existingEntry != nil && existingEntry.KeyValue != nil {
		// Preserve block comments from the existing entry
		kv.Block = existingEntry.KeyValue.Block
		// Preserve line comments
		kv.Value.Trailer = existingEntry.KeyValue.Value.Trailer
	}

	// Determine which section to add to
	var section *tomledit.Section
	if len(keys) == 1 {
		// Top-level key, add to global section
		section = d.doc.Global
		if section == nil {
			section = &tomledit.Section{}
			d.doc.Global = section
		}
	} else {
		// Nested key, find or create the appropriate table
		tableName := keys[:len(keys)-1]
		entry := transform.FindTable(d.doc, tableName...)

		if entry == nil {
			// Create the table
			section = &tomledit.Section{
				Heading: &parser.Heading{
					Name: tableName,
				},
			}
			d.doc.Sections = append(d.doc.Sections, section)
		} else {
			section = entry.Section
		}

		// Update the key-value to just use the last component
		kv.Name = parser.Key{keys[len(keys)-1]}
	}

	// Insert or replace the mapping
	transform.InsertMapping(section, kv, true)

	return nil
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

// parseKeyPath parses a dotted path into a parser.Key
func parseKeyPath(path string) (parser.Key, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}
	return strings.Split(path, "."), nil
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
