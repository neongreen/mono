package main

import (
	"fmt"
	"log"
	"slices"

	"github.com/neongreen/mono/claude-trace/pkg/storage"
	"github.com/spf13/cobra"
)

// runList shows all locations where traces are searched for and how many traces were found in each location
func runList(cmd *cobra.Command, args []string) {
	allLocations := storage.GetAllSearchedLocations()
	existingPaths, err := storage.DiscoverTraceLocations()
	if err != nil {
		log.Fatalf("Error discovering trace locations: %v", err)
	}
	fmt.Println("Searched locations:")
	totalTraces := 0
	foundLocations := 0
	for _, location := range allLocations {
		exists := slices.Contains(existingPaths, location)
		if exists {
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
