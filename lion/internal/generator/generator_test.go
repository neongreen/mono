package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neongreen/mono/lion/internal/extractor"
)

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test documentation
	docs := map[string][]extractor.DocEntry{
		"intro": {
			{
				Topic:   "intro",
				Content: "This is the introduction.",
				File:    "test.go",
				Line:    1,
				Entity:  "package test",
			},
		},
		"api": {
			{
				Topic:   "api",
				Content: "Config holds settings.",
				File:    "test.go",
				Line:    5,
				Entity:  "Config",
			},
			{
				Topic:   "api",
				Content: "Initialize creates Config.",
				File:    "test.go",
				Line:    10,
				Entity:  "Initialize",
			},
		},
	}

	// Generate markdown files
	if err := Generate(docs, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify index.md exists
	indexPath := filepath.Join(tmpDir, "index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("index.md was not created")
	}

	// Verify topic files exist
	introPath := filepath.Join(tmpDir, "intro.md")
	if _, err := os.Stat(introPath); os.IsNotExist(err) {
		t.Error("intro.md was not created")
	}

	apiPath := filepath.Join(tmpDir, "api.md")
	if _, err := os.Stat(apiPath); os.IsNotExist(err) {
		t.Error("api.md was not created")
	}

	// Verify index content
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read index.md: %v", err)
	}

	indexStr := string(indexContent)
	if !strings.Contains(indexStr, "# Documentation Index") {
		t.Error("index.md missing title")
	}
	if !strings.Contains(indexStr, "[Api](api.md)") {
		t.Error("index.md missing api link")
	}
	if !strings.Contains(indexStr, "[Intro](intro.md)") {
		t.Error("index.md missing intro link")
	}

	// Verify intro.md content
	introContent, err := os.ReadFile(introPath)
	if err != nil {
		t.Fatalf("failed to read intro.md: %v", err)
	}

	introStr := string(introContent)
	if !strings.Contains(introStr, "# Intro") {
		t.Error("intro.md missing title")
	}
	if !strings.Contains(introStr, "package test") {
		t.Error("intro.md missing entity name")
	}
	if !strings.Contains(introStr, "This is the introduction.") {
		t.Error("intro.md missing content")
	}
	if !strings.Contains(introStr, "`test.go:1`") {
		t.Error("intro.md missing source reference")
	}

	// Verify api.md content
	apiContent, err := os.ReadFile(apiPath)
	if err != nil {
		t.Fatalf("failed to read api.md: %v", err)
	}

	apiStr := string(apiContent)
	if !strings.Contains(apiStr, "# Api") {
		t.Error("api.md missing title")
	}
	if !strings.Contains(apiStr, "## Config") {
		t.Error("api.md missing Config section")
	}
	if !strings.Contains(apiStr, "## Initialize") {
		t.Error("api.md missing Initialize section")
	}
	if !strings.Contains(apiStr, "Config holds settings.") {
		t.Error("api.md missing Config content")
	}
	if !strings.Contains(apiStr, "Initialize creates Config.") {
		t.Error("api.md missing Initialize content")
	}
}

func TestGenerateWithCustomTitles(t *testing.T) {
	tmpDir := t.TempDir()

	docs := map[string][]extractor.DocEntry{
		"api": {
			{
				Topic:         "api",
				Content:       "Content",
				File:          "test.go",
				Line:          5,
				Entity:        "Config",
				TopicTitle:    "Custom API",
				HasTopicTitle: true,
				SectionTitle:  "Custom Section",
				HasSection:    true,
			},
			{
				Topic:   "api",
				Content: "Second",
				File:    "test.go",
				Line:    10,
				Entity:  "SecondFunc",
				// No section title override here; uses entity name.
			},
		},
	}

	if err := Generate(docs, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	apiPath := filepath.Join(tmpDir, "api.md")
	data, err := os.ReadFile(apiPath)
	if err != nil {
		t.Fatalf("failed to read api.md: %v", err)
	}
	out := string(data)

	if !strings.Contains(out, "# Custom API") {
		t.Errorf("expected custom topic title in output, got: %s", out)
	}
	if !strings.Contains(out, "## Custom Section") {
		t.Errorf("expected custom section title, got: %s", out)
	}
	if !strings.Contains(out, "## SecondFunc") {
		t.Errorf("expected default entity heading for second entry, got: %s", out)
	}

	indexData, err := os.ReadFile(filepath.Join(tmpDir, "index.md"))
	if err != nil {
		t.Fatalf("failed to read index.md: %v", err)
	}
	if !strings.Contains(string(indexData), "[Custom API](api.md)") {
		t.Errorf("expected custom title in index link, got: %s", string(indexData))
	}
}

func TestGenerateWarnsOnMissingHeading(t *testing.T) {
	tmpDir := t.TempDir()
	var buf bytes.Buffer
	orig := warningWriter
	warningWriter = &buf
	defer func() { warningWriter = orig }()

	docs := map[string][]extractor.DocEntry{
		"topic": {
			{
				Topic:   "topic",
				Content: "content without heading",
				File:    "file.go",
				Line:    10,
				// No entity, no section title -> should warn
			},
		},
	}

	if err := Generate(docs, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(buf.String(), "has no heading") {
		t.Fatalf("expected warning about missing heading, got: %q", buf.String())
	}
}

func TestGenerateEmptyDocs(t *testing.T) {
	tmpDir := t.TempDir()

	docs := map[string][]extractor.DocEntry{}

	if err := Generate(docs, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should still create index.md
	indexPath := filepath.Join(tmpDir, "index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("index.md was not created for empty docs")
	}
}

func TestFormatTopic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"api", "Api"},
		{"getting-started", "Getting Started"},
		{"multi-word-topic", "Multi Word Topic"},
		{"simple", "Simple"},
		{"with_underscore", "With Underscore"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatTopic(tt.input)
			if result != tt.expected {
				t.Errorf("formatTopic(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
