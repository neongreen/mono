package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	if path == "" {
		t.Error("GetConfigPath() returned empty path")
	}

	// Should contain expected components
	if !filepath.IsAbs(path) {
		t.Errorf("GetConfigPath() = %v, want absolute path", path)
	}

	expectedSuffix := filepath.Join(".config", "tk", "config.toml")
	if !endsWithPath(path, expectedSuffix) {
		t.Errorf("GetConfigPath() = %v, want path ending with %v", path, expectedSuffix)
	}
}

func TestGetStateDir(t *testing.T) {
	dir, err := GetStateDir()
	if err != nil {
		t.Fatalf("GetStateDir() error = %v", err)
	}

	if dir == "" {
		t.Error("GetStateDir() returned empty path")
	}

	// Should be absolute path
	if !filepath.IsAbs(dir) {
		t.Errorf("GetStateDir() = %v, want absolute path", dir)
	}

	// Should have been created
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("GetStateDir() directory not created: %v", dir)
	}

	expectedSuffix := filepath.Join(".local", "state", "tk")
	if !endsWithPath(dir, expectedSuffix) {
		t.Errorf("GetStateDir() = %v, want path ending with %v", dir, expectedSuffix)
	}
}

func TestLoadConfig_DefaultConfig(t *testing.T) {
	// Set HOME to a temporary directory to avoid using real config
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Check defaults
	if config.Remotes == nil {
		t.Error("LoadConfig() config.Remotes is nil")
	}

	if config.Sync.SegmentMaxBytes == 0 {
		t.Error("LoadConfig() config.Sync.SegmentMaxBytes not set")
	}

	if config.Blocking.BlockingAxis == "" {
		t.Error("LoadConfig() config.Blocking.BlockingAxis not set")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Set HOME to a temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	// Create a test config
	testConfig := &Config{
		Remotes: map[string]RemoteConfig{
			"test": {
				Type:   "folder",
				Path:   "/tmp/test-remote",
				Spaces: []string{"test-space"},
				Push:   true,
				Pull:   true,
			},
		},
		Sync:     DefaultSyncConfig(),
		Blocking: DefaultBlockingConfig(),
	}

	// Save the config
	if err := SaveConfig(testConfig); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Load it back
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() after save error = %v", err)
	}

	// Verify the loaded config
	if len(loadedConfig.Remotes) != 1 {
		t.Errorf("LoadConfig() remotes count = %v, want %v", len(loadedConfig.Remotes), 1)
	}

	remote, ok := loadedConfig.Remotes["test"]
	if !ok {
		t.Error("LoadConfig() missing 'test' remote")
	} else {
		if remote.Type != "folder" {
			t.Errorf("LoadConfig() remote.Type = %v, want %v", remote.Type, "folder")
		}
		if remote.Path != "/tmp/test-remote" {
			t.Errorf("LoadConfig() remote.Path = %v, want %v", remote.Path, "/tmp/test-remote")
		}
		if len(remote.Spaces) != 1 || remote.Spaces[0] != "test-space" {
			t.Errorf("LoadConfig() remote.Spaces = %v, want %v", remote.Spaces, []string{"test-space"})
		}
	}
}

// Helper function to check if a path ends with expected components
func endsWithPath(fullPath, suffix string) bool {
	// Normalize paths for comparison
	fullPath = filepath.Clean(fullPath)
	suffix = filepath.Clean(suffix)
	return filepath.Base(fullPath) == filepath.Base(suffix) &&
		(len(fullPath) >= len(suffix) && fullPath[len(fullPath)-len(suffix):] == suffix ||
			filepath.Dir(fullPath) == filepath.Dir(suffix) ||
			endsWithPath(filepath.Dir(fullPath), filepath.Dir(suffix)))
}
