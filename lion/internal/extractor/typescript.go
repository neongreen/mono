package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// tsExtractResult represents the JSON output from the TypeScript extractor.
type tsExtractResult struct {
	Entries []tsDocEntry `json:"entries"`
}

// tsDocEntry represents a single documentation entry from TypeScript.
type tsDocEntry struct {
	Topic        string `json:"topic"`
	Content      string `json:"content"`
	File         string `json:"file"`
	Line         int    `json:"line"`
	Entity       string `json:"entity"`
	TopicTitle   string `json:"topicTitle,omitempty"`
	SectionTitle string `json:"sectionTitle,omitempty"`
}

// findTSHelper locates the TypeScript helper relative to the lion binary or source.
func findTSHelper() (string, error) {
	// First, try to find it relative to the executable
	execPath, err := os.Executable()
	if err == nil {
		helperPath := filepath.Join(filepath.Dir(execPath), "ts-helper", "dist", "index.js")
		if _, err := os.Stat(helperPath); err == nil {
			return helperPath, nil
		}
	}

	// Try relative to current working directory (for development)
	cwd, err := os.Getwd()
	if err == nil {
		// Try lion/ts-helper (when running from repo root)
		helperPath := filepath.Join(cwd, "lion", "ts-helper", "dist", "index.js")
		if _, err := os.Stat(helperPath); err == nil {
			return helperPath, nil
		}
		// Try ts-helper (when running from lion directory)
		helperPath = filepath.Join(cwd, "ts-helper", "dist", "index.js")
		if _, err := os.Stat(helperPath); err == nil {
			return helperPath, nil
		}
	}

	// Try to find using go list to get the current module path
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	output, err := cmd.Output()
	if err == nil {
		modDir := strings.TrimSpace(string(output))
		helperPath := filepath.Join(modDir, "lion", "ts-helper", "dist", "index.js")
		if _, err := os.Stat(helperPath); err == nil {
			return helperPath, nil
		}
	}

	return "", fmt.Errorf("TypeScript helper not found. Run 'npm install && npm run build' in lion/ts-helper")
}

// ExtractTypeScript extracts lion documentation from TypeScript files in a directory.
// It uses the TypeScript compiler API via a Node.js helper script.
func ExtractTypeScript(dir string) (map[string][]DocEntry, error) {
	// Check if there are any TypeScript files first
	hasTS := false
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx")) {
			// Skip test and declaration files
			if !strings.HasSuffix(path, ".test.ts") &&
				!strings.HasSuffix(path, ".test.tsx") &&
				!strings.HasSuffix(path, ".spec.ts") &&
				!strings.HasSuffix(path, ".spec.tsx") &&
				!strings.HasSuffix(path, ".d.ts") {
				hasTS = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return nil, err
	}

	if !hasTS {
		return nil, nil
	}

	// Find the TypeScript helper
	helperPath, err := findTSHelper()
	if err != nil {
		return nil, err
	}

	// Run the TypeScript extractor
	cmd := exec.Command("node", helperPath, dir)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("TypeScript extractor failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run TypeScript extractor: %w", err)
	}

	// Parse the JSON output
	var result tsExtractResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TypeScript extractor output: %w", err)
	}

	// Convert to DocEntry map
	docs := make(map[string][]DocEntry)
	for _, entry := range result.Entries {
		docEntry := DocEntry{
			Topic:         entry.Topic,
			Content:       entry.Content,
			File:          entry.File,
			Line:          entry.Line,
			Entity:        entry.Entity,
			TopicTitle:    entry.TopicTitle,
			HasTopicTitle: entry.TopicTitle != "",
			SectionTitle:  entry.SectionTitle,
			HasSection:    entry.SectionTitle != "",
		}
		docs[entry.Topic] = append(docs[entry.Topic], docEntry)
	}

	return docs, nil
}
