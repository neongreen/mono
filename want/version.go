package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

var (
	// Version information - set via -ldflags at build time
	Version   = "0.1.0-mvp"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func printVersion() {
	printVersionWithJSON(false)
}

func printVersionWithJSON(jsonOutput bool) {
	if jsonOutput {
		// Output as JSON with ISO timestamp
		versionInfo := map[string]string{
			"version":    Version,
			"commit":     GitCommit,
			"build_time": BuildTime,
			"go_version": runtime.Version(),
		}
		output, err := json.MarshalIndent(versionInfo, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to marshal JSON: %v\n", err)
			return
		}
		fmt.Println(string(output))
	} else {
		// Format build time as human-readable
		buildTimeStr := BuildTime
		if t, err := time.Parse(time.RFC3339, BuildTime); err == nil {
			buildTimeStr = t.Format("Jan 2, 2006 15:04 MST")
		}

		fmt.Printf("want version %s\n", Version)
		fmt.Printf("  commit: %s\n", GitCommit)
		fmt.Printf("  built:  %s\n", buildTimeStr)
		fmt.Printf("  go:     %s\n", runtime.Version())
	}
}
