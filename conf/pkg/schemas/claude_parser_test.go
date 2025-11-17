package schemas

import (
	"slices"
	"testing"

	"github.com/neongreen/mono/lib/configschema"
)

func TestClaudeSchemaParser_ValidatePath(t *testing.T) {
	parser, err := NewClaudeSchemaParser()
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
			name:     "valid top-level path - model",
			path:     "model",
			expected: true,
		},
		{
			name:     "valid top-level path - hooks",
			path:     "hooks",
			expected: true,
		},
		{
			name:     "valid top-level path - env",
			path:     "env",
			expected: true,
		},
		{
			name:     "valid path - alwaysThinkingEnabled",
			path:     "alwaysThinkingEnabled",
			expected: true,
		},
		{
			name:     "invalid path",
			path:     "invalid.path",
			expected: false,
		},
		{
			name:     "nonexistent top-level path (allowed by additionalProperties)",
			path:     "nonexistent",
			expected: true, // Claude schema has additionalProperties: true
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

func TestClaudeSchemaParser_GetCompletionOptions(t *testing.T) {
	parser, err := NewClaudeSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		minCount       int // Minimum expected count
		expectedNames  []string
		unexpectedName string
	}{
		{
			name:          "top-level options",
			path:          "",
			minCount:      15, // At least 15 top-level properties
			expectedNames: []string{"model", "hooks", "env", "alwaysThinkingEnabled"},
		},
		{
			name:           "hooks has no simple nested options",
			path:           "hooks",
			minCount:       0,
			unexpectedName: "anything",
		},
		{
			name:           "model has no nested options",
			path:           "model",
			minCount:       0,
			unexpectedName: "anything",
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

			// Check unexpected name is not present
			if tt.unexpectedName != "" && len(options) > 0 {
				for _, option := range options {
					if option.Name == tt.unexpectedName {
						t.Errorf("Unexpected option %q found in results", tt.unexpectedName)
					}
				}
			}
		})
	}
}

func TestClaudeSchemaParser_GetAllPaths(t *testing.T) {
	parser, err := NewClaudeSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	paths := parser.GetAllPaths()

	// Check that we have the expected paths from official schema
	expectedPaths := []string{
		"model",
		"hooks",
		"env",
		"alwaysThinkingEnabled",
		"apiKeyHelper",
		"outputStyle",
	}

	if len(paths) < 15 {
		t.Errorf("Expected at least 15 paths from official schema, got %d", len(paths))
	}

	for _, expectedPath := range expectedPaths {
		found := slices.Contains(paths, expectedPath)
		if !found {
			t.Errorf("Expected path %q not found in results", expectedPath)
		}
	}
}

func TestClaudeSchemaParser_GetAllSettingsWithInfo(t *testing.T) {
	parser, err := NewClaudeSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	settings := parser.GetAllSettingsWithInfo()

	if len(settings) == 0 {
		t.Fatal("Expected settings to be non-empty")
	}

	// Check for specific settings from official schema
	var modelSetting *configschema.SettingInfo
	var hooksSetting *configschema.SettingInfo

	for i := range settings {
		if settings[i].Path == "model" {
			modelSetting = &settings[i]
		}
		if settings[i].Path == "hooks" {
			hooksSetting = &settings[i]
		}
	}

	if modelSetting == nil {
		t.Error("Expected to find 'model' setting")
	} else {
		if modelSetting.Type != "string" {
			t.Errorf("Expected model type to be 'string', got %q", modelSetting.Type)
		}
		if modelSetting.Description == "" {
			t.Error("Expected model to have a description")
		}
	}

	if hooksSetting == nil {
		t.Error("Expected to find 'hooks' setting")
	} else {
		if hooksSetting.Type != "object" {
			t.Errorf("Expected hooks type to be 'object', got %q", hooksSetting.Type)
		}
	}
}
