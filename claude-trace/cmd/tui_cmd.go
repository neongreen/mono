package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neongreen/mono/claude-trace/pkg/storage"
	"github.com/neongreen/mono/claude-trace/pkg/tui"
	"github.com/spf13/cobra"
)

// runTUI starts the TUI interface for viewing and annotating traces
func runTUI(cmd *cobra.Command, args []string) {
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
	traces, err := storage.LoadTraces(tracePaths)
	if err != nil {
		log.Fatalf("Error loading traces: %v", err)
	}
	if len(traces) == 0 {
		fmt.Println("No trace files found in discovered locations.")
		os.Exit(1)
	}
	model := tui.NewModel(traces)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}
