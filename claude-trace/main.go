package main

import (
	"fmt"
	"os"

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

var viewCmd = &cobra.Command{
	Use:   "view <trace-file>",
	Short: "Open a trace in the web viewer",
	Long:  `Open a trace file in a web-based viewer. The viewer will be opened in your default browser.`,
	Run:   runView,
	Args:  cobra.MinimumNArgs(1),
}

var (
	extractOutputDir string
	viewPort         int
)

func init() {
	extractCmd.Flags().StringVarP(&extractOutputDir, "output", "o", "./extracted-traces", "Output directory for extracted traces")
	viewCmd.Flags().IntVarP(&viewPort, "port", "p", 8080, "Port to run the viewer server on")
}

func main() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(extractCmd)
	rootCmd.AddCommand(viewCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
