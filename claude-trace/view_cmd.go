package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/neongreen/mono/claude-trace/pkg/storage"
	"github.com/neongreen/mono/claude-trace/pkg/viewer"
	"github.com/spf13/cobra"
)

func runView(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: trace file path required")
		fmt.Fprintln(os.Stderr, "Usage: claude-trace view <trace-file>")
		os.Exit(1)
	}

	tracePath := args[0]

	// Load the trace
	content, err := os.ReadFile(tracePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading trace file: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(tracePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading trace file info: %v\n", err)
		os.Exit(1)
	}

	trace := &storage.Trace{
		Path:         tracePath,
		Name:         info.Name(),
		Content:      string(content),
		ModTime:      info.ModTime(),
		Annotations:  []storage.Annotation{},
		Tags:         make(map[string]bool),
		FreeformNote: "",
	}

	// Load annotations if they exist
	if err := storage.LoadAnnotations(trace); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load annotations: %v\n", err)
	}

	// Start the server
	port := viewPort
	server := viewer.NewServer(trace, port)

	// Open browser after a short delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		url := fmt.Sprintf("http://127.0.0.1:%d", port)
		if err := openBrowser(url); err != nil {
			fmt.Printf("Server started at %s\n", url)
			fmt.Println("Please open the URL in your browser")
		}
	}()

	// Start the server (blocks)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}

// openBrowser opens the default browser to the given URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}
