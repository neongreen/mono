package editors

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTOMLEditor_SetValue(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "toml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.toml")
	editor := NewTOMLEditor(testFile)

	// Test setting simple value
	err = editor.SetValue("user.name", "Alice")
	if err != nil {
		t.Fatalf("Failed to set simple value: %v", err)
	}

	// Verify value was set correctly by reading it back
	value, err := editor.GetValue("user.name")
	if err != nil {
		t.Fatalf("Failed to read back user.name: %v", err)
	}
	if value != "Alice" {
		t.Errorf("Expected 'Alice', got %v", value)
	}

	// Test setting nested value
	err = editor.SetValue("snapshot.max-new-file-size", int64(0))
	if err != nil {
		t.Fatalf("Failed to set nested value: %v", err)
	}

	// Test setting another value in existing section
	err = editor.SetValue("user.email", "alice@example.com")
	if err != nil {
		t.Fatalf("Failed to set another value in existing section: %v", err)
	}

	// Verify all values are present by reading them back
	value, err = editor.GetValue("user.email")
	if err != nil {
		t.Fatalf("Failed to read back user.email: %v", err)
	}
	if value != "alice@example.com" {
		t.Errorf("Expected 'alice@example.com', got %v", value)
	}

	value, err = editor.GetValue("snapshot.max-new-file-size")
	if err != nil {
		t.Fatalf("Failed to read back snapshot.max-new-file-size: %v", err)
	}
	if value != int64(0) {
		t.Errorf("Expected 0, got %v", value)
	}
}

func TestTOMLEditor_GetValue(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "toml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.toml")

	// Create test file with known content
	testContent := `[user]
name = "Bob"
email = "bob@example.com"

[snapshot]
max-new-file-size = 1024
`
	err = os.WriteFile(testFile, []byte(testContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	editor := NewTOMLEditor(testFile)

	// Test getting simple string value
	value, err := editor.GetValue("user.name")
	if err != nil {
		t.Fatalf("Failed to get user.name: %v", err)
	}
	if value != "Bob" {
		t.Errorf("Expected 'Bob', got %v", value)
	}

	// Test getting integer value
	value, err = editor.GetValue("snapshot.max-new-file-size")
	if err != nil {
		t.Fatalf("Failed to get snapshot.max-new-file-size: %v", err)
	}
	if value != int64(1024) {
		t.Errorf("Expected 1024, got %v", value)
	}

	// Test getting non-existent value
	_, err = editor.GetValue("nonexistent.key")
	if err == nil {
		t.Error("Should get error for non-existent key")
	}
}

func TestTOMLEditor_UnsetValue(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "toml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.toml")

	// Create test file with known content
	testContent := `[user]
name = "Charlie"
email = "charlie@example.com"

[snapshot]
max-new-file-size = 2048
`
	err = os.WriteFile(testFile, []byte(testContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	editor := NewTOMLEditor(testFile)

	// Test unsetting a value
	err = editor.UnsetValue("user.email")
	if err != nil {
		t.Fatalf("Failed to unset user.email: %v", err)
	}

	// Verify value was removed
	_, err = editor.GetValue("user.email")
	if err == nil {
		t.Error("user.email should be removed")
	}

	// Verify other values still exist
	value, err := editor.GetValue("user.name")
	if err != nil {
		t.Error("user.name should still exist")
	}
	if value != "Charlie" {
		t.Errorf("user.name should still be 'Charlie', got %v", value)
	}

	// Test unsetting non-existent value (should not error)
	err = editor.UnsetValue("nonexistent.key")
	if err != nil {
		t.Errorf("Unsetting non-existent key should not error: %v", err)
	}
}

func TestTOMLEditor_FileCreation(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "toml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test creating file in nested directory
	nestedDir := filepath.Join(tempDir, "nested", "path")
	testFile := filepath.Join(nestedDir, "config.toml")
	editor := NewTOMLEditor(testFile)

	// Should create directories and file
	err = editor.SetValue("test.key", "value")
	if err != nil {
		t.Fatalf("Failed to set value in nested path: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File should be created")
	}

	// Verify content
	value, err := editor.GetValue("test.key")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if value != "value" {
		t.Errorf("Expected 'value', got %v", value)
	}
}

func TestTOMLEditor_PreservesExistingContent(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "toml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.toml")

	// Create test file with existing content
	initialContent := `[existing]
key = "original_value"
other_key = 42

[other_section]
some_setting = true
`
	err = os.WriteFile(testFile, []byte(initialContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write initial test file: %v", err)
	}

	editor := NewTOMLEditor(testFile)

	// Add new value
	err = editor.SetValue("new_section.new_key", "new_value")
	if err != nil {
		t.Fatalf("Failed to set new value: %v", err)
	}

	// Verify existing content is preserved
	value, err := editor.GetValue("existing.key")
	if err != nil {
		t.Error("Existing key should be preserved")
	}
	if value != "original_value" {
		t.Errorf("Existing value should be preserved, got %v", value)
	}

	value, err = editor.GetValue("other_section.some_setting")
	if err != nil {
		t.Error("Existing section should be preserved")
	}
	if value != true {
		t.Errorf("Existing boolean value should be preserved, got %v", value)
	}

	// Verify new value was added
	value, err = editor.GetValue("new_section.new_key")
	if err != nil {
		t.Error("New value should be added")
	}
	if value != "new_value" {
		t.Errorf("New value should be correct, got %v", value)
	}
}

func TestTOMLEditor_GetAllValues(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "toml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.toml")
	editor := NewTOMLEditor(testFile)

	// Test empty file
	values, err := editor.GetAllValues()
	if err != nil {
		t.Fatalf("Failed to get all values from non-existent file: %v", err)
	}
	if len(values) != 0 {
		t.Errorf("Expected empty map, got %d values", len(values))
	}

	// Set some values
	editor.SetValue("user.name", "Alice")
	editor.SetValue("user.email", "alice@example.com")
	editor.SetValue("snapshot.max-new-file-size", int64(1024))
	editor.SetValue("ui.diff-editor", "vimdiff")

	// Get all values
	values, err = editor.GetAllValues()
	if err != nil {
		t.Fatalf("Failed to get all values: %v", err)
	}

	// Verify all values are present in nested representation
	expected := map[string]any{
		"user": map[string]any{
			"name":  "Alice",
			"email": "alice@example.com",
		},
		"snapshot": map[string]any{
			"max-new-file-size": int64(1024),
		},
		"ui": map[string]any{
			"diff-editor": "vimdiff",
		},
	}

	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Expected values %v, got %v", expected, values)
	}
}

func TestTOMLEditor_GetAllValuesWithQuotedKeys(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "toml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.toml")

	// Create test file with quoted keys (like jj aliases)
	testContent := `[user]
name = "Alice"

[aliases]
"." = ["foo"]
".." = ["bar"]
normal = "baz"
`
	err = os.WriteFile(testFile, []byte(testContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	editor := NewTOMLEditor(testFile)

	// Get all values
	values, err := editor.GetAllValues()
	if err != nil {
		t.Fatalf("Failed to get all values: %v", err)
	}

	user, ok := values["user"].(map[string]any)
	if !ok {
		t.Fatalf("Expected user section to be a map, got %T", values["user"])
	}
	if user["name"] != "Alice" {
		t.Errorf("Expected user.name to be Alice, got %v", user["name"])
	}

	aliases, ok := values["aliases"].(map[string]any)
	if !ok {
		t.Fatalf("Expected aliases section to be a map, got %T", values["aliases"])
	}

	if val, ok := aliases["."]; ok {
		if arr, isArr := val.([]any); !isArr {
			t.Errorf("Expected aliases[\".\"] to be array, got %T", val)
		} else if len(arr) != 1 || arr[0] != "foo" {
			t.Errorf("Expected aliases[\".\"] = [\"foo\"], got %v", arr)
		}
	} else {
		t.Errorf("Key aliases[\".\"] not found")
	}

	if val, ok := aliases[".."]; ok {
		if arr, isArr := val.([]any); !isArr {
			t.Errorf("Expected aliases[\"..\"] to be array, got %T", val)
		} else if len(arr) != 1 || arr[0] != "bar" {
			t.Errorf("Expected aliases[\"..\"] = [\"bar\"], got %v", arr)
		}
	} else {
		t.Errorf("Key aliases[\"..\"] not found")
	}

	if val, ok := aliases["normal"]; !ok || val != "baz" {
		t.Errorf("Expected aliases[\"normal\"] = \"baz\", got %v", val)
	}
}
