package toml

import (
	"strings"
	"testing"
)

// TestComplexNestedStructures tests deeply nested tables and dotted keys
func TestComplexNestedStructures(t *testing.T) {
	input := `
[database]
server = "192.168.1.1"
ports = [8001, 8001, 8002]

[database.connection]
max_retries = 5
timeout = 30

[servers.alpha]
ip = "10.0.0.1"
dc = "eqdc10"

[servers.beta]
ip = "10.0.0.2"
dc = "eqdc10"
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Test getting deeply nested values
	val, err := doc.Get("database.connection.max_retries")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if val != int64(5) {
		t.Errorf("Get(database.connection.max_retries) = %v, want 5", val)
	}

	// Test setting a new deeply nested value
	err = doc.Set("database.connection.pool_size", 10)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	val, err = doc.Get("database.connection.pool_size")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if val != int64(10) {
		t.Errorf("Get(database.connection.pool_size) = %v, want 10", val)
	}

	// Verify the output is still valid TOML
	output := doc.String()
	_, err = ParseString(output)
	if err != nil {
		t.Errorf("Round-trip produced invalid TOML: %v\nOutput:\n%s", err, output)
	}
}

// TestArrayOfTables tests array of tables syntax
func TestArrayOfTables(t *testing.T) {
	input := `
[[products]]
name = "Hammer"
sku = 738594937

[[products]]
name = "Nail"
sku = 284758393
color = "gray"
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Verify round-trip
	output := doc.String()
	_, err = ParseString(output)
	if err != nil {
		t.Errorf("Round-trip produced invalid TOML: %v\nOutput:\n%s", err, output)
	}
}

// TestMultilineStrings tests handling of multiline strings
func TestMultilineStrings(t *testing.T) {
	input := `
key1 = """
Roses are red
Violets are blue"""
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	val, err := doc.Get("key1")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	str, ok := val.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", val)
	}

	if !strings.Contains(str, "Roses") || !strings.Contains(str, "Violets") {
		t.Errorf("Multiline string not preserved correctly: %q", str)
	}
}

// TestDateTimeValues tests handling of date-time values
func TestDateTimeValues(t *testing.T) {
	input := `
odt1 = 1979-05-27T07:32:00Z
odt2 = 1979-05-27T00:32:00-07:00
ld1 = 1979-05-27
lt1 = 07:32:00
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// We don't parse datetime values, but we should preserve them
	output := doc.String()
	if !strings.Contains(output, "1979-05-27T07:32:00Z") {
		t.Errorf("DateTime value not preserved in output:\n%s", output)
	}
}

// TestNumberFormats tests different number formats
func TestNumberFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		key   string
	}{
		{
			name:  "hex integer",
			input: `hex = 0xDEADBEEF`,
			key:   "hex",
		},
		{
			name:  "octal integer",
			input: `oct = 0o755`,
			key:   "oct",
		},
		{
			name:  "binary integer",
			input: `bin = 0b11010110`,
			key:   "bin",
		},
		{
			name:  "float with exponent",
			input: `flt = 5e+22`,
			key:   "flt",
		},
		{
			name:  "integer with underscores",
			input: `num = 1_000_000`,
			key:   "num",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			val, err := doc.Get(tt.key)
			if err != nil {
				t.Errorf("Get() error = %v", err)
			}
			if val == nil {
				t.Errorf("Get(%q) returned nil", tt.key)
			}

			// Verify round-trip
			output := doc.String()
			_, err = ParseString(output)
			if err != nil {
				t.Errorf("Round-trip produced invalid TOML: %v\nOutput:\n%s", err, output)
			}
		})
	}
}

// TestLargeDocument tests handling of a large, complex document
func TestLargeDocument(t *testing.T) {
	input := `
# This is a comment
title = "TOML Example"

[owner]
name = "Tom Preston-Werner"
dob = 1979-05-27T07:32:00-08:00

[database]
server = "192.168.1.1"
ports = [8001, 8001, 8002]
connection_max = 5000
enabled = true

[servers]

  # Indented comment
  [servers.alpha]
  ip = "10.0.0.1"
  dc = "eqdc10"

  [servers.beta]
  ip = "10.0.0.2"
  dc = "eqdc10"

[clients]
data = [ ["gamma", "delta"], [1, 2] ]

# Line breaks are OK when inside arrays
hosts = [
  "alpha",
  "omega"
]
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Test various operations
	tests := []struct {
		path  string
		value any
	}{
		{"title", "TOML Example Updated"},
		{"database.connection_max", 10000},
		{"servers.alpha.role", "primary"},
		{"new_section.new_key", "new_value"},
	}

	for _, tt := range tests {
		err := doc.Set(tt.path, tt.value)
		if err != nil {
			t.Errorf("Set(%q) error = %v", tt.path, err)
			continue
		}

		got, err := doc.Get(tt.path)
		if err != nil {
			t.Errorf("Get(%q) error = %v", tt.path, err)
			continue
		}

		// Convert ints to int64 for comparison
		want := tt.value
		if intVal, ok := tt.value.(int); ok {
			want = int64(intVal)
		}

		if got != want {
			t.Errorf("After Set(%q, %v), Get() = %v", tt.path, tt.value, got)
		}
	}

	// Verify comments are preserved
	output := doc.String()
	if !strings.Contains(output, "# This is a comment") {
		t.Errorf("Comments not preserved in output")
	}

	// Verify round-trip
	_, err = ParseString(output)
	if err != nil {
		t.Errorf("Round-trip produced invalid TOML: %v", err)
	}
}

// TestConcurrentAccess tests that the library is safe for concurrent reads
func TestConcurrentAccess(t *testing.T) {
	input := `
[server]
host = "localhost"
port = 8080

[database]
url = "postgres://localhost/db"
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Test concurrent reads (writes would need synchronization from the caller)
	done := make(chan bool)
	for range 10 {
		go func() {
			val, _ := doc.Get("server.host")
			if val != "localhost" {
				t.Errorf("Concurrent Get() returned unexpected value: %v", val)
			}
			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}

// TestEmptyValues tests handling of empty values
func TestEmptyValues(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{"empty string", ""},
		{"empty array", []any{}},
		{"empty map", map[string]any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := ParseString("")
			err := doc.Set("key", tt.value)
			if err != nil {
				t.Errorf("Set() error = %v", err)
				return
			}

			got, err := doc.Get("key")
			if err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}

			// Check types match
			switch tt.value.(type) {
			case string:
				if _, ok := got.(string); !ok {
					t.Errorf("Expected string, got %T", got)
				}
			case []any:
				if _, ok := got.([]any); !ok {
					t.Errorf("Expected []interface{}, got %T", got)
				}
			case map[string]any:
				if _, ok := got.(map[string]any); !ok {
					t.Errorf("Expected map[string]interface{}, got %T", got)
				}
			}
		})
	}
}

// TestUpdateMultipleValues tests updating multiple values and verifying all changes
func TestUpdateMultipleValues(t *testing.T) {
	input := `
[app]
name = "myapp"
version = 1
debug = false

[app.server]
host = "localhost"
port = 8080
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Update multiple values
	updates := map[string]any{
		"app.name":        "newapp",
		"app.version":     2,
		"app.debug":       true,
		"app.server.port": 9090,
	}

	for path, value := range updates {
		if err := doc.Set(path, value); err != nil {
			t.Errorf("Set(%q) error = %v", path, err)
		}
	}

	// Verify all updates
	for path, expected := range updates {
		got, err := doc.Get(path)
		if err != nil {
			t.Errorf("Get(%q) error = %v", path, err)
			continue
		}

		// Convert ints to int64 for comparison
		if intVal, ok := expected.(int); ok {
			expected = int64(intVal)
		}

		if got != expected {
			t.Errorf("Get(%q) = %v, want %v", path, got, expected)
		}
	}

	// Verify the output is valid
	output := doc.String()
	_, err = ParseString(output)
	if err != nil {
		t.Errorf("Round-trip produced invalid TOML: %v\nOutput:\n%s", err, output)
	}
}

// TestDeleteMultipleKeys tests deleting multiple keys
func TestDeleteMultipleKeys(t *testing.T) {
	input := `
name = "test"
version = 1
enabled = true
count = 42

[server]
host = "localhost"
port = 8080
debug = true
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Delete multiple keys
	keysToDelete := []string{
		"version",
		"count",
		"server.debug",
	}

	for _, key := range keysToDelete {
		if err := doc.Delete(key); err != nil {
			t.Errorf("Delete(%q) error = %v", key, err)
		}
	}

	// Verify deletions
	for _, key := range keysToDelete {
		if doc.Has(key) {
			t.Errorf("After Delete(%q), Has() returned true", key)
		}
	}

	// Verify remaining keys still exist
	if !doc.Has("name") {
		t.Error("name key was incorrectly deleted")
	}
	if !doc.Has("enabled") {
		t.Error("enabled key was incorrectly deleted")
	}
	if !doc.Has("server.host") {
		t.Error("server.host key was incorrectly deleted")
	}
	if !doc.Has("server.port") {
		t.Error("server.port key was incorrectly deleted")
	}
}

// TestUnicodeSupport tests handling of Unicode characters
func TestUnicodeSupport(t *testing.T) {
	doc, _ := ParseString("")

	unicodeStrings := []string{
		"Hello 世界",
		"Привет мир",
		"مرحبا بالعالم",
		"🚀 Rocket",
		"Emoji: 😀😃😄",
	}

	for i, str := range unicodeStrings {
		key := "unicode" + string(rune('0'+i))
		err := doc.Set(key, str)
		if err != nil {
			t.Errorf("Set(%q, %q) error = %v", key, str, err)
			continue
		}

		got, err := doc.Get(key)
		if err != nil {
			t.Errorf("Get(%q) error = %v", key, err)
			continue
		}

		if got != str {
			t.Errorf("Get(%q) = %q, want %q", key, got, str)
		}
	}

	// Verify round-trip preserves Unicode
	output := doc.String()
	doc2, err := ParseString(output)
	if err != nil {
		t.Errorf("Round-trip with Unicode failed: %v\nOutput:\n%s", err, output)
		return
	}

	for i, str := range unicodeStrings {
		key := "unicode" + string(rune('0'+i))
		got, _ := doc2.Get(key)
		if got != str {
			t.Errorf("After round-trip, Get(%q) = %q, want %q", key, got, str)
		}
	}
}
