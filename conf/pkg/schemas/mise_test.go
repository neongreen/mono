package schemas

import (
	"strings"
	"testing"
)

func TestLoadMiseSchema(t *testing.T) {
	schema, err := LoadMiseSchema()
	if err != nil {
		t.Fatalf("Failed to load mise schema: %v", err)
	}

	if schema == nil {
		t.Fatal("Mise schema should not be nil")
	}

	if schema.Schema.Name != "mise" {
		t.Errorf("Schema name should be 'mise', got '%s'", schema.Schema.Name)
	}

	if len(schema.Fields) == 0 {
		t.Error("Schema should have fields defined")
	}

	// Check for expected top-level fields
	expectedFields := []string{"tasks", "env", "tools", "settings"}
	for _, fieldName := range expectedFields {
		if _, exists := schema.Fields[fieldName]; !exists {
			t.Errorf("Schema should contain field '%s'", fieldName)
		}
	}
}

func TestMiseSchemaEmbedded(t *testing.T) {
	if len(MiseSchemaData) == 0 {
		t.Fatal("Mise schema data should not be empty")
	}

	// Check that it contains expected TOML content
	if !strings.Contains(MiseSchemaData, "[schema]") {
		t.Error("Mise schema should contain [schema] section")
	}

	if !strings.Contains(MiseSchemaData, "[fields]") {
		t.Error("Mise schema should contain [fields] section")
	}

	if !strings.Contains(MiseSchemaData, "tasks") {
		t.Error("Mise schema should contain tasks field")
	}
}

func TestMiseSchemaCompletionOptions(t *testing.T) {
	schema, err := LoadMiseSchema()
	if err != nil {
		t.Fatalf("Failed to load mise schema: %v", err)
	}

	// Test top-level completion
	options := schema.GetCompletionOptions("")
	if len(options) == 0 {
		t.Error("Should get completion options for top level")
	}

	// Check that we get expected top-level options
	var foundTasks bool
	for _, option := range options {
		if option.Name == "tasks" {
			foundTasks = true
			if option.Type != "object" {
				t.Errorf("Tasks field should be object type, got '%s'", option.Type)
			}
			if option.Description == "" {
				t.Error("Tasks field should have description")
			}
		}
	}

	if !foundTasks {
		t.Error("Should find 'tasks' in top-level completion options")
	}

	// Test nested completion for tasks
	taskOptions := schema.GetCompletionOptions("tasks")
	if len(taskOptions) == 0 {
		t.Error("Should get completion options for tasks field")
	}

	// Check for expected task properties
	var foundRun bool
	for _, option := range taskOptions {
		if option.Name == "run" {
			foundRun = true
			if option.Type != "string" {
				t.Errorf("Run property should be string type, got '%s'", option.Type)
			}
		}
	}

	if !foundRun {
		t.Error("Should find 'run' property in tasks completion")
	}
}

func TestMiseSchemaSettings(t *testing.T) {
	schema, err := LoadMiseSchema()
	if err != nil {
		t.Fatalf("Failed to load mise schema: %v", err)
	}

	// Test settings field specifically
	settingsField, exists := schema.Fields["settings"]
	if !exists {
		t.Fatal("Settings field should exist")
	}

	if settingsField.Type != "object" {
		t.Errorf("Settings should be object type, got '%s'", settingsField.Type)
	}

	if len(settingsField.Properties) == 0 {
		t.Error("Settings should have properties defined")
	}

	// Check for some expected settings
	settingsOptions := schema.GetCompletionOptions("settings")
	var foundExperimental bool
	for _, option := range settingsOptions {
		if option.Name == "experimental" {
			foundExperimental = true
			if option.Type != "boolean" {
				t.Errorf("Experimental setting should be boolean, got '%s'", option.Type)
			}
		}
	}

	if !foundExperimental {
		t.Error("Should find 'experimental' in settings options")
	}
}
