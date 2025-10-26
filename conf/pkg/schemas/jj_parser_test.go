package schemas

import (
	"strings"
	"testing"
)

func TestNewJJSchemaParser(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create jj schema parser: %v", err)
	}

	if parser == nil {
		t.Fatal("Parser should not be nil")
	}

	if parser.schema == nil {
		t.Fatal("Parser schema should not be nil")
	}
}

func TestJJSchemaParser_GetCompletionOptions_TopLevel(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	options := parser.GetCompletionOptions("")
	if len(options) == 0 {
		t.Error("Should get top-level completion options")
	}

	// Check for expected top-level properties from jj schema
	var foundUser bool
	for _, option := range options {
		if option.Name == "user" {
			foundUser = true
			if option.Type != "object" {
				t.Errorf("User should be object type, got %s", option.Type)
			}
			if !strings.Contains(option.Description, "user") {
				t.Errorf("User description should contain 'user', got: %s", option.Description)
			}
		}
	}

	if !foundUser {
		t.Error("Should find 'user' in top-level options")
	}
}

func TestJJSchemaParser_GetCompletionOptions_Nested(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test nested completion for user properties
	options := parser.GetCompletionOptions("user")
	if len(options) == 0 {
		t.Error("Should get completion options for user")
	}

	// Check for expected user properties
	var foundName, foundEmail bool
	for _, option := range options {
		if option.Name == "name" {
			foundName = true
			if option.Type != "string" {
				t.Errorf("Name should be string type, got %s", option.Type)
			}
		}
		if option.Name == "email" {
			foundEmail = true
			if option.Type != "string" {
				t.Errorf("Email should be string type, got %s", option.Type)
			}
		}
	}

	if !foundName {
		t.Error("Should find 'name' in user options")
	}
	if !foundEmail {
		t.Error("Should find 'email' in user options")
	}
}

func TestJJSchemaParser_ValidatePath(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test valid paths
	validPaths := []string{
		"",
		"user",
		"user.name",
		"user.email",
	}

	for _, path := range validPaths {
		if !parser.ValidatePath(path) {
			t.Errorf("Path '%s' should be valid", path)
		}
	}

	// Test invalid paths
	invalidPaths := []string{
		"nonexistent",
		"user.nonexistent",
		"user.name.invalid",
	}

	for _, path := range invalidPaths {
		if parser.ValidatePath(path) {
			t.Errorf("Path '%s' should be invalid", path)
		}
	}
}

func TestJJSchemaParser_ValidatePath_AdditionalProperties(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test paths for properties that use additionalProperties
	// These should accept any arbitrary key name
	testCases := []struct {
		path          string
		shouldBeValid bool
		description   string
	}{
		// aliases - accepts any alias name
		{"aliases", true, "aliases object itself"},
		{"aliases.r", true, "arbitrary alias name"},
		{"aliases.rebase", true, "arbitrary alias name"},
		{"aliases.my-custom-command", true, "arbitrary alias name with dash"},

		// revset-aliases - accepts any alias name
		{"revset-aliases", true, "revset-aliases object itself"},
		{"revset-aliases.trunk", true, "arbitrary revset alias"},
		{"revset-aliases.immutable_heads()", true, "predefined revset alias with parens"},
		{"revset-aliases.my-alias", true, "arbitrary revset alias name"},

		// template-aliases - accepts any alias name
		{"template-aliases", true, "template-aliases object itself"},
		{"template-aliases.my-template", true, "arbitrary template alias"},

		// colors - accepts any color label
		{"colors", true, "colors object itself"},
		{"colors.error", true, "arbitrary color label"},
		{"colors.warning", true, "arbitrary color label"},

		// merge-tools - accepts any tool name and has nested properties
		{"merge-tools", true, "merge-tools object itself"},
		{"merge-tools.vimdiff", true, "arbitrary merge tool name"},
		{"merge-tools.vimdiff.program", true, "merge tool property"},
		{"merge-tools.vimdiff.diff-args", true, "merge tool property"},
		{"merge-tools.custom-tool.merge-args", true, "custom merge tool property"},
		{"merge-tools.my-custom-tool", true, "arbitrary merge tool name"},

		// hints - accepts any hint name
		{"hints", true, "hints object itself"},
		{"hints.my-hint", true, "arbitrary hint name"},

		// templates - accepts any template name
		{"templates", true, "templates object itself"},
		{"templates.log", true, "arbitrary template name"},

		// revsets - accepts any revset name
		{"revsets", true, "revsets object itself"},
		{"revsets.log", true, "arbitrary revset name"},

		// These should still be invalid - too deep nesting or invalid base path
		{"aliases.r.invalid", false, "aliases values are arrays, not objects"},
		{"revset-aliases.trunk.invalid", false, "revset-aliases values are strings, not objects"},
		{"nonexistent.path", false, "nonexistent base property"},
		{"merge-tools.vimdiff.nonexistent", false, "merge tool nonexistent property"},
	}

	for _, tc := range testCases {
		result := parser.ValidatePath(tc.path)
		if result != tc.shouldBeValid {
			t.Errorf("Path '%s' (%s): expected valid=%v, got valid=%v",
				tc.path, tc.description, tc.shouldBeValid, result)
		}
	}
}

func TestJJSchemaParser_ValidatePath_QuotedKeys(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test paths with quoted keys (special characters that require quoting in TOML)
	// These are produced by FlattenValues when keys contain special characters
	testCases := []struct {
		path          string
		shouldBeValid bool
		description   string
	}{
		// Quoted keys in aliases (the main issue from the problem statement)
		{`aliases."."`, true, "alias with dot as key name"},
		{`aliases."*"`, true, "alias with asterisk as key name"},
		{`aliases."foo bar"`, true, "alias with space in key name"},
		{`aliases."foo-bar"`, true, "alias with hyphen (could be bare but quoted is also valid)"},
		{`aliases."123"`, true, "alias starting with number"},

		// Quoted keys in other additionalProperties fields
		{`revset-aliases."."`, true, "revset-alias with dot as key"},
		{`template-aliases."."`, true, "template-alias with dot as key"},
		{`colors."."`, true, "color label with dot as key"},
		{`hints."."`, true, "hint with dot as key"},

		// Quoted keys in nested structures
		{`merge-tools."my.tool"`, true, "merge tool name with dot"},
		{`merge-tools."my.tool".program`, true, "merge tool with dot in name, accessing property"},

		// Invalid cases - too deep nesting
		{`aliases.".".invalid`, false, "aliases values are arrays, cannot nest further"},
		{`revset-aliases.".".invalid`, false, "revset-aliases values are strings, cannot nest further"},
	}

	for _, tc := range testCases {
		result := parser.ValidatePath(tc.path)
		if result != tc.shouldBeValid {
			t.Errorf("Path '%s' (%s): expected valid=%v, got valid=%v",
				tc.path, tc.description, tc.shouldBeValid, result)
		}
	}
}

func TestJJSchemaParser_GetAllPaths(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	paths := parser.GetAllPaths()
	if len(paths) == 0 {
		t.Error("Should get some paths from schema")
	}

	// Check that paths are sorted
	for i := 1; i < len(paths); i++ {
		if paths[i-1] > paths[i] {
			t.Error("Paths should be sorted")
			break
		}
	}

	// Check for expected paths
	expectedPaths := []string{"user", "user.name", "user.email"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Should find path '%s'", expected)
		}
	}
}

func TestJJSchemaParser_GetPropertyInfo(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test getting info for user.name
	info, err := parser.GetPropertyInfo("user.name")
	if err != nil {
		t.Fatalf("Failed to get property info for user.name: %v", err)
	}

	if info.Name != "name" {
		t.Errorf("Property name should be 'name', got '%s'", info.Name)
	}

	if info.Type != "string" {
		t.Errorf("Property type should be 'string', got '%s'", info.Type)
	}

	if info.Description == "" {
		t.Error("Property should have description")
	}

	// Test invalid property
	_, err = parser.GetPropertyInfo("invalid.path")
	if err == nil {
		t.Error("Should get error for invalid property path")
	}

	// Test empty path
	_, err = parser.GetPropertyInfo("")
	if err == nil {
		t.Error("Should get error for empty path")
	}
}

func TestJJSchemaParser_OptionsAreSorted(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	options := parser.GetCompletionOptions("")
	if len(options) < 2 {
		t.Skip("Need at least 2 options to test sorting")
	}

	// Check that options are sorted by name
	for i := 1; i < len(options); i++ {
		if options[i-1].Name > options[i].Name {
			t.Error("Completion options should be sorted by name")
			break
		}
	}
}

func TestJJSchemaParser_HandlesMissingProperties(t *testing.T) {
	parser, err := NewJJSchemaParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test completion for non-existent path
	options := parser.GetCompletionOptions("nonexistent.path")
	if len(options) != 0 {
		t.Error("Should get empty options for non-existent path")
	}

	// Test deeply nested non-existent path
	options = parser.GetCompletionOptions("user.nonexistent.deep")
	if len(options) != 0 {
		t.Error("Should get empty options for deeply nested non-existent path")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test getTypeFromProperty
	prop := map[string]interface{}{
		"type":        "string",
		"description": "Test property",
	}

	if getTypeFromProperty(prop) != "string" {
		t.Error("Should extract type correctly")
	}

	if getDescriptionFromProperty(prop) != "Test property" {
		t.Error("Should extract description correctly")
	}

	// Test with missing properties
	emptyProp := map[string]interface{}{}
	if getTypeFromProperty(emptyProp) != "unknown" {
		t.Error("Should return 'unknown' for missing type")
	}

	if getDescriptionFromProperty(emptyProp) != "" {
		t.Error("Should return empty string for missing description")
	}

	// Test getDefaultFromProperty
	propWithDefault := map[string]interface{}{
		"default": "test_value",
	}

	if getDefaultFromProperty(propWithDefault) != "test_value" {
		t.Error("Should extract default value correctly")
	}

	if getDefaultFromProperty(emptyProp) != nil {
		t.Error("Should return nil for missing default")
	}

	// Test getEnumFromProperty
	propWithEnum := map[string]interface{}{
		"enum": []interface{}{"option1", "option2", "option3"},
	}

	enum := getEnumFromProperty(propWithEnum)
	if len(enum) != 3 {
		t.Errorf("Should extract 3 enum values, got %d", len(enum))
	}

	if enum[0] != "option1" {
		t.Errorf("First enum value should be 'option1', got '%s'", enum[0])
	}
}
