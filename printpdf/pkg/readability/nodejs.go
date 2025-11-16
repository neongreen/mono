package readability

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed bundles/mozilla/dist/mozilla.bundle.js
var mozillaBundleJS []byte

// NodeJSEngine runs JavaScript readability engines using Node.js
type NodeJSEngine struct {
	name       string
	bundleData []byte
}

// NewMozillaEngine creates a Mozilla Readability engine
func NewMozillaEngine() *NodeJSEngine {
	return &NodeJSEngine{
		name:       "mozilla",
		bundleData: mozillaBundleJS,
	}
}

// Name returns the engine name
func (e *NodeJSEngine) Name() string {
	return e.name
}

// Extract extracts readable content using Node.js
func (e *NodeJSEngine) Extract(html []byte, sourceURL string) ([]byte, error) {
	// Check Node.js availability first
	if err := e.IsAvailable(); err != nil {
		return nil, err
	}

	// Create a temporary file for the bundle
	tmpDir := os.TempDir()
	bundlePath := filepath.Join(tmpDir, fmt.Sprintf("readability-%s.bundle.js", e.name))

	// Write bundle to temp file
	if err := os.WriteFile(bundlePath, e.bundleData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write bundle: %w", err)
	}
	defer os.Remove(bundlePath)

	// Run Node.js with the bundle
	cmd := exec.Command("node", bundlePath)

	// Create pipes for stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr for error messages
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start node: %w", err)
	}

	// Write HTML to stdin
	if _, err := stdin.Write(html); err != nil {
		return nil, fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdin.Close()

	// Read stdout
	output, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdout: %w", err)
	}

	// Read stderr
	stderrOutput, err := io.ReadAll(stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to read stderr: %w", err)
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		if len(stderrOutput) > 0 {
			return nil, fmt.Errorf("node process failed: %s", string(stderrOutput))
		}
		return nil, fmt.Errorf("node process failed: %w", err)
	}

	return output, nil
}

// IsAvailable checks if Node.js is installed
func (e *NodeJSEngine) IsAvailable() error {
	cmd := exec.Command("node", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Node.js is required for %s engine but not found in PATH. Please install Node.js from https://nodejs.org/", e.name)
	}
	return nil
}
