package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestPrivateReleaseDownload tests the full end-to-end flow of downloading
// and executing a binary from a private GitHub release.
//
// This test is skipped by default and only runs when PRRUN_INTEGRATION_TEST=1
// is set. It requires:
// 1. Access to the private neongreen/mono repository
// 2. A valid GitHub token (GITHUB_TOKEN, MISE_GITHUB_TOKEN, or gh CLI auth)
// 3. A PR with a release that has assets
func TestPrivateReleaseDownload(t *testing.T) {
	if os.Getenv("PRRUN_INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test (requires PRRUN_INTEGRATION_TEST=1 and access to private repo)")
	}

	// Test against PR #45 which should have a dissect release
	prURL := "https://github.com/neongreen/mono/pull/45"
	project := "dissect"

	// Build the prrun binary for testing
	cmd := exec.Command("go", "build", "-o", "prrun-test", ".")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build prrun: %v", err)
	}
	defer os.Remove("prrun-test")

	// Test downloading the binary (should not fail with 404)
	downloadCmd := exec.Command("./prrun-test", prURL, project, "--", "--help")
	downloadCmd.Dir = "."
	output, err := downloadCmd.CombinedOutput()

	if err != nil {
		t.Errorf("prrun failed: %v\nOutput: %s", err, string(output))
		return
	}

	// Verify the binary was downloaded and executed successfully
	// The --help flag should produce output, indicating the binary works
	if len(output) == 0 {
		t.Error("Expected help output from dissect binary, got empty output")
	}

	// Verify the binary was cached
	cacheDir, err := getCacheDir()
	if err != nil {
		t.Errorf("Failed to get cache dir: %v", err)
		return
	}

	// Look for cached binary (exact name depends on release)
	cacheFiles, err := filepath.Glob(filepath.Join(cacheDir, "dissect-*"))
	if err != nil {
		t.Errorf("Failed to check cache files: %v", err)
		return
	}

	if len(cacheFiles) == 0 {
		t.Error("Expected binary to be cached, but no files found in cache directory")
	}
}
