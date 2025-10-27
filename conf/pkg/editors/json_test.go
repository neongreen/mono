package editors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONEditor_SetValue(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	editor := NewJSONEditor(testFile)

	tests := []struct {
		name     string
		path     string
		value    interface{}
		expected map[string]interface{}
	}{
		{
			name:  "set simple string value",
			path:  "user.name",
			value: "Alice",
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
				},
			},
		},
		{
			name:  "set boolean value",
			path:  "settings.enabled",
			value: true,
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
				},
				"settings": map[string]interface{}{
					"enabled": true,
				},
			},
		},
		{
			name:  "set number value",
			path:  "settings.timeout",
			value: float64(30),
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
				},
				"settings": map[string]interface{}{
					"enabled": true,
					"timeout": float64(30),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := editor.SetValue(tt.path, tt.value); err != nil {
				t.Fatalf("SetValue failed: %v", err)
			}

			// Read back and verify
			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(content, &result); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			if !deepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestJSONEditor_GetValue(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	// Create initial JSON file
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
		},
		"settings": map[string]interface{}{
			"enabled": true,
			"timeout": float64(30),
		},
	}

	content, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(testFile, content, 0644)

	editor := NewJSONEditor(testFile)

	tests := []struct {
		name     string
		path     string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "get string value",
			path:     "user.name",
			expected: "Alice",
			wantErr:  false,
		},
		{
			name:     "get boolean value",
			path:     "settings.enabled",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "get number value",
			path:     "settings.timeout",
			expected: float64(30),
			wantErr:  false,
		},
		{
			name:     "get nested object",
			path:     "user",
			expected: map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
			wantErr:  false,
		},
		{
			name:     "get non-existent path",
			path:     "user.age",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := editor.GetValue(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetValue error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !deepEqual(value, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, value)
			}
		})
	}
}

func TestJSONEditor_UnsetValue(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	// Create initial JSON file
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
		},
		"settings": map[string]interface{}{
			"enabled": true,
		},
	}

	content, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(testFile, content, 0644)

	editor := NewJSONEditor(testFile)

	// Unset user.email
	if err := editor.UnsetValue("user.email"); err != nil {
		t.Fatalf("UnsetValue failed: %v", err)
	}

	// Verify
	value, err := editor.GetValue("user.email")
	if err == nil {
		t.Errorf("Expected error for unset value, got value: %v", value)
	}

	// Verify user.name still exists
	value, err = editor.GetValue("user.name")
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}
	if value != "Alice" {
		t.Errorf("Expected 'Alice', got %v", value)
	}

	// Unset entire settings section
	if err := editor.UnsetValue("settings"); err != nil {
		t.Fatalf("UnsetValue failed: %v", err)
	}

	// Verify settings is gone
	value, err = editor.GetValue("settings")
	if err == nil {
		t.Errorf("Expected error for unset section, got value: %v", value)
	}
}

func TestJSONEditor_GetAllValues(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	// Create initial JSON file
	expected := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
		},
		"settings": map[string]interface{}{
			"enabled": true,
		},
	}

	content, _ := json.MarshalIndent(expected, "", "  ")
	os.WriteFile(testFile, content, 0644)

	editor := NewJSONEditor(testFile)

	values, err := editor.GetAllValues()
	if err != nil {
		t.Fatalf("GetAllValues failed: %v", err)
	}

	if !deepEqual(values, expected) {
		t.Errorf("Expected %v, got %v", expected, values)
	}
}

func TestJSONEditor_SetAllValues(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	editor := NewJSONEditor(testFile)

	values := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "Bob",
			"email": "bob@example.com",
		},
		"settings": map[string]interface{}{
			"theme": "dark",
		},
	}

	if err := editor.SetAllValues(values); err != nil {
		t.Fatalf("SetAllValues failed: %v", err)
	}

	// Read back and verify
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if !deepEqual(result, values) {
		t.Errorf("Expected %v, got %v", values, result)
	}
}

func TestJSONEditor_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	editor := NewJSONEditorWithDryRun(testFile, true)

	// SetValue in dry-run mode should not create the file
	if err := editor.SetValue("user.name", "Alice"); err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("File should not exist in dry-run mode")
	}
}

// Helper function for deep equality comparison
func deepEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}
