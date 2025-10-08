package main

import (
	"fmt"
	"log"
	"os"

	"claude-trace/pkg/storage"
	"claude-trace/pkg/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Discover Claude Code trace locations
	tracePaths, err := storage.DiscoverTraceLocations()
	if err != nil {
		log.Fatalf("Error discovering trace locations: %v", err)
	}

	if len(tracePaths) == 0 {
		fmt.Println("No Claude Code traces found.")
		fmt.Println("\nSearched in:")
		fmt.Println("  - ~/.config/Claude/traces")
		fmt.Println("  - ~/Library/Application Support/Claude/traces")
		fmt.Println("  - ~/.local/share/Claude/traces")
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
