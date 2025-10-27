package schemas

import (
	"testing"
)

func TestNewClaudeSchemaParser(t *testing.T) {
	parser, err := NewClaudeSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create Claude schema parser: %v", err)
	}

	if parser == nil {
		t.Fatal("Expected parser to be non-nil")
	}

	if parser.schema == nil {
		t.Fatal("Expected schema to be non-nil")
	}
}

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
			name:     "valid top-level path",
			path:     "model",
			expected: true,
		},
		{
			name:     "valid nested path",
			path:     "api.key",
			expected: true,
		},
		{
			name:     "another valid nested path",
			path:     "api.url",
			expected: true,
		},
		{
			name:     "valid simple path",
			path:     "temperature",
			expected: true,
		},
		{
			name:     "invalid path",
			path:     "invalid.path",
			expected: false,
		},
		{
			name:     "invalid nested path",
			path:     "api.invalid",
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

func TestClaudeSchemaParser_GetCompletionOptions(t *testing.T) {
	parser, err := NewClaudeSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		expectedCount  int
		expectedNames  []string
		unexpectedName string
	}{
		{
			name:          "top-level options",
			path:          "",
			expectedCount: 8, // api, model, max_tokens, temperature, top_p, top_k, stop_sequences, system
			expectedNames: []string{"model", "temperature", "api"},
		},
		{
			name:          "api nested options",
			path:          "api",
			expectedCount: 2, // key, url
			expectedNames: []string{"key", "url"},
		},
		{
			name:           "model has no nested options",
			path:           "model",
			expectedCount:  0,
			unexpectedName: "anything",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := parser.GetCompletionOptions(tt.path)

			if len(options) != tt.expectedCount {
				t.Errorf("Expected %d options, got %d", tt.expectedCount, len(options))
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
			if tt.unexpectedName != "" {
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

	// Check that we have the expected paths
	expectedPaths := []string{
		"model",
		"max_tokens",
		"temperature",
		"api",
		"api.key",
		"api.url",
		"top_p",
		"top_k",
		"stop_sequences",
		"system",
	}

	if len(paths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
	}

	for _, expectedPath := range expectedPaths {
		found := false
		for _, path := range paths {
			if path == expectedPath {
				found = true
				break
			}
		}
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

	// Check for specific settings
	var modelSetting *SettingInfo
	var apiKeySetting *SettingInfo

	for i := range settings {
		if settings[i].Path == "model" {
			modelSetting = &settings[i]
		}
		if settings[i].Path == "api.key" {
			apiKeySetting = &settings[i]
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

	if apiKeySetting == nil {
		t.Error("Expected to find 'api.key' setting")
	} else {
		if apiKeySetting.Type != "string" {
			t.Errorf("Expected api.key type to be 'string', got %q", apiKeySetting.Type)
		}
	}
}
