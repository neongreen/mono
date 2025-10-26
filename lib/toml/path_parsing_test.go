package toml

import (
	"strings"
	"testing"
)

// TestParseKeyPath tests the parseKeyPath function with various valid and invalid inputs
func TestParseKeyPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		want      []string
		wantErr   bool
		errSubstr string // substring that should appear in error message
	}{
		// Valid simple paths
		{
			name: "simple key",
			path: "simple",
			want: []string{"simple"},
		},
		{
			name: "dotted key",
			path: "server.host",
			want: []string{"server", "host"},
		},
		{
			name: "deeply nested path",
			path: "app.server.tls.cert",
			want: []string{"app", "server", "tls", "cert"},
		},

		// Valid quoted keys (TOML allows special characters in quoted keys)
		{
			name: "quoted dot key with double quotes",
			path: `aliases."."`,
			want: []string{"aliases", "."},
		},
		{
			name: "quoted double dot key",
			path: `aliases.".."`,
			want: []string{"aliases", ".."},
		},
		{
			name: "quoted key with spaces",
			path: `section."key with spaces"`,
			want: []string{"section", "key with spaces"},
		},
		{
			name: "quoted key with special chars",
			path: `section."key-with-dashes"`,
			want: []string{"section", "key-with-dashes"},
		},
		{
			name: "multiple quoted keys",
			path: `"first section"."second key"`,
			want: []string{"first section", "second key"},
		},

		// Invalid paths that should be rejected
		{
			name:      "empty path",
			path:      "",
			wantErr:   true,
			errSubstr: "empty path",
		},
		{
			name:      "trailing dot",
			path:      "aliases.",
			wantErr:   true,
			errSubstr: "invalid path",
		},
		{
			name:      "leading dot",
			path:      ".aliases",
			wantErr:   true,
			errSubstr: "invalid path",
		},
		{
			name:      "double dot in middle",
			path:      "server..host",
			wantErr:   true,
			errSubstr: "invalid path",
		},
		{
			name:      "just a dot",
			path:      ".",
			wantErr:   true,
			errSubstr: "invalid path",
		},
		{
			name:      "double dots",
			path:      "..",
			wantErr:   true,
			errSubstr: "invalid path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKeyPath(tt.path)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseKeyPath(%q) expected error but got nil", tt.path)
					return
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("parseKeyPath(%q) error = %v, want error containing %q", tt.path, err, tt.errSubstr)
				}
				return
			}

			// Check success expectations
			if err != nil {
				t.Errorf("parseKeyPath(%q) unexpected error: %v", tt.path, err)
				return
			}

			// Compare results
			if len(got) != len(tt.want) {
				t.Errorf("parseKeyPath(%q) got %d parts, want %d parts", tt.path, len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseKeyPath(%q)[%d] = %q, want %q", tt.path, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestDocumentWithQuotedKeys tests Get/Set/Delete with quoted keys in actual TOML documents
func TestDocumentWithQuotedKeys(t *testing.T) {
	// Create a document with quoted keys
	input := `
[aliases]
"." = "status"
".." = "show @-"
s = "status"
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse TOML: %v", err)
	}

	// Test getting quoted keys
	t.Run("get quoted dot key", func(t *testing.T) {
		val, err := doc.Get(`aliases."."`)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "status" {
			t.Errorf("Get(`aliases.\".\"`) = %v, want %q", val, "status")
		}
	})

	t.Run("get quoted double dot key", func(t *testing.T) {
		val, err := doc.Get(`aliases.".."`)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "show @-" {
			t.Errorf(`Get("aliases.\"..\"") = %v, want %q`, val, "show @-")
		}
	})

	t.Run("get regular key in same section", func(t *testing.T) {
		val, err := doc.Get("aliases.s")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "status" {
			t.Errorf("Get(\"aliases.s\") = %v, want %q", val, "status")
		}
	})

	// Test setting quoted keys
	t.Run("set quoted key", func(t *testing.T) {
		err := doc.Set(`aliases."..."`, "show @--")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		val, err := doc.Get(`aliases."..."`)
		if err != nil {
			t.Fatalf("Get after Set failed: %v", err)
		}
		if val != "show @--" {
			t.Errorf("After Set, Get = %v, want %q", val, "show @--")
		}
	})

	// Test deleting quoted keys
	t.Run("delete quoted key", func(t *testing.T) {
		err := doc.Delete(`aliases."."`)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		val, err := doc.Get(`aliases."."`)
		if err != nil {
			t.Fatalf("Get after Delete failed: %v", err)
		}
		if val != nil {
			t.Errorf("After Delete, Get = %v, want nil", val)
		}
	})
}

// TestInvalidPathsRejected ensures invalid paths are rejected in document operations
func TestInvalidPathsRejected(t *testing.T) {
	doc, err := ParseString("[section]\nkey = \"value\"")
	if err != nil {
		t.Fatalf("Failed to parse TOML: %v", err)
	}

	invalidPaths := []string{
		"",             // empty
		"section.",     // trailing dot
		".section",     // leading dot
		"section..key", // double dot
	}

	for _, path := range invalidPaths {
		t.Run("Get rejects "+path, func(t *testing.T) {
			_, err := doc.Get(path)
			if err == nil {
				t.Errorf("Get(%q) should return error but got nil", path)
			}
		})

		t.Run("Set rejects "+path, func(t *testing.T) {
			err := doc.Set(path, "value")
			if err == nil {
				t.Errorf("Set(%q) should return error but got nil", path)
			}
		})

		t.Run("Delete rejects "+path, func(t *testing.T) {
			err := doc.Delete(path)
			if err == nil {
				t.Errorf("Delete(%q) should return error but got nil", path)
			}
		})
	}
}

// TestComplexQuotedKeyScenarios tests more complex real-world scenarios
func TestComplexQuotedKeyScenarios(t *testing.T) {
	t.Run("jj config style aliases", func(t *testing.T) {
		// This mimics how jj (Jujutsu VCS) uses quoted keys for single-character aliases
		input := `
[aliases]
"." = "status"
".." = "show @-"
"..." = "show @--"
l = "log"
`
		doc, err := ParseString(input)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Verify all aliases can be read
		aliases := map[string]string{
			`aliases."."`:   "status",
			`aliases.".."`:  "show @-",
			`aliases."..."`: "show @--",
			"aliases.l":     "log",
		}

		for path, expected := range aliases {
			val, err := doc.Get(path)
			if err != nil {
				t.Errorf("Get(%q) failed: %v", path, err)
				continue
			}
			if val != expected {
				t.Errorf("Get(%q) = %v, want %q", path, val, expected)
			}
		}
	})

	t.Run("quoted keys with special characters", func(t *testing.T) {
		doc, err := ParseString("")
		if err != nil {
			t.Fatalf("Failed to create document: %v", err)
		}

		// Set various quoted keys
		testCases := []struct {
			path  string
			value string
		}{
			{`config."key-with-dashes"`, "value1"},
			{`config."key with spaces"`, "value2"},
			{`config."key.with.dots"`, "value3"},
		}

		for _, tc := range testCases {
			err := doc.Set(tc.path, tc.value)
			if err != nil {
				t.Errorf("Set(%q) failed: %v", tc.path, err)
			}

			val, err := doc.Get(tc.path)
			if err != nil {
				t.Errorf("Get(%q) after Set failed: %v", tc.path, err)
			}
			if val != tc.value {
				t.Errorf("Get(%q) = %v, want %q", tc.path, val, tc.value)
			}
		}
	})
}
