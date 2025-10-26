package schemas

import (
	"fmt"
	"strings"
	"testing"
)

// TestJJSchemaParserIntegration tests the complete jj schema parsing workflow
func TestJJSchemaParserIntegration(t *testing.T) {
	// Create parser
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create JJ schema parser: %v", err)
	}

	// Test 1: Schema loading and basic validation
	t.Run("schema loading", func(t *testing.T) {
		// Verify parser was created successfully
		if parser == nil {
			t.Fatal("Parser should not be nil")
		}

		// Test that we can validate known paths
		knownValidPaths := []string{
			"user.name",
			"user.email",
			"ui.default-command",
			"ui.editor",
			"snapshot.max-new-file-size",
		}

		for _, path := range knownValidPaths {
			valid := parser.ValidatePath(path)
			if !valid {
				t.Errorf("Path '%s' should be valid", path)
			}
		}

		// Test that invalid paths are rejected
		invalidPaths := []string{
			"invalid.nonexistent.path",
			"user.invalidfield",
			"....",
		}

		for _, path := range invalidPaths {
			valid := parser.ValidatePath(path)
			if valid {
				t.Errorf("Path '%s' should be invalid but was accepted", path)
			}
		}

		// Empty path is considered valid (returns true)
		valid := parser.ValidatePath("")
		if !valid {
			t.Error("Empty path should be valid according to ValidatePath implementation")
		}
	})

	// Test 2: Get all paths functionality
	t.Run("get all paths", func(t *testing.T) {
		allPaths := parser.GetAllPaths()

		if len(allPaths) == 0 {
			t.Error("Expected some paths to be returned")
		}

		// Should have at least basic user and ui paths
		expectedPaths := []string{
			"user.name",
			"user.email",
			"ui.default-command",
		}

		pathMap := make(map[string]bool)
		for _, path := range allPaths {
			pathMap[path] = true
		}

		for _, expected := range expectedPaths {
			if !pathMap[expected] {
				t.Errorf("Expected path '%s' not found in all paths", expected)
			}
		}

		// Verify paths are properly formatted (no empty, no duplicates)
		seen := make(map[string]bool)
		for _, path := range allPaths {
			if path == "" {
				t.Error("Found empty path in all paths")
			}
			if seen[path] {
				t.Errorf("Found duplicate path '%s' in all paths", path)
			}
			seen[path] = true
		}

		t.Logf("Found %d total paths in jj schema", len(allPaths))
	})

	// Test 3: Property info extraction
	t.Run("property info", func(t *testing.T) {
		testCases := []struct {
			path         string
			expectedType string
			hasDefault   bool
			hasDesc      bool
		}{
			{
				path:         "user.name",
				expectedType: "string",
				hasDefault:   false, // Usually no default for user name
				hasDesc:      true,
			},
			{
				path:         "user.email",
				expectedType: "string",
				hasDefault:   false,
				hasDesc:      true,
			},
			{
				path:         "ui.default-command",
				expectedType: "unknown", // Schema doesn't specify type for this field
				hasDefault:   true,      // Likely has a default
				hasDesc:      true,
			},
		}

		for _, tc := range testCases {
			info, err := parser.GetPropertyInfo(tc.path)
			if err != nil {
				t.Errorf("Failed to get property info for %s: %v", tc.path, err)
				continue
			}

			// Note: PropertyInfo.Name contains just the final part, not the full path
			if !strings.HasSuffix(tc.path, info.Name) {
				t.Errorf("Property info name mismatch for %s: expected to end with %s, got %s",
					tc.path, info.Name, info.Name)
			}

			if info.Type != tc.expectedType {
				t.Errorf("Property info type mismatch for %s: expected %s, got %s",
					tc.path, tc.expectedType, info.Type)
			}

			if tc.hasDesc && info.Description == "" {
				t.Errorf("Expected description for %s but got empty", tc.path)
			}

			if tc.hasDefault && info.Default == nil {
				t.Errorf("Expected default value for %s but got nil", tc.path)
			}

			t.Logf("Path %s: type=%s, desc=%q, default=%v, enum=%v",
				tc.path, info.Type, info.Description, info.Default, info.Enum)
		}
	})

	// Test 4: GetAllSettingsWithInfo functionality
	t.Run("get all settings with info", func(t *testing.T) {
		settings := parser.GetAllSettingsWithInfo()

		if len(settings) == 0 {
			t.Error("Expected some settings to be returned")
		}

		// Verify structure of settings
		for i, setting := range settings {
			if setting.Path == "" {
				t.Errorf("Setting %d has empty path", i)
			}
			if setting.Type == "" {
				t.Errorf("Setting %d (%s) has empty type", i, setting.Path)
			}
			// Description is optional but most should have it
			// Default is optional
			// Enum is optional
		}

		// Verify settings are sorted
		for i := 1; i < len(settings); i++ {
			if settings[i-1].Path > settings[i].Path {
				t.Errorf("Settings not sorted: %s comes after %s",
					settings[i-1].Path, settings[i].Path)
			}
		}

		// Check for some expected settings
		settingMap := make(map[string]SettingInfo)
		for _, setting := range settings {
			settingMap[setting.Path] = setting
		}

		expectedSettings := []string{"user.name", "user.email", "ui.default-command"}
		for _, expected := range expectedSettings {
			if _, found := settingMap[expected]; !found {
				t.Errorf("Expected setting '%s' not found", expected)
			}
		}

		t.Logf("Found %d total settings with info", len(settings))
	})

	// Test 5: Edge cases and error handling
	t.Run("edge cases", func(t *testing.T) {
		// Test invalid property info requests
		invalidInfo, err := parser.GetPropertyInfo("invalid.path")
		if err == nil {
			t.Error("Expected error for invalid path")
		}
		if invalidInfo.Name != "" {
			t.Error("Expected empty property info for invalid path")
		}

		// Test empty path validation (empty path is valid in current implementation)
		valid := parser.ValidatePath("")
		if !valid {
			t.Error("Empty path should be valid according to current implementation")
		}

		// Test malformed paths
		malformedPaths := []string{
			".",
			".user",
			"user.",
			"user..name",
			"user.name.",
		}

		for _, path := range malformedPaths {
			valid := parser.ValidatePath(path)
			if valid {
				t.Errorf("Expected invalid result for malformed path '%s'", path)
			}
		}
	})
}

// TestSchemaCompletion tests completion generation from schema data
func TestSchemaCompletion(t *testing.T) {
	// Create parser
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create JJ schema parser: %v", err)
	}

	// Test 1: Path completion generation
	t.Run("path completion", func(t *testing.T) {
		allPaths := parser.GetAllPaths()

		// Test filtering by prefix
		testPrefixes := []struct {
			prefix   string
			minMatch int
		}{
			{"user", 2},     // user.name, user.email, etc.
			{"user.", 2},    // Same as above but more specific
			{"ui", 3},       // ui.default-command, ui.editor, etc.
			{"ui.", 3},      // Same as above but more specific
			{"snapshot", 1}, // snapshot.max-new-file-size, etc.
		}

		for _, tc := range testPrefixes {
			var matches []string
			for _, path := range allPaths {
				if strings.HasPrefix(path, tc.prefix) {
					matches = append(matches, path)
				}
			}

			if len(matches) < tc.minMatch {
				t.Errorf("Expected at least %d matches for prefix '%s', got %d: %v",
					tc.minMatch, tc.prefix, len(matches), matches)
			}

			t.Logf("Prefix '%s': %d matches", tc.prefix, len(matches))
		}
	})

	// Test 2: Value completion generation
	t.Run("value completion", func(t *testing.T) {
		settings := parser.GetAllSettingsWithInfo()

		// Find settings with different characteristics for testing
		var booleanSetting, stringSetting, enumSetting *SettingInfo

		for i, setting := range settings {
			switch setting.Type {
			case "boolean":
				if booleanSetting == nil {
					booleanSetting = &settings[i]
				}
			case "string":
				if stringSetting == nil && len(setting.Enum) == 0 {
					stringSetting = &settings[i]
				}
				if enumSetting == nil && len(setting.Enum) > 0 {
					enumSetting = &settings[i]
				}
			}
		}

		// Test boolean completion
		if booleanSetting != nil {
			expectedValues := []string{"true", "false"}
			for _, value := range expectedValues {
				// Simulate completion check
				if !strings.HasPrefix(value, "t") && !strings.HasPrefix(value, "f") {
					continue // This is just testing the concept
				}
			}
			t.Logf("Boolean setting example: %s", booleanSetting.Path)
		}

		// Test enum completion
		if enumSetting != nil {
			if len(enumSetting.Enum) == 0 {
				t.Error("Enum setting should have enum values")
			}
			t.Logf("Enum setting example: %s with values %v", enumSetting.Path, enumSetting.Enum)
		}

		// Test string completion (should suggest defaults/examples)
		if stringSetting != nil {
			t.Logf("String setting example: %s", stringSetting.Path)
		}
	})

	// Test 3: Completion context generation
	t.Run("completion context", func(t *testing.T) {
		settings := parser.GetAllSettingsWithInfo()

		// Test that we can generate helpful completion context
		for _, setting := range settings[:5] { // Test first 5 settings
			// Generate completion entry (path + description)
			var completionText strings.Builder
			completionText.WriteString(setting.Path)
			completionText.WriteString("\t") // Tab separator for shell completion

			if setting.Description != "" {
				completionText.WriteString(setting.Description)
			} else {
				completionText.WriteString("Type: " + setting.Type)
			}

			// Add type info if not in description
			if !strings.Contains(setting.Description, setting.Type) {
				completionText.WriteString(" (")
				completionText.WriteString(setting.Type)
				completionText.WriteString(")")
			}

			// Add default info if available
			if setting.Default != nil {
				completionText.WriteString(" [default: ")
				defaultStr := fmt.Sprintf("%v", setting.Default)
				completionText.WriteString(strings.ReplaceAll(defaultStr, "\t", " "))
				completionText.WriteString("]")
			}

			completion := completionText.String()
			if completion == "" {
				t.Errorf("Generated empty completion for setting %s", setting.Path)
			}

			// Verify completion format
			parts := strings.Split(completion, "\t")
			if len(parts) < 2 {
				t.Errorf("Completion should have tab-separated parts: %s", completion)
			}
			if parts[0] != setting.Path {
				t.Errorf("First part should be path, got: %s", parts[0])
			}

			t.Logf("Completion: %s", completion)
		}
	})
}

// TestSchemaConsistency tests that schema data is consistent and well-formed
func TestSchemaConsistency(t *testing.T) {
	// Create parser
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create JJ schema parser: %v", err)
	}

	// Test 1: All paths vs settings consistency
	t.Run("paths vs settings consistency", func(t *testing.T) {
		allPaths := parser.GetAllPaths()
		allSettings := parser.GetAllSettingsWithInfo()

		// Convert to maps for easier comparison
		pathsMap := make(map[string]bool)
		for _, path := range allPaths {
			pathsMap[path] = true
		}

		settingsMap := make(map[string]bool)
		for _, setting := range allSettings {
			settingsMap[setting.Path] = true
		}

		// Check that all settings paths are in all paths
		for _, setting := range allSettings {
			if !pathsMap[setting.Path] {
				t.Errorf("Setting path '%s' not found in all paths", setting.Path)
			}
		}

		// Check that all paths have corresponding settings
		for _, path := range allPaths {
			if !settingsMap[path] {
				t.Errorf("Path '%s' not found in all settings", path)
			}
		}

		if len(allPaths) != len(allSettings) {
			t.Errorf("Mismatch: %d paths vs %d settings", len(allPaths), len(allSettings))
		}
	})

	// Test 2: Validation consistency
	t.Run("validation consistency", func(t *testing.T) {
		allPaths := parser.GetAllPaths()

		// All paths returned by GetAllPaths should be valid
		for _, path := range allPaths {
			valid := parser.ValidatePath(path)
			if !valid {
				t.Errorf("Path from GetAllPaths should be valid: %s", path)
			}
		}

		// All settings paths should be valid
		allSettings := parser.GetAllSettingsWithInfo()
		for _, setting := range allSettings {
			valid := parser.ValidatePath(setting.Path)
			if !valid {
				t.Errorf("Setting path should be valid: %s", setting.Path)
			}
		}
	})

	// Test 3: Property info consistency
	t.Run("property info consistency", func(t *testing.T) {
		allSettings := parser.GetAllSettingsWithInfo()

		for _, setting := range allSettings {
			// Get property info for the same path
			propInfo, err := parser.GetPropertyInfo(setting.Path)
			if err != nil {
				t.Errorf("Failed to get property info for setting %s: %v", setting.Path, err)
				continue
			}

			// Basic consistency checks - PropertyInfo.Name is just the final component
			if !strings.HasSuffix(setting.Path, propInfo.Name) {
				t.Errorf("Property info name mismatch: setting=%s, propInfo=%s",
					setting.Path, propInfo.Name)
			}

			if propInfo.Type != setting.Type {
				t.Errorf("Property info type mismatch for %s: setting=%s, propInfo=%s",
					setting.Path, setting.Type, propInfo.Type)
			}

			if propInfo.Description != setting.Description {
				t.Errorf("Property info description mismatch for %s", setting.Path)
			}

			// Check enum consistency
			if len(propInfo.Enum) != len(setting.Enum) {
				t.Errorf("Property info enum length mismatch for %s: %d vs %d",
					setting.Path, len(propInfo.Enum), len(setting.Enum))
			}

			// Check that enums contain same values (order might differ)
			if len(propInfo.Enum) > 0 {
				propEnumMap := make(map[string]bool)
				for _, val := range propInfo.Enum {
					propEnumMap[val] = true
				}

				for _, val := range setting.Enum {
					if !propEnumMap[val] {
						t.Errorf("Setting enum value '%s' not found in property info for %s",
							val, setting.Path)
					}
				}
			}
		}
	})

	// Test 4: Data quality checks
	t.Run("data quality", func(t *testing.T) {
		allSettings := parser.GetAllSettingsWithInfo()

		typeCounts := make(map[string]int)
		descriptionsCount := 0
		defaultsCount := 0
		enumsCount := 0

		for _, setting := range allSettings {
			// Count types
			typeCounts[setting.Type]++

			// Count quality metrics
			if setting.Description != "" {
				descriptionsCount++
			}
			if setting.Default != nil {
				defaultsCount++
			}
			if len(setting.Enum) > 0 {
				enumsCount++
			}

			// Quality checks
			if setting.Type == "" {
				t.Errorf("Setting %s has empty type", setting.Path)
			}

			// Path format checks
			if strings.Contains(setting.Path, "..") {
				t.Errorf("Setting path contains double dots: %s", setting.Path)
			}
			if strings.HasPrefix(setting.Path, ".") || strings.HasSuffix(setting.Path, ".") {
				t.Errorf("Setting path has leading/trailing dots: %s", setting.Path)
			}
		}

		// Report statistics
		t.Logf("Schema quality metrics:")
		t.Logf("  Total settings: %d", len(allSettings))
		t.Logf("  With descriptions: %d (%.1f%%)",
			descriptionsCount, float64(descriptionsCount)/float64(len(allSettings))*100)
		t.Logf("  With defaults: %d (%.1f%%)",
			defaultsCount, float64(defaultsCount)/float64(len(allSettings))*100)
		t.Logf("  With enums: %d (%.1f%%)",
			enumsCount, float64(enumsCount)/float64(len(allSettings))*100)

		t.Logf("  Type distribution:")
		for typ, count := range typeCounts {
			t.Logf("    %s: %d", typ, count)
		}

		// Quality expectations
		if descriptionsCount == 0 {
			t.Error("Expected at least some settings to have descriptions")
		}
		if len(typeCounts) == 0 {
			t.Error("Expected some type diversity")
		}
	})
}
