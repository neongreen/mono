package main

import (
	"fmt"
	"ingest/pkg/database"
	"ingest/pkg/git"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ingest",
	Short: "A tool to ingest git repository metadata into SQLite",
	Long:  `Ingest walks through git repositories and stores all commit and file metadata into an SQLite database.`,
}

var ingestCmd = &cobra.Command{
	Use:   "ingest [repository-path]",
	Short: "Ingest a git repository",
	Long:  `Ingest walks through all commits in a git repository and stores metadata in the database.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := args[0]

		// Convert to absolute path
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			log.Fatalf("Failed to get absolute path: %v", err)
		}

		// Verify it's a git repository
		if _, err := os.Stat(filepath.Join(absPath, ".git")); os.IsNotExist(err) {
			log.Fatalf("Not a git repository: %s", absPath)
		}

		fmt.Printf("Ingesting repository: %s\n", absPath)

		// Open database
		db, err := database.Open()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()

		// Create a new run
		runID, err := db.CreateRun(absPath)
		if err != nil {
			log.Fatalf("Failed to create run: %v", err)
		}

		fmt.Printf("Started ingestion run #%d\n", runID)

		// Walk the repository
		commits, err := git.WalkRepository(absPath)
		if err != nil {
			db.FinishRun(runID, "failed")
			log.Fatalf("Failed to walk repository: %v", err)
		}

		fmt.Printf("Found %d commits\n", len(commits))

		// Store commits and files
		totalFiles := 0
		for i, commit := range commits {
			if (i+1)%100 == 0 || (i+1) == len(commits) {
				fmt.Printf("Processing commit %d/%d...\r", i+1, len(commits))
			}

			commitID, err := db.CreateCommit(
				runID,
				commit.Hash,
				commit.Author,
				commit.AuthorEmail,
				commit.Date,
				commit.Message,
			)
			if err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to create commit: %v", err)
			}

			for _, file := range commit.Files {
				err := db.CreateFile(commitID, file.Path, file.Size, file.Mode)
				if err != nil {
					db.FinishRun(runID, "failed")
					log.Fatalf("Failed to create file: %v", err)
				}
				totalFiles++
			}
		}

		fmt.Printf("\nProcessed %d commits with %d files\n", len(commits), totalFiles)

		// Update counts and finish run
		if err := db.UpdateRunCounts(runID); err != nil {
			log.Fatalf("Failed to update run counts: %v", err)
		}

		if err := db.FinishRun(runID, "completed"); err != nil {
			log.Fatalf("Failed to finish run: %v", err)
		}

		fmt.Printf("Ingestion completed successfully!\n")
	},
}

var listRunsCmd = &cobra.Command{
	Use:   "list-runs",
	Short: "List all ingestion runs",
	Long:  `Display all ingestion runs with statistics about commits and files.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Open()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()

		runs, err := db.GetAllRuns()
		if err != nil {
			log.Fatalf("Failed to get runs: %v", err)
		}

		if len(runs) == 0 {
			fmt.Println("No ingestion runs found.")
			return
		}

		fmt.Printf("\n%-5s %-20s %-50s %-8s %-10s %-10s\n", "ID", "Start Time", "Repository", "Status", "Commits", "Files")
		fmt.Println("-----------------------------------------------------------------------------------------------------------------------------")

		for _, run := range runs {
			startTime := run.StartTime.Format("2006-01-02 15:04:05")
			duration := ""
			if run.EndTime != nil {
				dur := run.EndTime.Sub(run.StartTime)
				duration = fmt.Sprintf(" (%.1fs)", dur.Seconds())
			}

			// Truncate repo path if too long
			repoPath := run.RepoPath
			if len(repoPath) > 50 {
				repoPath = "..." + repoPath[len(repoPath)-47:]
			}

			fmt.Printf("%-5d %-20s %-50s %-8s %-10d %-10d%s\n",
				run.ID,
				startTime,
				repoPath,
				run.Status,
				run.CommitCount,
				run.FileCount,
				duration,
			)
		}

		fmt.Println()

		// Display summary statistics
		totalCommits := 0
		totalFiles := 0
		completedRuns := 0
		for _, run := range runs {
			totalCommits += run.CommitCount
			totalFiles += run.FileCount
			if run.Status == "completed" {
				completedRuns++
			}
		}

		fmt.Printf("Summary: %d total runs (%d completed), %d commits, %d files\n",
			len(runs), completedRuns, totalCommits, totalFiles)
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(listRunsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
