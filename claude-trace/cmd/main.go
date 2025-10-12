package main

import (
	"fmt"
	"log"
	"os"

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

func main() {
	rootCmd.AddCommand(listCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
