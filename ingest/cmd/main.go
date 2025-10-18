package main

import (
	"context"
	"encoding/json"
	"fmt"
	"ingest/pkg/command"
	"ingest/pkg/database"
	"ingest/pkg/fs"
	"ingest/pkg/git"
	"ingest/pkg/github"
	mcppkg "ingest/pkg/mcp"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = newRootCmd()

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "A tool to ingest git repository metadata into SQLite",
		Long:  `Ingest walks through git repositories and stores all commit and file metadata into an SQLite database.`,
	}

	cmd.AddCommand(newGitCmd())
	cmd.AddCommand(newFSCmd())
	cmd.AddCommand(newCmdCmd())
	cmd.AddCommand(newGitHubCmd())
	cmd.AddCommand(newListRunsCmd())
	cmd.AddCommand(newQueryCmd())
	cmd.AddCommand(newMCPCmd())

	return cmd
}

func newGitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "git [repository-path]",
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
			runID, err := db.CreateRun(absPath, "git")
			if err != nil {
				log.Fatalf("Failed to create run: %v", err)
			}

			fmt.Printf("Started ingestion run #%d\n", runID)

			// Get repository metadata
			fmt.Println("Collecting repository metadata...")
			metadata, err := git.GetRepoMetadata(absPath)
			if err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to get repository metadata: %v", err)
			}

			// Store remotes
			for _, remote := range metadata.Remotes {
				err := db.CreateGitRemote(runID, remote.Name, remote.URL)
				if err != nil {
					db.FinishRun(runID, "failed")
					log.Fatalf("Failed to create remote: %v", err)
				}
			}

			// Store refs (branches and tags)
			for _, ref := range metadata.Refs {
				err := db.CreateGitRef(runID, ref.Type, ref.Name, ref.TargetHash)
				if err != nil {
					db.FinishRun(runID, "failed")
					log.Fatalf("Failed to create ref: %v", err)
				}
			}

			fmt.Printf("Found %d remotes and %d refs\n", len(metadata.Remotes), len(metadata.Refs))

			fmt.Println("Looking for commits...")

			// Walk the repository with progress callback
			commits, err := git.WalkRepository(absPath, func(count int) {
				fmt.Printf("Found %d commits so far...\r", count)
			})
			if err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to walk repository: %v", err)
			}

			fmt.Printf("\nFound %d commits total\n", len(commits))

			// Store commits and files
			totalFiles := 0
			totalBlobs := 0
			for i, commit := range commits {
				if (i+1)%100 == 0 || (i+1) == len(commits) {
					fmt.Printf("Processing commit %d/%d...\r", i+1, len(commits))
				}

				commitID, err := db.CreateCommit(
					runID,
					commit.Hash,
					commit.Author,
					commit.AuthorEmail,
					commit.Committer,
					commit.CommitterEmail,
					commit.Date,
					commit.Message,
					commit.ParentHashes,
				)
				if err != nil {
					db.FinishRun(runID, "failed")
					log.Fatalf("Failed to create commit: %v", err)
				}

				for _, file := range commit.Files {
					var blobID *int64

					// Store file content as blob if available
					if len(file.Content) > 0 {
						id, err := db.GetOrCreateBlob(file.Content, file.SHA256Hash)
						if err != nil {
							db.FinishRun(runID, "failed")
							log.Fatalf("Failed to create blob: %v", err)
						}
						blobID = &id
						totalBlobs++
					}

					err := db.CreateFile(commitID, file.Path, file.Size, file.Mode, blobID)
					if err != nil {
						db.FinishRun(runID, "failed")
						log.Fatalf("Failed to create file: %v", err)
					}
					totalFiles++
				}
			}

			fmt.Printf("\nProcessed %d commits with %d files and %d blobs\n", len(commits), totalFiles, totalBlobs)

			// Update counts and finish run
			if err := db.UpdateRunItemCount(runID); err != nil {
				log.Fatalf("Failed to update run counts: %v", err)
			}

			if err := db.FinishRun(runID, "completed"); err != nil {
				log.Fatalf("Failed to finish run: %v", err)
			}

			fmt.Printf("Ingestion completed successfully!\n")
		},
	}
}

func newFSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fs [path]",
		Short: "Ingest filesystem entries recursively",
		Long:  `Ingest walks through a filesystem path recursively and stores file/directory metadata and contents in the database.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fsPath := args[0]

			// Convert to absolute path
			absPath, err := filepath.Abs(fsPath)
			if err != nil {
				log.Fatalf("Failed to get absolute path: %v", err)
			}

			// Verify path exists
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				log.Fatalf("Path does not exist: %s", absPath)
			}

			fmt.Printf("Ingesting filesystem: %s\n", absPath)

			// Open database
			db, err := database.Open()
			if err != nil {
				log.Fatalf("Failed to open database: %v", err)
			}
			defer db.Close()

			// Create a new run
			runID, err := db.CreateRun(absPath, "fs")
			if err != nil {
				log.Fatalf("Failed to create run: %v", err)
			}

			fmt.Printf("Started ingestion run #%d\n", runID)
			fmt.Println("Walking filesystem...")

			respectGitignore, err := cmd.Flags().GetBool("respect-gitignore")
			if err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to read respect-gitignore flag: %v", err)
			}

			// Walk the filesystem with progress callback
			entries, err := fs.WalkFilesystemWithOptions(absPath, func(count int) {
				fmt.Printf("Found %d entries so far...\r", count)
			}, fs.WalkOptions{RespectGitignore: respectGitignore})
			if err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to walk filesystem: %v", err)
			}

			fmt.Printf("\nFound %d entries total\n", len(entries))

			// Store entries
			totalBlobs := 0
			for i, entry := range entries {
				if (i+1)%100 == 0 || (i+1) == len(entries) {
					fmt.Printf("Processing entry %d/%d...\r", i+1, len(entries))
				}

				var blobID *int64

				// Store file content as blob if available
				if len(entry.Content) > 0 {
					id, err := db.GetOrCreateBlob(entry.Content, entry.SHA256Hash)
					if err != nil {
						db.FinishRun(runID, "failed")
						log.Fatalf("Failed to create blob: %v", err)
					}
					blobID = &id
					totalBlobs++
				}

				err := db.CreateFSEntry(
					runID,
					entry.Path,
					entry.IsDir,
					entry.Size,
					entry.Mode,
					entry.ModTime,
					blobID,
				)
				if err != nil {
					db.FinishRun(runID, "failed")
					log.Fatalf("Failed to create fs entry: %v", err)
				}
			}

			fmt.Printf("\nProcessed %d entries with %d blobs\n", len(entries), totalBlobs)

			// Update counts and finish run
			if err := db.UpdateRunItemCount(runID); err != nil {
				log.Fatalf("Failed to update run counts: %v", err)
			}

			if err := db.FinishRun(runID, "completed"); err != nil {
				log.Fatalf("Failed to finish run: %v", err)
			}

			fmt.Printf("Ingestion completed successfully!\n")
		},
	}

	cmd.Flags().Bool("respect-gitignore", true, "respect .gitignore patterns when walking the filesystem")
	return cmd
}

func newCmdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cmd [command]",
		Short: "Run a shell command and ingest its output",
		Long:  `Run a shell command and store its output (stdout/stderr), exit code, and execution time in the database.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shellCmd := args[0]
			if len(args) > 1 {
				// Join all args as a single command
				shellCmd = ""
				for i, arg := range args {
					if i > 0 {
						shellCmd += " "
					}
					shellCmd += arg
				}
			}

			fmt.Printf("Running command: %s\n", shellCmd)

			// Open database
			db, err := database.Open()
			if err != nil {
				log.Fatalf("Failed to open database: %v", err)
			}
			defer db.Close()

			// Create a new run
			runID, err := db.CreateRun(shellCmd, "cmd")
			if err != nil {
				log.Fatalf("Failed to create run: %v", err)
			}

			fmt.Printf("Started ingestion run #%d\n", runID)

			// Run the command
			result, err := command.RunCommand(shellCmd)
			if err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to run command: %v", err)
			}

			fmt.Printf("Command completed with exit code: %d (took %dms)\n", result.ExitCode, result.DurationMs)
			if len(result.Stdout) > 0 {
				fmt.Printf("Stdout length: %d bytes\n", len(result.Stdout))
			}
			if len(result.Stderr) > 0 {
				fmt.Printf("Stderr length: %d bytes\n", len(result.Stderr))
			}

			// Store command result
			err = db.CreateCmdRun(
				runID,
				result.Command,
				result.ExitCode,
				result.Stdout,
				result.Stderr,
				result.DurationMs,
			)
			if err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to create cmd run: %v", err)
			}

			if err := db.UpdateRunItemCount(runID); err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to update run counts: %v", err)
			}

			// Finish run
			if err := db.FinishRun(runID, "completed"); err != nil {
				log.Fatalf("Failed to finish run: %v", err)
			}

			fmt.Printf("Ingestion completed successfully!\n")
		},
	}
}

func newListRunsCmd() *cobra.Command {
	return &cobra.Command{
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

			fmt.Printf("\n%-5s %-6s %-19s %-50s %-8s %-10s\n", "ID", "Type", "Start Time", "Path/Command", "Status", "Items")
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

				fmt.Printf("%-5d %-6s %-19s %-50s %-8s %-10d%s\n",
					run.ID,
					run.RunType,
					startTime,
					repoPath,
					run.Status,
					run.ItemCount,
					duration,
				)
			}

			fmt.Println()

			// Display summary statistics
			totalItems := 0
			completedRuns := 0
			for _, run := range runs {
				totalItems += run.ItemCount
				if run.Status == "completed" {
					completedRuns++
				}
			}

			fmt.Printf("Summary: %d total runs (%d completed), %d items\n",
				len(runs), completedRuns, totalItems)
		},
	}
}

func newQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query [sql]",
		Short: "Query the database with SQL and output JSON",
		Long:  `Execute a SQL query against the ingest database and output the results as JSON.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sqlQuery := args[0]

			// Open database
			db, err := database.Open()
			if err != nil {
				log.Fatalf("Failed to open database: %v", err)
			}
			defer db.Close()

			// Execute query
			results, err := db.Query(sqlQuery)
			if err != nil {
				log.Fatalf("Failed to execute query: %v", err)
			}

			// Output as JSON
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(results); err != nil {
				log.Fatalf("Failed to encode results as JSON: %v", err)
			}
		},
	}
}

func newGitHubCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "github [owner/repo]",
		Short: "Ingest GitHub issues and pull requests",
		Long:  `Ingest fetches all issues and pull requests from a GitHub repository, including all comments and metadata.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			repoSpec := args[0]

			// Parse owner/repo
			parts := strings.Split(repoSpec, "/")
			if len(parts) != 2 {
				log.Fatalf("Invalid repository specification: %s (expected format: owner/repo)", repoSpec)
			}
			owner := parts[0]
			repo := parts[1]

			fmt.Printf("Ingesting GitHub repository: %s/%s\n", owner, repo)

			// Open database
			db, err := database.Open()
			if err != nil {
				log.Fatalf("Failed to open database: %v", err)
			}
			defer db.Close()

			// Create a new run
			runID, err := db.CreateRun(repoSpec, "github")
			if err != nil {
				log.Fatalf("Failed to create run: %v", err)
			}

			fmt.Printf("Started ingestion run #%d\n", runID)

			// Create GitHub client
			client := github.NewClient()

			// Fetch issues
			fmt.Println("Fetching issues...")
			issues, err := client.FetchIssues(owner, repo, "all", func(count int) {
				fmt.Printf("Found %d issues so far...\r", count)
			})
			if err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to fetch issues: %v", err)
			}
			fmt.Printf("\nFound %d issues total\n", len(issues))

			// Store issues with comments
			for i, issue := range issues {
				if (i+1)%10 == 0 || (i+1) == len(issues) {
					fmt.Printf("Processing issue %d/%d...\r", i+1, len(issues))
				}

				// Convert labels to comma-separated string
				var labelNames []string
				for _, label := range issue.Labels {
					labelNames = append(labelNames, label.Name)
				}
				labels := strings.Join(labelNames, ",")

				// Convert assignees to comma-separated string
				var assigneeNames []string
				for _, assignee := range issue.Assignees {
					assigneeNames = append(assigneeNames, assignee.Login)
				}
				assignees := strings.Join(assigneeNames, ",")

				// Get milestone title
				milestone := ""
				if issue.Milestone != nil {
					milestone = issue.Milestone.Title
				}

				err := db.CreateGitHubIssue(
					runID,
					issue.Number,
					issue.Title,
					issue.Body,
					issue.State,
					issue.User.Login,
					issue.CreatedAt,
					issue.UpdatedAt,
					issue.ClosedAt,
					labels,
					assignees,
					milestone,
				)
				if err != nil {
					db.FinishRun(runID, "failed")
					log.Fatalf("Failed to create issue: %v", err)
				}

				// Fetch and store comments
				comments, err := client.FetchIssueComments(owner, repo, issue.Number)
				if err != nil {
					db.FinishRun(runID, "failed")
					log.Fatalf("Failed to fetch comments for issue #%d: %v", issue.Number, err)
				}

				for _, comment := range comments {
					err := db.CreateGitHubComment(
						runID,
						"issue",
						issue.Number,
						comment.ID,
						comment.User.Login,
						comment.Body,
						comment.CreatedAt,
						comment.UpdatedAt,
					)
					if err != nil {
						db.FinishRun(runID, "failed")
						log.Fatalf("Failed to create comment: %v", err)
					}
				}
			}

			fmt.Printf("\nProcessed %d issues with their comments\n", len(issues))

			// Fetch pull requests
			fmt.Println("Fetching pull requests...")
			prs, err := client.FetchPullRequests(owner, repo, "all", func(count int) {
				fmt.Printf("Found %d pull requests so far...\r", count)
			})
			if err != nil {
				db.FinishRun(runID, "failed")
				log.Fatalf("Failed to fetch pull requests: %v", err)
			}
			fmt.Printf("\nFound %d pull requests total\n", len(prs))

			// Store PRs with comments
			for i, pr := range prs {
				if (i+1)%10 == 0 || (i+1) == len(prs) {
					fmt.Printf("Processing PR %d/%d...\r", i+1, len(prs))
				}

				// Convert labels to comma-separated string
				var labelNames []string
				for _, label := range pr.Labels {
					labelNames = append(labelNames, label.Name)
				}
				labels := strings.Join(labelNames, ",")

				// Convert assignees to comma-separated string
				var assigneeNames []string
				for _, assignee := range pr.Assignees {
					assigneeNames = append(assigneeNames, assignee.Login)
				}
				assignees := strings.Join(assigneeNames, ",")

				// Convert reviewers to comma-separated string
				var reviewerNames []string
				for _, reviewer := range pr.RequestedReviewers {
					reviewerNames = append(reviewerNames, reviewer.Login)
				}
				reviewers := strings.Join(reviewerNames, ",")

				// Get milestone title
				milestone := ""
				if pr.Milestone != nil {
					milestone = pr.Milestone.Title
				}

				err := db.CreateGitHubPR(
					runID,
					pr.Number,
					pr.Title,
					pr.Body,
					pr.State,
					pr.User.Login,
					pr.CreatedAt,
					pr.UpdatedAt,
					pr.ClosedAt,
					pr.MergedAt,
					pr.Merged,
					pr.Draft,
					pr.Base.Ref,
					pr.Head.Ref,
					labels,
					assignees,
					reviewers,
					milestone,
				)
				if err != nil {
					db.FinishRun(runID, "failed")
					log.Fatalf("Failed to create pull request: %v", err)
				}

				// Fetch and store comments (including review comments)
				comments, err := client.FetchPRComments(owner, repo, pr.Number)
				if err != nil {
					db.FinishRun(runID, "failed")
					log.Fatalf("Failed to fetch comments for PR #%d: %v", pr.Number, err)
				}

				for _, comment := range comments {
					err := db.CreateGitHubComment(
						runID,
						"pr",
						pr.Number,
						comment.ID,
						comment.User.Login,
						comment.Body,
						comment.CreatedAt,
						comment.UpdatedAt,
					)
					if err != nil {
						db.FinishRun(runID, "failed")
						log.Fatalf("Failed to create comment: %v", err)
					}
				}
			}

			fmt.Printf("\nProcessed %d pull requests with their comments\n", len(prs))

			// Update counts and finish run
			if err := db.UpdateRunItemCount(runID); err != nil {
				log.Fatalf("Failed to update run counts: %v", err)
			}

			if err := db.FinishRun(runID, "completed"); err != nil {
				log.Fatalf("Failed to finish run: %v", err)
			}

			fmt.Printf("Ingestion completed successfully!\n")
		},
	}
}

func newMCPCmd() *cobra.Command {
	var provider string
	var endpoint string
	var token string
	var headers []string
	timeout := 30 * time.Second
	maxAttempts := 3
	initialBackoff := 500 * time.Millisecond
	maxBackoff := 5 * time.Second

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Interact with Model Context Protocol servers",
		Long: `Connect to an MCP server over SSE. Flags override environment variables.

Environment variables:
  INGEST_MCP_ENDPOINT, INGEST_<PROVIDER>_MCP_ENDPOINT
  INGEST_MCP_TOKEN,    INGEST_<PROVIDER>_MCP_TOKEN
  INGEST_MCP_HEADERS,  INGEST_<PROVIDER>_MCP_HEADERS (comma-separated key=value pairs)
  INGEST_MCP_TIMEOUT, INGEST_MCP_RETRY_MAX_ATTEMPTS, INGEST_MCP_RETRY_INITIAL_BACKOFF, INGEST_MCP_RETRY_MAX_BACKOFF`,
	}

	cmd.PersistentFlags().StringVar(&provider, "provider", "", "provider preset name (e.g. linear, github)")
	cmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "MCP SSE endpoint URL")
	cmd.PersistentFlags().StringVar(&token, "token", "", "Bearer token for MCP server")
	cmd.PersistentFlags().StringSliceVar(&headers, "header", nil, "Additional HTTP header key=value (repeatable)")
	cmd.PersistentFlags().DurationVar(&timeout, "timeout", timeout, "HTTP timeout (supports Go duration syntax)")
	cmd.PersistentFlags().IntVar(&maxAttempts, "retry-max-attempts", maxAttempts, "Maximum connection attempts before failing")
	cmd.PersistentFlags().DurationVar(&initialBackoff, "retry-initial-backoff", initialBackoff, "Initial retry backoff duration")
	cmd.PersistentFlags().DurationVar(&maxBackoff, "retry-max-backoff", maxBackoff, "Maximum retry backoff duration")

	cmd.AddCommand(newMCPListToolsCmd(func() (mcppkg.Config, error) {
		headerMap, err := parseHeaderPairs(headers)
		if err != nil {
			return mcppkg.Config{}, err
		}

		cfgOverrides := mcppkg.Config{
			Endpoint:  endpoint,
			AuthToken: token,
			Headers:   headerMap,
			Timeout:   timeout,
			Retry: mcppkg.RetryConfig{
				MaxAttempts:    maxAttempts,
				InitialBackoff: initialBackoff,
				MaxBackoff:     maxBackoff,
			},
		}

		return mcppkg.ResolveConfig(provider, cfgOverrides)
	}))

	return cmd
}

func newMCPListToolsCmd(resolveConfig func() (mcppkg.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list-tools",
		Short: "List tools exposed by the MCP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := resolveConfig()
			if err != nil {
				return err
			}

			client, err := mcppkg.NewClient(cfg)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), cfg.Timeout)
			defer cancel()

			session, err := client.Connect(ctx)
			if err != nil {
				return err
			}
			defer session.Close()

			tools, err := session.ListTools(ctx)
			if err != nil {
				return err
			}

			if len(tools) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No tools reported by the MCP server.")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Available tools:")
			for _, tool := range tools {
				desc := strings.TrimSpace(tool.Description)
				if desc == "" {
					fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", tool.Name)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "- %s: %s\n", tool.Name, desc)
				}
			}
			return nil
		},
	}
}

func parseHeaderPairs(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	headers := make(map[string]string)
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header %q (expected key=value)", pair)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("invalid header %q (empty key)", pair)
		}
		headers[key] = value
	}
	return headers, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
