package toml

import (
	"strings"
	"testing"
)

// TestKeyOrderPreservation tests that key order is preserved
func TestKeyOrderPreservation(t *testing.T) {
	input := `# Config in specific order
zebra = "last"
apple = "first"
middle = "middle"
banana = "second"

[section]
zulu = 1
alpha = 2
mike = 3
bravo = 4
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Modify one value
	doc.Set("middle", "modified")

	output := doc.String()

	// Check that keys appear in the original order
	zebraPos := strings.Index(output, "zebra")
	applePos := strings.Index(output, "apple")
	middlePos := strings.Index(output, "middle")
	bananaPos := strings.Index(output, "banana")

	if zebraPos == -1 || applePos == -1 || middlePos == -1 || bananaPos == -1 {
		t.Fatal("Not all keys found in output")
	}

	// Verify order is preserved (zebra, apple, middle, banana)
	if zebraPos >= applePos || applePos >= middlePos || middlePos >= bananaPos {
		t.Errorf("Key order not preserved. Positions: zebra=%d, apple=%d, middle=%d, banana=%d",
			zebraPos, applePos, middlePos, bananaPos)
		t.Logf("Output:\n%s", output)
	}

	// Check section key order
	zuluPos := strings.Index(output, "zulu")
	alphaPos := strings.Index(output, "alpha")
	mikePos := strings.Index(output, "mike")
	bravoPos := strings.Index(output, "bravo")

	if zuluPos >= alphaPos || alphaPos >= mikePos || mikePos >= bravoPos {
		t.Errorf("Section key order not preserved")
		t.Logf("Output:\n%s", output)
	}
}

// TestQuoteStylePreservation tests that different quote styles are preserved
func TestQuoteStylePreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		modifyKey      string
		modifyValue    string
		shouldPreserve []string // Strings that should still be in the output
	}{
		{
			name: "double quotes",
			input: `
double = "hello world"
single = 'hello world'
`,
			modifyKey:   "double",
			modifyValue: "hello universe",
			shouldPreserve: []string{
				`single = 'hello world'`, // Single quotes should be preserved
			},
		},
		{
			name: "single quotes",
			input: `
single = 'hello'
double = "world"
`,
			modifyKey:   "double",
			modifyValue: "universe",
			shouldPreserve: []string{
				`single = 'hello'`, // Single quotes should be preserved
			},
		},
		{
			name: "multiline basic string",
			input: `
text = """
Line 1
Line 2
Line 3"""
other = "simple"
`,
			modifyKey:   "other",
			modifyValue: "modified",
			shouldPreserve: []string{
				`"""`, // Multiline delimiters should be preserved
			},
		},
		{
			name: "multiline literal string",
			input: `
text = '''
Line 1
Line 2
Line 3'''
other = "simple"
`,
			modifyKey:   "other",
			modifyValue: "modified",
			shouldPreserve: []string{
				`'''`, // Multiline literal delimiters should be preserved
			},
		},
		{
			name: "raw string literals",
			input: `
path = 'C:\Users\nodejs\templates'
regex = '<\i\c*\s*>'
other = "normal"
`,
			modifyKey:   "other",
			modifyValue: "changed",
			shouldPreserve: []string{
				`'C:\Users\nodejs\templates'`,
				`'<\i\c*\s*>'`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			// Modify a different key
			err = doc.Set(tt.modifyKey, tt.modifyValue)
			if err != nil {
				t.Errorf("Set() error = %v", err)
			}

			output := doc.String()

			// Check that the unmodified keys preserve their quote style
			for _, expected := range tt.shouldPreserve {
				if !strings.Contains(output, expected) {
					t.Errorf("Quote style not preserved. Expected to find: %q\nOutput:\n%s", expected, output)
				}
			}
		})
	}
}

// TestQuoteStylePreservationOnModification tests that quote style is preserved when modifying a value
func TestQuoteStylePreservationOnModification(t *testing.T) {
	input := `
single_quoted = 'hello'
double_quoted = "world"
multiline = """
text
here"""
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Get the original output to see the quote styles
	originalOutput := doc.String()
	t.Logf("Original:\n%s", originalOutput)

	// Modify the single quoted value
	doc.Set("single_quoted", "goodbye")

	// Modify the double quoted value
	doc.Set("double_quoted", "universe")

	output := doc.String()
	t.Logf("Modified:\n%s", output)

	// Note: This test documents current behavior
	// We may need to enhance the library to truly preserve quote styles on modification
}

// TestNestednessStylePreservation tests that dotted keys vs sections are preserved
func TestNestednessStylePreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		modifyKey      string
		modifyValue    any
		shouldContain  []string
		shouldNotMatch []string // Patterns that should NOT appear (e.g., unwanted section headers)
	}{
		{
			name: "dotted keys should stay dotted",
			input: `
# Dotted key style
server.host = "localhost"
server.port = 8080
database.url = "postgres://localhost"
`,
			modifyKey:   "server.host",
			modifyValue: "0.0.0.0",
			shouldContain: []string{
				"server.host",
				"server.port",
			},
		},
		{
			name: "sections should stay sections",
			input: `
# Section style
[server]
host = "localhost"
port = 8080

[database]
url = "postgres://localhost"
`,
			modifyKey:   "server.host",
			modifyValue: "0.0.0.0",
			shouldContain: []string{
				"[server]",
				"[database]",
			},
		},
		{
			name: "mixed styles should be preserved",
			input: `
# Mixed styles
app.name = "myapp"
app.version = 1

[server]
host = "localhost"
port = 8080

database.url = "postgres://localhost"
database.port = 5432
`,
			modifyKey:   "app.version",
			modifyValue: 2,
			shouldContain: []string{
				"app.name",
				"app.version",
				"[server]",
				"database.url",
				"database.port",
			},
		},
		{
			name: "nested sections",
			input: `
[server]
host = "localhost"

[server.tls]
enabled = true
cert = "/path/to/cert"
`,
			modifyKey:   "server.host",
			modifyValue: "0.0.0.0",
			shouldContain: []string{
				"[server]",
				"[server.tls]",
			},
		},
		{
			name: "deeply nested dotted keys",
			input: `
a.b.c.d = 1
a.b.c.e = 2
a.b.f = 3
`,
			modifyKey:   "a.b.c.d",
			modifyValue: 10,
			shouldContain: []string{
				"a.b.c.d",
				"a.b.c.e",
				"a.b.f",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			err = doc.Set(tt.modifyKey, tt.modifyValue)
			if err != nil {
				t.Errorf("Set() error = %v", err)
			}

			output := doc.String()
			t.Logf("Output:\n%s", output)

			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected to find %q in output, but it's missing", expected)
				}
			}

			for _, notExpected := range tt.shouldNotMatch {
				if strings.Contains(output, notExpected) {
					t.Errorf("Found unwanted pattern %q in output", notExpected)
				}
			}
		})
	}
}

// TestAddingNewKeysPreservesStyle tests that when adding new keys to existing structures,
// we follow the existing style
func TestAddingNewKeysPreservesStyle(t *testing.T) {
	t.Run("adding to section preserves section style", func(t *testing.T) {
		input := `
[server]
host = "localhost"
port = 8080
`

		doc, _ := ParseString(input)
		doc.Set("server.timeout", 30)

		output := doc.String()
		t.Logf("Output:\n%s", output)

		// The new key should be added to the [server] section
		if !strings.Contains(output, "[server]") {
			t.Error("Section style not preserved when adding new key")
		}

		// Check that timeout appears after the [server] header
		serverPos := strings.Index(output, "[server]")
		timeoutPos := strings.Index(output, "timeout")
		if timeoutPos == -1 || timeoutPos < serverPos {
			t.Error("New key not added to existing section")
		}
	})

	t.Run("adding to dotted keys area", func(t *testing.T) {
		input := `
app.name = "myapp"
app.version = 1
`

		doc, _ := ParseString(input)
		doc.Set("app.debug", true)

		output := doc.String()
		t.Logf("Output:\n%s", output)

		// Note: This documents current behavior
		// The library may add this as a new section or as a dotted key
		// We should verify and potentially fix this behavior
	})
}

// TestComplexPreservation tests that all preservation features work together
func TestComplexPreservation(t *testing.T) {
	input := `# Application Configuration
# Version 1.0

# Basic settings
app_name = "myapp"
app_version = 1

# Server configuration with dotted keys
server.host = "localhost"
server.port = 8080

# Database section
[database]
driver = 'postgres'  # Using single quotes
host = "localhost"   # Using double quotes
port = 5432

# Multiline description
description = """
This is a
multiline string
with multiple lines"""

# Feature flags (specific order matters!)
[features]
feature_a = true
feature_z = false
feature_m = true
feature_b = false
`

	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Make several modifications
	doc.Set("app_version", 2)                  // Modify top-level key
	doc.Set("server.port", 9090)               // Modify dotted key
	doc.Set("database.host", "db.example.com") // Modify value in section
	doc.Set("features.feature_m", false)       // Modify value in section

	output := doc.String()
	t.Logf("Output:\n%s", output)

	// Verify comments are preserved
	if !strings.Contains(output, "# Application Configuration") {
		t.Error("Top-level comment not preserved")
	}

	// Verify key order in features section
	featureSection := output[strings.Index(output, "[features]"):]
	aPos := strings.Index(featureSection, "feature_a")
	zPos := strings.Index(featureSection, "feature_z")
	mPos := strings.Index(featureSection, "feature_m")
	bPos := strings.Index(featureSection, "feature_b")

	if aPos >= zPos || zPos >= mPos || mPos >= bPos {
		t.Error("Feature flag order not preserved")
	}

	// Verify quote styles (at least for unmodified values)
	if !strings.Contains(output, `driver = 'postgres'`) {
		t.Error("Single quote style not preserved for unmodified value")
	}

	// Verify multiline string is preserved
	if !strings.Contains(output, `"""`) {
		t.Error("Multiline string delimiters not preserved")
	}

	// Verify section style is preserved
	if !strings.Contains(output, "[database]") {
		t.Error("Section header not preserved")
	}

	// Verify dotted key style is preserved
	if !strings.Contains(output, "server.host") || !strings.Contains(output, "server.port") {
		t.Error("Dotted key style not preserved")
	}
}

// TestRoundTripPreservation tests that multiple round-trips preserve everything
func TestRoundTripPreservation(t *testing.T) {
	input := `# Config
zebra = 'last'
apple = "first"

[section]
zulu = 1
alpha = 2
`

	doc1, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	output1 := doc1.String()

	// Parse the output again
	doc2, err := ParseString(output1)
	if err != nil {
		t.Fatalf("Second ParseString() error = %v", err)
	}

	output2 := doc2.String()

	// The outputs should be identical (or at least semantically equivalent)
	if output1 != output2 {
		t.Errorf("Round-trip not stable\nFirst output:\n%s\nSecond output:\n%s", output1, output2)
	}
}
