package toml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/creachadair/tomledit/parser"
)

// TestWriteFile tests writing values to TOML files
func TestWriteFile(t *testing.T) {
	t.Run("creates new file with values", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.toml")

		values := map[string]any{
			"name":    "test",
			"version": "1.0.0",
			"enabled": true,
		}

		err := WriteFile(filePath, values)
		if err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatalf("file was not created at %s", filePath)
		}

		// Verify content
		doc, err := Parse(mustReadFile(t, filePath))
		if err != nil {
			t.Fatalf("failed to parse written file: %v", err)
		}

		if got, _ := doc.Get("name"); got != "test" {
			t.Errorf("name = %v, want %q", got, "test")
		}
		if got, _ := doc.Get("version"); got != "1.0.0" {
			t.Errorf("version = %v, want %q", got, "1.0.0")
		}
		if got, _ := doc.Get("enabled"); got != true {
			t.Errorf("enabled = %v, want true", got)
		}
	})

	t.Run("updates existing file preserving other keys", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.toml")

		// Create initial file
		initial := `name = "old"
version = "0.9.0"
other = "preserved"`
		if err := os.WriteFile(filePath, []byte(initial), 0o644); err != nil {
			t.Fatalf("failed to create initial file: %v", err)
		}

		// Update with new values
		values := map[string]any{
			"name":    "new",
			"version": "1.0.0",
		}

		err := WriteFile(filePath, values)
		if err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		// Verify content
		doc, err := Parse(mustReadFile(t, filePath))
		if err != nil {
			t.Fatalf("failed to parse written file: %v", err)
		}

		if got, _ := doc.Get("name"); got != "new" {
			t.Errorf("name = %v, want %q", got, "new")
		}
		if got, _ := doc.Get("version"); got != "1.0.0" {
			t.Errorf("version = %v, want %q", got, "1.0.0")
		}
		if got, _ := doc.Get("other"); got != "preserved" {
			t.Errorf("other = %v, want %q (should be preserved)", got, "preserved")
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "subdir1", "subdir2", "test.toml")

		values := map[string]any{
			"test": "value",
		}

		err := WriteFile(filePath, values)
		if err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatalf("file was not created at %s", filePath)
		}
	})

	t.Run("handles nested values", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.toml")

		values := map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		}

		err := WriteFile(filePath, values)
		if err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		doc, err := Parse(mustReadFile(t, filePath))
		if err != nil {
			t.Fatalf("failed to parse written file: %v", err)
		}

		if got, _ := doc.Get("database.host"); got != "localhost" {
			t.Errorf("database.host = %v, want %q", got, "localhost")
		}
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		// Try to write to a directory that can't be created (e.g., inside a file)
		tempDir := t.TempDir()
		existingFile := filepath.Join(tempDir, "file.txt")
		if err := os.WriteFile(existingFile, []byte("content"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		invalidPath := filepath.Join(existingFile, "subdir", "test.toml")
		values := map[string]any{"test": "value"}

		err := WriteFile(invalidPath, values)
		if err == nil {
			t.Error("expected error for invalid path, got nil")
		}
	})
}

// TestApplyMap tests merging values into documents
func TestApplyMap(t *testing.T) {
	t.Run("adds new keys to empty document", func(t *testing.T) {
		doc, err := ParseString("")
		if err != nil {
			t.Fatalf("ParseString() error = %v", err)
		}

		values := map[string]any{
			"name":    "test",
			"version": "1.0.0",
		}

		err = doc.ApplyMap(values)
		if err != nil {
			t.Fatalf("ApplyMap() error = %v", err)
		}

		if got, _ := doc.Get("name"); got != "test" {
			t.Errorf("name = %v, want %q", got, "test")
		}
		if got, _ := doc.Get("version"); got != "1.0.0" {
			t.Errorf("version = %v, want %q", got, "1.0.0")
		}
	})

	t.Run("updates existing keys", func(t *testing.T) {
		doc, err := ParseString(`name = "old"
version = "0.9.0"`)
		if err != nil {
			t.Fatalf("ParseString() error = %v", err)
		}

		values := map[string]any{
			"name": "new",
		}

		err = doc.ApplyMap(values)
		if err != nil {
			t.Fatalf("ApplyMap() error = %v", err)
		}

		if got, _ := doc.Get("name"); got != "new" {
			t.Errorf("name = %v, want %q", got, "new")
		}
		if got, _ := doc.Get("version"); got != "0.9.0" {
			t.Errorf("version = %v, want %q (should be preserved)", got, "0.9.0")
		}
	})

	t.Run("preserves unmodified keys", func(t *testing.T) {
		doc, err := ParseString(`name = "test"
version = "1.0.0"
other = "value"`)
		if err != nil {
			t.Fatalf("ParseString() error = %v", err)
		}

		values := map[string]any{
			"name": "updated",
		}

		err = doc.ApplyMap(values)
		if err != nil {
			t.Fatalf("ApplyMap() error = %v", err)
		}

		if got, _ := doc.Get("name"); got != "updated" {
			t.Errorf("name = %v, want %q", got, "updated")
		}
		if got, _ := doc.Get("version"); got != "1.0.0" {
			t.Errorf("version = %v, want %q (should be preserved)", got, "1.0.0")
		}
		if got, _ := doc.Get("other"); got != "value" {
			t.Errorf("other = %v, want %q (should be preserved)", got, "value")
		}
	})

	t.Run("handles nested values", func(t *testing.T) {
		doc, err := ParseString("")
		if err != nil {
			t.Fatalf("ParseString() error = %v", err)
		}

		values := map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		}

		err = doc.ApplyMap(values)
		if err != nil {
			t.Fatalf("ApplyMap() error = %v", err)
		}

		if got, _ := doc.Get("database.host"); got != "localhost" {
			t.Errorf("database.host = %v, want %q", got, "localhost")
		}
		if got, _ := doc.Get("database.port"); got != int64(5432) {
			t.Errorf("database.port = %v, want 5432", got)
		}
	})

	t.Run("returns error for nil document", func(t *testing.T) {
		var doc *Document
		values := map[string]any{"test": "value"}

		err := doc.ApplyMap(values)
		if err == nil {
			t.Error("expected error for nil document, got nil")
		}
		if !strings.Contains(err.Error(), "nil document") {
			t.Errorf("error should mention nil document, got: %v", err)
		}
	})

	t.Run("handles arrays", func(t *testing.T) {
		doc, err := ParseString("")
		if err != nil {
			t.Fatalf("ParseString() error = %v", err)
		}

		values := map[string]any{
			"items": []any{"a", "b", "c"},
		}

		err = doc.ApplyMap(values)
		if err != nil {
			t.Fatalf("ApplyMap() error = %v", err)
		}

		got, _ := doc.Get("items")
		if got == nil {
			t.Fatal("items should not be nil")
		}
	})
}

// TestReplaceMap tests replacing document content
func TestReplaceMap(t *testing.T) {
	t.Run("removes keys not in values", func(t *testing.T) {
		doc, err := ParseString(`name = "test"
version = "1.0.0"
other = "value"`)
		if err != nil {
			t.Fatalf("ParseString() error = %v", err)
		}

		values := map[string]any{
			"name": "updated",
		}

		err = doc.ReplaceMap(values)
		if err != nil {
			t.Fatalf("ReplaceMap() error = %v", err)
		}

		if got, _ := doc.Get("name"); got != "updated" {
			t.Errorf("name = %v, want %q", got, "updated")
		}
		if doc.Has("version") {
			t.Error("version should have been removed")
		}
		if doc.Has("other") {
			t.Error("other should have been removed")
		}
	})

	t.Run("adds new keys", func(t *testing.T) {
		doc, err := ParseString(`name = "test"`)
		if err != nil {
			t.Fatalf("ParseString() error = %v", err)
		}

		values := map[string]any{
			"name":    "test",
			"version": "1.0.0",
		}

		err = doc.ReplaceMap(values)
		if err != nil {
			t.Fatalf("ReplaceMap() error = %v", err)
		}

		if got, _ := doc.Get("name"); got != "test" {
			t.Errorf("name = %v, want %q", got, "test")
		}
		if got, _ := doc.Get("version"); got != "1.0.0" {
			t.Errorf("version = %v, want %q", got, "1.0.0")
		}
	})

	t.Run("handles empty document", func(t *testing.T) {
		doc, err := ParseString("")
		if err != nil {
			t.Fatalf("ParseString() error = %v", err)
		}

		values := map[string]any{
			"name": "test",
		}

		err = doc.ReplaceMap(values)
		if err != nil {
			t.Fatalf("ReplaceMap() error = %v", err)
		}

		if got, _ := doc.Get("name"); got != "test" {
			t.Errorf("name = %v, want %q", got, "test")
		}
	})

	t.Run("returns error for nil document", func(t *testing.T) {
		var doc *Document
		values := map[string]any{"test": "value"}

		err := doc.ReplaceMap(values)
		if err == nil {
			t.Error("expected error for nil document, got nil")
		}
		if !strings.Contains(err.Error(), "nil document") {
			t.Errorf("error should mention nil document, got: %v", err)
		}
	})
}

// TestFlattenValues tests flattening nested maps
func TestFlattenValues(t *testing.T) {
	t.Run("flattens simple map", func(t *testing.T) {
		values := map[string]any{
			"name":    "test",
			"version": "1.0.0",
		}

		result := flattenValues(values)
		if len(result) != 2 {
			t.Errorf("len(result) = %d, want 2", len(result))
		}
		if result["name"] != "test" {
			t.Errorf("name = %v, want %q", result["name"], "test")
		}
		if result["version"] != "1.0.0" {
			t.Errorf("version = %v, want %q", result["version"], "1.0.0")
		}
	})

	t.Run("flattens nested map", func(t *testing.T) {
		values := map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		}

		result := flattenValues(values)
		if result["database.host"] != "localhost" {
			t.Errorf("database.host = %v, want %q", result["database.host"], "localhost")
		}
		if result["database.port"] != 5432 {
			t.Errorf("database.port = %v, want 5432", result["database.port"])
		}
	})

	t.Run("handles nil map", func(t *testing.T) {
		result := flattenValues(nil)
		if result == nil {
			t.Error("result should not be nil")
		}
		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
	})

	t.Run("handles special characters in keys", func(t *testing.T) {
		values := map[string]any{
			"key.with.dots":   "value1",
			"key with spaces": "value2",
		}

		result := flattenValues(values)
		// Keys with special characters should be quoted
		found := false
		for k := range result {
			if strings.Contains(k, "dots") || strings.Contains(k, "spaces") {
				found = true
				break
			}
		}
		if !found {
			t.Error("should handle special characters in keys")
		}
	})

	t.Run("handles deeply nested maps", func(t *testing.T) {
		values := map[string]any{
			"level1": map[string]any{
				"level2": map[string]any{
					"level3": "value",
				},
			},
		}

		result := flattenValues(values)
		if result["level1.level2.level3"] != "value" {
			t.Errorf("deeply nested value not flattened correctly")
		}
	})
}

// TestNormalizeValues tests value normalization
func TestNormalizeValues(t *testing.T) {
	t.Run("normalizes simple values", func(t *testing.T) {
		values := map[string]any{
			"string": "test",
			"number": 42,
			"bool":   true,
		}

		result := normalizeValues(values)
		if result["string"] != "test" {
			t.Errorf("string = %v, want %q", result["string"], "test")
		}
		if result["number"] != 42 {
			t.Errorf("number = %v, want 42", result["number"])
		}
		if result["bool"] != true {
			t.Errorf("bool = %v, want true", result["bool"])
		}
	})

	t.Run("normalizes nested maps", func(t *testing.T) {
		values := map[string]any{
			"outer": map[string]any{
				"inner": "value",
			},
		}

		result := normalizeValues(values)
		outer, ok := result["outer"].(map[string]any)
		if !ok {
			t.Fatal("outer should be a map")
		}
		if outer["inner"] != "value" {
			t.Errorf("inner = %v, want %q", outer["inner"], "value")
		}
	})

	t.Run("handles nil map", func(t *testing.T) {
		result := normalizeValues(nil)
		if result != nil {
			t.Errorf("result = %v, want nil", result)
		}
	})

	t.Run("normalizes arrays", func(t *testing.T) {
		values := map[string]any{
			"items": []any{"a", "b", "c"},
		}

		result := normalizeValues(values)
		items, ok := result["items"].([]any)
		if !ok {
			t.Fatal("items should be an array")
		}
		if len(items) != 3 {
			t.Errorf("len(items) = %d, want 3", len(items))
		}
	})

	t.Run("handles dotted keys", func(t *testing.T) {
		values := map[string]any{
			"database.host": "localhost",
			"database.port": 5432,
		}

		result := normalizeValues(values)
		// Should create nested structure
		if database, ok := result["database"].(map[string]any); ok {
			if database["host"] != "localhost" {
				t.Errorf("database.host = %v, want %q", database["host"], "localhost")
			}
			if database["port"] != 5432 {
				t.Errorf("database.port = %v, want 5432", database["port"])
			}
		} else {
			t.Error("database should be a nested map")
		}
	})
}

// TestHelperFunctions tests utility functions
func TestHelperFunctions(t *testing.T) {
	t.Run("isBareKey identifies bare keys", func(t *testing.T) {
		tests := []struct {
			key  string
			want bool
		}{
			{"simple", true},
			{"with_underscore", true},
			{"with-dash", true},
			{"with123numbers", true},
			{"", false},
			{"with.dot", false},
			{"with space", false},
			{"with\"quote", false},
		}

		for _, tt := range tests {
			got := isBareKey(tt.key)
			if got != tt.want {
				t.Errorf("isBareKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		}
	})

	t.Run("formatKeySegment quotes special keys", func(t *testing.T) {
		tests := []struct {
			key  string
			want string
		}{
			{"simple", "simple"},
			{"with_underscore", "with_underscore"},
			{"with-dash", "with-dash"},
			{"with.dot", `"with.dot"`},
			{"with space", `"with space"`},
		}

		for _, tt := range tests {
			got := formatKeySegment(tt.key)
			if got != tt.want {
				t.Errorf("formatKeySegment(%q) = %q, want %q", tt.key, got, tt.want)
			}
		}
	})

	t.Run("ensureMap returns non-nil map", func(t *testing.T) {
		result := ensureMap(nil)
		if result == nil {
			t.Error("ensureMap(nil) should return non-nil map")
		}

		existingMap := map[string]any{"test": "value"}
		result = ensureMap(existingMap)
		if result["test"] != "value" {
			t.Error("ensureMap should preserve existing values")
		}
	})

	t.Run("setNestedValue creates nested structure", func(t *testing.T) {
		m := make(map[string]any)
		key, err := parser.ParseKey("database.host")
		if err != nil {
			t.Fatalf("ParseKey() error = %v", err)
		}

		setNestedValue(m, key, "localhost")

		database, ok := m["database"].(map[string]any)
		if !ok {
			t.Fatal("database should be a map")
		}
		if database["host"] != "localhost" {
			t.Errorf("host = %v, want %q", database["host"], "localhost")
		}
	})

	t.Run("setNestedValue handles empty key", func(t *testing.T) {
		m := make(map[string]any)
		key := parser.Key{}

		// Should not panic
		setNestedValue(m, key, "value")
	})
}

// Helper function to read file contents
func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	return content
}
