package toml

import (
	"strings"
	"testing"
)

// TestParse tests basic parsing functionality
func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "simple key-value",
			input: `
name = "test"
version = 1
`,
			wantErr: false,
		},
		{
			name: "with comments",
			input: `
# This is a comment
name = "test"  # inline comment
# Another comment
version = 1
`,
			wantErr: false,
		},
		{
			name: "with sections",
			input: `
[server]
host = "localhost"
port = 8080

[database]
url = "postgres://localhost/db"
`,
			wantErr: false,
		},
		{
			name: "empty document",
			input: `
# Just comments
# No actual data
`,
			wantErr: false,
		},
		{
			name: "invalid toml",
			input: `
this is not valid toml at all
`,
			wantErr: true,
		},
		{
			name: "nested sections",
			input: `
[app]
name = "myapp"

[app.server]
host = "localhost"

[app.server.tls]
enabled = true
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGet tests retrieving values from the document
func TestGet(t *testing.T) {
	input := `
# Configuration file
name = "myapp"
version = 2
pi = 3.14
enabled = true
tags = ["go", "toml", "parser"]

[server]
host = "localhost"
port = 8080

[server.tls]
enabled = false
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	tests := []struct {
		name     string
		path     string
		want     any
		wantType string
	}{
		{
			name:     "get string",
			path:     "name",
			want:     "myapp",
			wantType: "string",
		},
		{
			name:     "get integer",
			path:     "version",
			want:     int64(2),
			wantType: "int64",
		},
		{
			name:     "get float",
			path:     "pi",
			want:     3.14,
			wantType: "float64",
		},
		{
			name:     "get boolean",
			path:     "enabled",
			want:     true,
			wantType: "bool",
		},
		{
			name:     "get nested value",
			path:     "server.host",
			want:     "localhost",
			wantType: "string",
		},
		{
			name:     "get deeply nested value",
			path:     "server.tls.enabled",
			want:     false,
			wantType: "bool",
		},
		{
			name:     "get non-existent key",
			path:     "nonexistent",
			want:     nil,
			wantType: "nil",
		},
		{
			name:     "get non-existent nested key",
			path:     "server.nonexistent",
			want:     nil,
			wantType: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := doc.Get(tt.path)
			if err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %v (type %T), want %v (type %s)", tt.path, got, got, tt.want, tt.wantType)
			}
		})
	}
}

// TestSet tests setting values in the document
func TestSet(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		path     string
		value    any
		expected string
	}{
		{
			name:     "set new top-level string",
			initial:  `# Config`,
			path:     "name",
			value:    "test",
			expected: "name",
		},
		{
			name:     "set new integer",
			initial:  ``,
			path:     "count",
			value:    42,
			expected: "count",
		},
		{
			name:     "set new float",
			initial:  ``,
			path:     "ratio",
			value:    3.14,
			expected: "ratio",
		},
		{
			name:     "set new boolean",
			initial:  ``,
			path:     "enabled",
			value:    true,
			expected: "enabled",
		},
		{
			name: "update existing value",
			initial: `
name = "old"
`,
			path:     "name",
			value:    "new",
			expected: "name",
		},
		{
			name: "set nested value",
			initial: `
[server]
host = "localhost"
`,
			path:     "server.port",
			value:    8080,
			expected: "port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.initial)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			err = doc.Set(tt.path, tt.value)
			if err != nil {
				t.Errorf("Set() error = %v", err)
				return
			}

			// Verify the value was set
			got, err := doc.Get(tt.path)
			if err != nil {
				t.Errorf("Get() after Set() error = %v", err)
				return
			}

			// Convert value to comparable type
			want := tt.value
			if intVal, ok := tt.value.(int); ok {
				want = int64(intVal)
			}

			if got != want {
				t.Errorf("After Set(%q, %v), Get() = %v, want %v", tt.path, tt.value, got, want)
			}

			// Verify the output contains the expected key
			output := doc.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Output doesn't contain %q:\n%s", tt.expected, output)
			}
		})
	}
}

// TestDelete tests deleting values from the document
func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		path    string
	}{
		{
			name: "delete top-level key",
			initial: `
name = "test"
version = 1
`,
			path: "name",
		},
		{
			name: "delete nested key",
			initial: `
[server]
host = "localhost"
port = 8080
`,
			path: "server.port",
		},
		{
			name: "delete non-existent key",
			initial: `
name = "test"
`,
			path: "nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.initial)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			err = doc.Delete(tt.path)
			if err != nil {
				t.Errorf("Delete() error = %v", err)
				return
			}

			// Verify the value was deleted
			got, err := doc.Get(tt.path)
			if err != nil {
				t.Errorf("Get() after Delete() error = %v", err)
				return
			}
			if got != nil {
				t.Errorf("After Delete(%q), Get() = %v, want nil", tt.path, got)
			}
		})
	}
}

// TestHas tests checking if a path exists
func TestHas(t *testing.T) {
	input := `
name = "test"
[server]
host = "localhost"
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"existing top-level", "name", true},
		{"existing nested", "server.host", true},
		{"non-existent top-level", "version", false},
		{"non-existent nested", "server.port", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := doc.Has(tt.path)
			if got != tt.want {
				t.Errorf("Has(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestCommentPreservation tests that comments are preserved during round-trip
func TestCommentPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		path           string
		value          any
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "preserve top-level comments",
			input: `
# This is a header comment
name = "test"
# This is a footer comment
version = 1
`,
			path:  "name",
			value: "modified",
			mustContain: []string{
				"# This is a header comment",
				"# This is a footer comment",
			},
		},
		{
			name: "preserve section comments",
			input: `
# Server configuration
[server]
# The host to bind to
host = "localhost"
# The port to listen on
port = 8080
`,
			path:  "server.port",
			value: 9090,
			mustContain: []string{
				"# Server configuration",
				"# The host to bind to",
				"# The port to listen on",
			},
		},
		{
			name: "preserve inline comments",
			input: `
name = "test"  # application name
version = 1    # version number
`,
			path:  "version",
			value: 2,
			mustContain: []string{
				"# application name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			err = doc.Set(tt.path, tt.value)
			if err != nil {
				t.Errorf("Set() error = %v", err)
				return
			}

			output := doc.String()

			for _, mustHave := range tt.mustContain {
				if !strings.Contains(output, mustHave) {
					t.Errorf("Output missing expected comment %q:\n%s", mustHave, output)
				}
			}

			for _, mustNotHave := range tt.mustNotContain {
				if strings.Contains(output, mustNotHave) {
					t.Errorf("Output contains unexpected text %q:\n%s", mustNotHave, output)
				}
			}
		})
	}
}

// TestRoundTrip tests that documents can be parsed and serialized without loss
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple document",
			input: `# Configuration
name = "myapp"
version = 1
`,
		},
		{
			name: "document with sections",
			input: `# Main config
app = "test"

# Server settings
[server]
host = "localhost"
port = 8080

# Database settings
[database]
url = "postgres://localhost/db"
`,
		},
		{
			name: "complex document with nested sections",
			input: `# Application configuration
name = "myapp"
version = 2

# Server configuration
[server]
# Network settings
host = "0.0.0.0"
port = 8080

# TLS configuration
[server.tls]
enabled = true
cert = "/path/to/cert"

# Database configuration
[database]
url = "postgres://localhost/db"
pool_size = 10
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			output := doc.String()

			// Parse the output again to ensure it's valid TOML
			_, err = ParseString(output)
			if err != nil {
				t.Errorf("Round-trip produced invalid TOML: %v\nOutput:\n%s", err, output)
			}
		})
	}
}

// TestArrays tests handling of array values
func TestArrays(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		value any
	}{
		{
			name:  "string array",
			path:  "tags",
			value: []any{"go", "toml", "parser"},
		},
		{
			name:  "integer array",
			path:  "numbers",
			value: []any{1, 2, 3, 4, 5},
		},
		{
			name:  "empty array",
			path:  "empty",
			value: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString("")
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			err = doc.Set(tt.path, tt.value)
			if err != nil {
				t.Errorf("Set() error = %v", err)
				return
			}

			got, err := doc.Get(tt.path)
			if err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}

			gotArr, ok := got.([]any)
			if !ok {
				t.Errorf("Get() returned %T, want []interface{}", got)
				return
			}

			wantArr := tt.value.([]any)
			if len(gotArr) != len(wantArr) {
				t.Errorf("Array length = %d, want %d", len(gotArr), len(wantArr))
				return
			}

			for i := range gotArr {
				// Convert ints to int64 for comparison
				gotVal := gotArr[i]
				wantVal := wantArr[i]
				if intVal, ok := wantVal.(int); ok {
					wantVal = int64(intVal)
				}
				if gotVal != wantVal {
					t.Errorf("Array[%d] = %v, want %v", i, gotVal, wantVal)
				}
			}
		})
	}
}

// TestInlineTables tests handling of inline table values
func TestInlineTables(t *testing.T) {
	input := `
person = { name = "John", age = 30 }
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	got, err := doc.Get("person")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	table, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("Get() returned %T, want map[string]interface{}", got)
	}

	if table["name"] != "John" {
		t.Errorf("table[name] = %v, want John", table["name"])
	}

	if table["age"] != int64(30) {
		t.Errorf("table[age] = %v, want 30", table["age"])
	}
}

// TestSetInlineTable tests setting inline table values
func TestSetInlineTable(t *testing.T) {
	doc, err := ParseString("")
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	value := map[string]any{
		"name": "Alice",
		"age":  25,
	}

	err = doc.Set("person", value)
	if err != nil {
		t.Errorf("Set() error = %v", err)
		return
	}

	got, err := doc.Get("person")
	if err != nil {
		t.Errorf("Get() error = %v", err)
		return
	}

	table, ok := got.(map[string]any)
	if !ok {
		t.Errorf("Get() returned %T, want map[string]interface{}", got)
		return
	}

	if table["name"] != "Alice" {
		t.Errorf("table[name] = %v, want Alice", table["name"])
	}

	// Note: integers are returned as int64
	if table["age"] != int64(25) {
		t.Errorf("table[age] = %v, want 25", table["age"])
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		doc, _ := ParseString("name = 'test'")
		_, err := doc.Get("")
		if err == nil {
			t.Error("Get('') should return error")
		}
	})

	t.Run("empty document", func(t *testing.T) {
		doc, err := ParseString("")
		if err != nil {
			t.Errorf("ParseString('') error = %v", err)
		}
		output := doc.String()
		if output != "" {
			t.Errorf("Empty document String() = %q, want empty string", output)
		}
	})

	t.Run("only comments", func(t *testing.T) {
		input := `# Just a comment`
		doc, err := ParseString(input)
		if err != nil {
			t.Errorf("ParseString() error = %v", err)
		}
		output := doc.String()
		if !strings.Contains(output, "# Just a comment") {
			t.Errorf("Comment not preserved in output: %s", output)
		}
	})

	t.Run("special characters in strings", func(t *testing.T) {
		doc, _ := ParseString("")
		specialChars := `Hello "World"\nNew line\tTab`
		err := doc.Set("text", specialChars)
		if err != nil {
			t.Errorf("Set() error = %v", err)
		}

		got, err := doc.Get("text")
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}

		if got != specialChars {
			t.Errorf("Get() = %q, want %q", got, specialChars)
		}
	})
}

// TestBytes tests the Bytes method
func TestBytes(t *testing.T) {
	input := `name = "test"`
	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	bytes := doc.Bytes()
	if len(bytes) == 0 {
		t.Error("Bytes() returned empty slice")
	}

	str := doc.String()
	if string(bytes) != str {
		t.Errorf("Bytes() and String() don't match")
	}
}
