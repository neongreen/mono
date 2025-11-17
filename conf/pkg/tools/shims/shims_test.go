package shims

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShimsTool_SetDryRun(t *testing.T) {
	tool, err := NewShimsTool()
	if err != nil {
		t.Fatalf("Failed to create shims tool: %v", err)
	}

	// Test setting dry-run to true
	tool.SetDryRun(true)
	if !tool.IsDryRun() {
		t.Error("Expected dry-run to be true")
	}

	// Test setting dry-run to false
	tool.SetDryRun(false)
	if tool.IsDryRun() {
		t.Error("Expected dry-run to be false")
	}
}

func TestShimsTool_CreateShim(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "shims-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tool := &ShimsTool{
		shimsDir: tempDir,
		dryRun:   false,
	}

	// Test creating a simple shim
	err = tool.CreateShim("ll", "ls -la --color=auto")
	if err != nil {
		t.Fatalf("Failed to create shim: %v", err)
	}

	// Verify shim file exists
	shimPath := filepath.Join(tempDir, "ll")
	if _, err := os.Stat(shimPath); os.IsNotExist(err) {
		t.Error("Expected shim file to be created")
	}

	// Verify shim content
	content, err := os.ReadFile(shimPath)
	if err != nil {
		t.Fatalf("Failed to read shim file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "#!/bin/bash") {
		t.Error("Expected shim to have bash shebang")
	}
	if !strings.Contains(contentStr, "# Managed by conf") {
		t.Error("Expected shim to have management comment")
	}
	if !strings.Contains(contentStr, "exec ls -la --color=auto") {
		t.Error("Expected shim to contain exec command")
	}

	// Verify shim is executable
	info, err := os.Stat(shimPath)
	if err != nil {
		t.Fatalf("Failed to stat shim file: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Error("Expected shim to be executable")
	}
}

func TestShimsTool_CreateShimDryRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "shims-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tool := &ShimsTool{
		shimsDir: tempDir,
		dryRun:   true,
	}

	// Test dry-run creation
	err = tool.CreateShim("gst", "git status")
	if err != nil {
		t.Fatalf("Failed to create shim in dry-run: %v", err)
	}

	// Verify shim file was NOT created
	shimPath := filepath.Join(tempDir, "gst")
	if _, err := os.Stat(shimPath); !os.IsNotExist(err) {
		t.Error("Expected shim file to NOT be created in dry-run mode")
	}
}

func TestShimsTool_CreateShimDuplicate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "shims-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tool := &ShimsTool{
		shimsDir: tempDir,
		dryRun:   false,
	}

	// Create first shim
	err = tool.CreateShim("test", "echo test")
	if err != nil {
		t.Fatalf("Failed to create first shim: %v", err)
	}

	// Try to create duplicate - should fail
	err = tool.CreateShim("test", "echo duplicate")
	if err == nil {
		t.Error("Expected error when creating duplicate shim")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected error about existing shim, got: %v", err)
	}
}

func TestShimsTool_RemoveShim(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "shims-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tool := &ShimsTool{
		shimsDir: tempDir,
		dryRun:   false,
	}

	// Create a shim first
	err = tool.CreateShim("test", "echo test")
	if err != nil {
		t.Fatalf("Failed to create shim: %v", err)
	}

	// Remove the shim
	err = tool.RemoveShim("test")
	if err != nil {
		t.Fatalf("Failed to remove shim: %v", err)
	}

	// Verify shim file is gone
	shimPath := filepath.Join(tempDir, "test")
	if _, err := os.Stat(shimPath); !os.IsNotExist(err) {
		t.Error("Expected shim file to be removed")
	}
}

func TestShimsTool_RemoveNonExistentShim(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "shims-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tool := &ShimsTool{
		shimsDir: tempDir,
		dryRun:   false,
	}

	// Try to remove non-existent shim
	err = tool.RemoveShim("nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent shim")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error about non-existent shim, got: %v", err)
	}
}

func TestShimsTool_ListShims(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "shims-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tool := &ShimsTool{
		shimsDir: tempDir,
		dryRun:   false,
	}

	// Initially empty
	shims, err := tool.ListShims()
	if err != nil {
		t.Fatalf("Failed to list shims: %v", err)
	}
	if len(shims) != 0 {
		t.Error("Expected empty shims list initially")
	}

	// Create some shims
	err = tool.CreateShim("ll", "ls -la")
	if err != nil {
		t.Fatalf("Failed to create ll shim: %v", err)
	}

	err = tool.CreateShim("gst", "git status")
	if err != nil {
		t.Fatalf("Failed to create gst shim: %v", err)
	}

	// List shims
	shims, err = tool.ListShims()
	if err != nil {
		t.Fatalf("Failed to list shims: %v", err)
	}

	if len(shims) != 2 {
		t.Errorf("Expected 2 shims, got %d", len(shims))
	}

	// Verify shim information
	shimMap := make(map[string]string)
	for _, shim := range shims {
		shimMap[shim.Name] = shim.Command
	}

	if shimMap["ll"] != "ls -la" {
		t.Errorf("Expected ll command to be 'ls -la', got %v", shimMap["ll"])
	}

	if shimMap["gst"] != "git status" {
		t.Errorf("Expected gst command to be 'git status', got %v", shimMap["gst"])
	}
}

func TestShimsTool_ValidateShimName(t *testing.T) {
	tool := &ShimsTool{}

	// Valid names
	validNames := []string{"ll", "gst", "test-command", "test_command", "k8s"}
	for _, name := range validNames {
		if err := tool.validateShimName(name); err != nil {
			t.Errorf("Expected name '%s' to be valid, got error: %v", name, err)
		}
	}

	// Invalid names
	invalidNames := map[string]string{
		"":          "empty",
		"test/path": "contains slash",
		"test\\win": "contains backslash",
		".hidden":   "starts with dot",
	}

	for name, reason := range invalidNames {
		if err := tool.validateShimName(name); err == nil {
			t.Errorf("Expected name '%s' to be invalid (%s), but it was accepted", name, reason)
		}
	}
}

func TestShimsTool_GenerateShimContent(t *testing.T) {
	tool := &ShimsTool{}

	content, err := tool.generateShimContent("test", "echo hello")
	if err != nil {
		t.Fatalf("Failed to generate shim content: %v", err)
	}

	// Verify content structure
	if !strings.Contains(content, "#!/bin/bash") {
		t.Error("Expected shebang line")
	}
	if !strings.Contains(content, "# Managed by conf") {
		t.Error("Expected management comment")
	}
	if !strings.Contains(content, "# Shim: test") {
		t.Error("Expected shim name comment")
	}
	if !strings.Contains(content, "# Command: echo hello") {
		t.Error("Expected command comment")
	}
	if !strings.Contains(content, "exec echo hello \"$@\"") {
		t.Error("Expected exec line with argument passing")
	}
}
