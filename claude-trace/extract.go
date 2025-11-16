package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/claude-trace/pkg/render"
	"github.com/neongreen/mono/claude-trace/pkg/storage"
	"github.com/spf13/cobra"
)

// runExtract extracts all found traces as structured JSON and rendered Markdown files
func runExtract(cmd *cobra.Command, args []string) {
	tracePaths, err := storage.DiscoverTraceLocations()
	if err != nil {
		log.Fatalf("Error discovering trace locations: %v", err)
	}
	if len(tracePaths) == 0 {
		fmt.Println("No Claude Code traces found.")
		os.Exit(1)
	}
	traces, err := storage.LoadTraces(tracePaths)
	if err != nil {
		log.Fatalf("Error loading traces: %v", err)
	}
	for _, trace := range traces {
		if err := storage.LoadAnnotations(trace); err != nil {
			log.Printf("Warning: failed to load annotations for %s: %v", trace.Name, err)
		}
	}
	if len(traces) == 0 {
		fmt.Println("No trace files found in discovered locations.")
		os.Exit(1)
	}
	jsonDir := filepath.Join(extractOutputDir, "json")
	markdownDir := filepath.Join(extractOutputDir, "markdown")
	if err := os.MkdirAll(jsonDir, 0o755); err != nil {
		log.Fatalf("Error creating JSON output directory: %v", err)
	}
	if err := os.MkdirAll(markdownDir, 0o755); err != nil {
		log.Fatalf("Error creating Markdown output directory: %v", err)
	}
	fmt.Printf("Extracting %d traces to %s\n", len(traces), extractOutputDir)
	for i, trace := range traces {
		traceData := render.ToTraceData(trace)
		baseName := strings.TrimSuffix(trace.Name, filepath.Ext(trace.Name))
		baseName = sanitizeFilename(baseName)
		jsonData, err := render.RenderToJSON(traceData)
		if err != nil {
			log.Printf("Error rendering %s to JSON: %v", trace.Name, err)
			continue
		}
		jsonPath := filepath.Join(jsonDir, baseName+".json")
		if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
			log.Printf("Error writing JSON file for %s: %v", trace.Name, err)
			continue
		}
		markdownData, err := render.RenderToMarkdown(traceData)
		if err != nil {
			log.Printf("Error rendering %s to Markdown: %v", trace.Name, err)
			continue
		}
		markdownPath := filepath.Join(markdownDir, baseName+".md")
		if err := os.WriteFile(markdownPath, markdownData, 0o644); err != nil {
			log.Printf("Error writing Markdown file for %s: %v", trace.Name, err)
			continue
		}
		fmt.Printf("  [%d/%d] Extracted: %s\n", i+1, len(traces), trace.Name)
	}
	fmt.Printf("\nSuccessfully extracted %d traces:\n", len(traces))
	fmt.Printf("  - JSON files: %s\n", jsonDir)
	fmt.Printf("  - Markdown files: %s\n", markdownDir)
}

// sanitizeFilename removes or replaces characters that are problematic in filenames
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer("/", "_",
		"\\", "_",
		":", "_", "*", "_", "?", "_", "\"", "_", "<", "_",
		">",
		"_", "|", "_")
	return replacer.Replace(name)
}
