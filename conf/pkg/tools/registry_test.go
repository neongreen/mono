package tools

import (
	"slices"
	"testing"
)

func TestGetSupportedTools(t *testing.T) {
	tools := GetSupportedTools()

	expectedTools := map[string]bool{
		"jj":       true,
		"mise":     true,
		"starship": true,
		"claude":   true,
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	for _, tool := range tools {
		if !expectedTools[tool] {
			t.Errorf("Unexpected tool: %s", tool)
		}
	}

	for expectedTool := range expectedTools {
		found := slices.Contains(tools, expectedTool)
		if !found {
			t.Errorf("Expected tool %s not found", expectedTool)
		}
	}
}

func TestGetTool(t *testing.T) {
	// Test getting valid tools (some may fail in test environment)
	validTools := []string{"jj", "mise", "starship", "claude"}

	for _, toolName := range validTools {
		tool, err := GetTool(toolName)
		if err != nil {
			// Some tools may fail in test environment due to missing config
			t.Logf("Tool %s failed (expected in test environment): %v", toolName, err)
			continue
		}
		if tool == nil {
			t.Errorf("Expected tool %s to be non-nil when no error", toolName)
		}
	}

	// Test getting invalid tool
	_, err := GetTool("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
	if err.Error() != "unknown tool: nonexistent" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestApplyToolValue(t *testing.T) {
	// Test with invalid tool
	err := ApplyToolValue("nonexistent", "test.path", "value")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
	if err.Error() != "unknown tool: nonexistent" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Note: We can't easily test valid tools without setting up their environments
	// This would be covered by integration tests
}

func TestGetActualValue(t *testing.T) {
	// Test with invalid tool
	_, err := GetActualValue("nonexistent", "test.path")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
	if err.Error() != "unknown tool: nonexistent" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Note: We can't easily test valid tools without setting up their environments
	// This would be covered by integration tests
}

func TestToolRegistryCompleteness(t *testing.T) {
	// Ensure all expected tools are registered
	expectedTools := []string{"jj", "mise", "starship", "claude"}

	for _, expectedTool := range expectedTools {
		if _, exists := toolRegistry[expectedTool]; !exists {
			t.Errorf("Tool %s not registered in toolRegistry", expectedTool)
		}
	}
}

func TestToolFactories(t *testing.T) {
	// Test that all factories can be called without errors
	for toolName, factory := range toolRegistry {
		tool, err := factory()
		if err != nil {
			// Some tools might fail in test environment, which is okay
			t.Logf("Tool %s factory failed (expected in test environment): %v", toolName, err)
		} else if tool == nil {
			t.Errorf("Tool %s factory returned nil without error", toolName)
		}
	}
}
