package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"claude-trace/pkg/render"
	"claude-trace/pkg/storage"
	"claude-trace/pkg/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claude-trace",
	Short: "A tool for viewing and annotating Claude Code traces",
	Long:  `claude-trace is a TUI tool for viewing and annotating Claude Code conversation traces.`,
	Run:   runTUI,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List discovered trace locations and counts",
	Long:  `Show all locations where traces are searched for and how many traces were found in each location.`,
	Run:   runList,
}

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract traces as JSON and Markdown files",
	Long:  `Extract all found traces as structured JSON and rendered Markdown files, one file per trace per format.`,
	Run:   runExtract,
}

var extractOutputDir string

func init() {
	extractCmd.Flags().StringVarP(&extractOutputDir, "output", "o", "./extracted-traces", "Output directory for extracted traces")
}

func runTUI(cmd *cobra.Command, args []string) {
	// Discover Claude Code trace locations
	tracePaths, err := storage.DiscoverTraceLocations()
	if err != nil {
		log.Fatalf("Error discovering trace locations: %v", err)
	}

	if len(tracePaths) == 0 {
		fmt.Println("No Claude Code traces found.")
		fmt.Println("\nSearched in:")
		fmt.Println("  - ~/.claude/projects/ (conversation histories)")
		fmt.Println("  - ~/.claude/debug/ (debug logs)")
		fmt.Println("  - ~/.claude/traces/ (user traces)")
		fmt.Println("  - ~/.config/Claude/traces (legacy)")
		fmt.Println("  - ~/Library/Application Support/Claude/traces (legacy)")
		fmt.Println("  - ~/.local/share/Claude/traces (legacy)")
		fmt.Println("  - ./traces (current directory)")
		os.Exit(1)
	}

	// Load traces
	traces, err := storage.LoadTraces(tracePaths)
	if err != nil {
		log.Fatalf("Error loading traces: %v", err)
	}

	if len(traces) == 0 {
		fmt.Println("No trace files found in discovered locations.")
		os.Exit(1)
	}

	// Start TUI
	model := tui.NewModel(traces)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}

func runList(cmd *cobra.Command, args []string) {
	// Get all searched locations
	allLocations := storage.GetAllSearchedLocations()

	// Get existing trace directories
	existingPaths, err := storage.DiscoverTraceLocations()
	if err != nil {
		log.Fatalf("Error discovering trace locations: %v", err)
	}

	fmt.Println("Searched locations:")

	totalTraces := 0
	foundLocations := 0

	for _, location := range allLocations {
		// Check if this location exists
		exists := false
		for _, existingPath := range existingPaths {
			if existingPath == location {
				exists = true
				break
			}
		}

		if exists {
			// Count traces in this directory
			count, err := storage.CountTracesInDirectory(location)
			if err != nil {
				fmt.Printf("  ✓ %s (error counting: %v)\n", location, err)
			} else {
				fmt.Printf("  ✓ %s (%d traces)\n", location, count)
				totalTraces += count
				foundLocations++
			}
		} else {
			fmt.Printf("  ✗ %s\n", location)
		}
	}

	fmt.Printf("\nTotal: %d traces found in %d locations\n", totalTraces, foundLocations)
}

func runExtract(cmd *cobra.Command, args []string) {
	// Discover Claude Code trace locations
	tracePaths, err := storage.DiscoverTraceLocations()
	if err != nil {
		log.Fatalf("Error discovering trace locations: %v", err)
	}

	if len(tracePaths) == 0 {
		fmt.Println("No Claude Code traces found.")
		os.Exit(1)
	}

	// Load traces
	traces, err := storage.LoadTraces(tracePaths)
	if err != nil {
		log.Fatalf("Error loading traces: %v", err)
	}

	// Load annotations for each trace
	for _, trace := range traces {
		if err := storage.LoadAnnotations(trace); err != nil {
			log.Printf("Warning: failed to load annotations for %s: %v", trace.Name, err)
		}
	}

	if len(traces) == 0 {
		fmt.Println("No trace files found in discovered locations.")
		os.Exit(1)
	}

	// Create output directories
	jsonDir := filepath.Join(extractOutputDir, "json")
	markdownDir := filepath.Join(extractOutputDir, "markdown")

	if err := os.MkdirAll(jsonDir, 0755); err != nil {
		log.Fatalf("Error creating JSON output directory: %v", err)
	}
	if err := os.MkdirAll(markdownDir, 0755); err != nil {
		log.Fatalf("Error creating Markdown output directory: %v", err)
	}

	fmt.Printf("Extracting %d traces to %s\n", len(traces), extractOutputDir)

	// Process each trace
	for i, trace := range traces {
		// Convert to intermediate representation
		traceData := render.ToTraceData(trace)

		// Generate base filename (remove extension and sanitize)
		baseName := strings.TrimSuffix(trace.Name, filepath.Ext(trace.Name))
		baseName = sanitizeFilename(baseName)

		// Render to JSON
		jsonData, err := render.RenderToJSON(traceData)
		if err != nil {
			log.Printf("Error rendering %s to JSON: %v", trace.Name, err)
			continue
		}

		jsonPath := filepath.Join(jsonDir, baseName+".json")
		if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
			log.Printf("Error writing JSON file for %s: %v", trace.Name, err)
			continue
		}

		// Render to Markdown
		markdownData, err := render.RenderToMarkdown(traceData)
		if err != nil {
			log.Printf("Error rendering %s to Markdown: %v", trace.Name, err)
			continue
		}

		markdownPath := filepath.Join(markdownDir, baseName+".md")
		if err := os.WriteFile(markdownPath, markdownData, 0644); err != nil {
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
	// Replace problematic characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}

func main() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(extractCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
