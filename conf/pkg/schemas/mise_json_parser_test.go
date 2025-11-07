package schemas

import (
	"slices"
	"testing"
)

func TestMiseSchemaParser_ValidatePath(t *testing.T) {
	parser, err := NewMiseSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "empty path",
			path:     "",
			expected: true,
		},
		{
			name:     "valid top-level path - env",
			path:     "env",
			expected: true,
		},
		{
			name:     "valid top-level path - tools",
			path:     "tools",
			expected: true,
		},
		{
			name:     "valid top-level path - settings",
			path:     "settings",
			expected: true,
		},
		{
			name:     "invalid path",
			path:     "invalid.path",
			expected: false,
		},
		{
			name:     "invalid top-level path",
			path:     "nonexistent",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ValidatePath(tt.path)
			if result != tt.expected {
				t.Errorf("ValidatePath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestMiseSchemaParser_GetCompletionOptions(t *testing.T) {
	parser, err := NewMiseSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name          string
		path          string
		minCount      int
		expectedNames []string
	}{
		{
			name:          "top-level options",
			path:          "",
			minCount:      5, // At least 5 top-level properties
			expectedNames: []string{"env", "tools", "settings"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := parser.GetCompletionOptions(tt.path)

			if len(options) < tt.minCount {
				t.Errorf("Expected at least %d options, got %d", tt.minCount, len(options))
			}

			// Check expected names are present
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, option := range options {
					if option.Name == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected option %q not found in results", expectedName)
				}
			}
		})
	}
}

func TestMiseSchemaParser_GetAllPaths(t *testing.T) {
	parser, err := NewMiseSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	paths := parser.GetAllPaths()

	// Check that we have paths from official schema
	expectedPaths := []string{
		"env",
		"tools",
		"settings",
	}

	if len(paths) < 5 {
		t.Errorf("Expected at least 5 paths from official schema, got %d", len(paths))
	}

	for _, expectedPath := range expectedPaths {
		found := slices.Contains(paths, expectedPath)
		if !found {
			t.Errorf("Expected path %q not found in results", expectedPath)
		}
	}
}

func TestMiseSchemaParser_GetAllSettingsWithInfo(t *testing.T) {
	parser, err := NewMiseSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	settings := parser.GetAllSettingsWithInfo()

	if len(settings) == 0 {
		t.Fatal("Expected settings to be non-empty")
	}

	// Check for specific settings from official schema
	var envSetting *SettingInfo
	var toolsSetting *SettingInfo

	for i := range settings {
		if settings[i].Path == "env" {
			envSetting = &settings[i]
		}
		if settings[i].Path == "tools" {
			toolsSetting = &settings[i]
		}
	}

	if envSetting == nil {
		t.Error("Expected to find 'env' setting")
	}

	if toolsSetting == nil {
		t.Error("Expected to find 'tools' setting")
	}
}
