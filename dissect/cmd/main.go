package main

import (
	"dissect/pkg/commands"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang-cz/devslog"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dissect",
	Short: "A tool to refactor Go code by extracting functions to separate files.",
	Long: `Dissect is a CLI tool that helps refactor Go code by extracting top-level functions 
into their own files, following Go's best practices for code organization.`,
}

var splitCmd = &cobra.Command{
	Use:   "split [paths]...",
	Short: "Split Go files by extracting functions to separate files",
	Long: `Split processes Go files and extracts each top-level function into its own file,
following Go's best practices for code organization.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			slog.Error("Error getting current working directory", "error", err)
			os.Exit(1)
		}

		var filesToProcess []string
		for _, path := range args {
			absPath, err := filepath.Abs(path)
			if err != nil {
				slog.Error("Error getting absolute path", "path", path, "error", err)
				continue
			}

			info, err := os.Stat(absPath)
			if err != nil {
				slog.Error("Error getting file info", "path", absPath, "workingDir", cwd, "error", err)
				continue
			}

			if info.IsDir() {
				goFiles, err := commands.FindGoFiles(absPath)
				if err != nil {
					slog.Error("Error finding Go files in directory", "path", absPath, "error", err)
					continue
				}
				filesToProcess = append(filesToProcess, goFiles...)
			} else {
				if strings.HasSuffix(info.Name(), ".go") {
					filesToProcess = append(filesToProcess, absPath)
				}
			}
		}

		// Remove duplicates from the list of strings
		uniqueFiles := make(map[string]struct{})
		for _, file := range filesToProcess {
			uniqueFiles[file] = struct{}{}
		}
		filesToProcess = []string{}
		for file := range uniqueFiles {
			filesToProcess = append(filesToProcess, file)
		}
		sort.Strings(filesToProcess)

		slog.Info("Found Go files", "count", len(filesToProcess))

		refactored := 0
		nothingToDo := 0
		skipped := 0
		failed := 0

		for _, file := range filesToProcess {
			result, _, _ := ProcessFile(file) // Ignore `err` since it's logged inside ProcessFile already
			switch result {
			case Refactored:
				refactored++
			case NothingToDo:
				nothingToDo++
			case Skipped:
				skipped++
			case Failed:
				slog.Error("Failed to process file", "file", file)
				failed++
			}
		}

		slog.Info("Files were processed",
			"refactored", refactored,
			"skipped", skipped,
			"failed", failed,
			"nothingToDo", nothingToDo,
		)
	},
}

func initLogging() {
	programLevel := new(slog.Level)
	if err := programLevel.UnmarshalText([]byte(os.Getenv("LOG_LEVEL"))); err != nil {
		programLevel = new(slog.Level)
		*programLevel = slog.LevelInfo // Default to info if not set or invalid
	}

	// Only show timestamps in debug mode
	timeFormat := "-"
	if *programLevel == slog.LevelDebug {
		timeFormat = "[15:04:05]"
	}

	slog.SetDefault(slog.New(
		devslog.NewHandler(
			os.Stdout,
			&devslog.Options{
				NewLineAfterLog: true,
				TimeFormat:      timeFormat,
				HandlerOptions: &slog.HandlerOptions{
					Level:     programLevel,
					AddSource: *programLevel == slog.LevelDebug,
				},
			},
		),
	))
	slog.Debug("Logging level set", "level", programLevel.String())
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(splitCmd)
	rootCmd.AddCommand(moveCmd)
}

func main() {
	initLogging()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
