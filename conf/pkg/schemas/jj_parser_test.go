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
