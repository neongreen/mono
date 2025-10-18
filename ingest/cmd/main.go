package main

import (
	"context"
	"encoding/json"
	"fmt"
	"ingest/pkg/database"
	"ingest/pkg/jobs"
	mcppkg "ingest/pkg/mcp"
	"ingest/pkg/runconfig"
	"log"
	"os"
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
	cmd.AddCommand(newGitHubMCPCmd())
	cmd.AddCommand(newListRunsCmd())
	cmd.AddCommand(newQueryCmd())
	cmd.AddCommand(newLinearCmd())
	cmd.AddCommand(newMCPCmd())
	cmd.AddCommand(newRunConfigCmd())
	cmd.AddCommand(newConfigCmd())

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
			_, err := jobs.RunGit(cmd.Context(), cmd.OutOrStdout(), jobs.GitOptions{Path: repoPath})
			if err != nil {
				log.Fatalf("Git ingestion failed: %v", err)
			}
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
			respectGitignore, err := cmd.Flags().GetBool("respect-gitignore")
			if err != nil {
				log.Fatalf("Failed to read respect-gitignore flag: %v", err)
			}
			_, err = jobs.RunFS(cmd.Context(), cmd.OutOrStdout(), jobs.FSOptions{Path: fsPath, RespectGitignore: respectGitignore})
			if err != nil {
				log.Fatalf("Filesystem ingestion failed: %v", err)
			}
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
			shellCmd := strings.Join(args, " ")
			if _, err := jobs.RunCommand(cmd.Context(), cmd.OutOrStdout(), jobs.CommandOptions{Command: shellCmd}); err != nil {
				log.Fatalf("Command ingestion failed: %v", err)
			}
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

			if _, err := jobs.RunGitHub(cmd.Context(), cmd.OutOrStdout(), jobs.GitHubOptions{
				Owner: owner,
				Repo:  repo,
			}); err != nil {
				log.Fatalf("GitHub ingestion failed: %v", err)
			}
		},
	}
}

func newGitHubMCPCmd() *cobra.Command {
	var endpoint string
	var token string
	var headers []string
	timeout := 30 * time.Second
	maxAttempts := 3
	initialBackoff := 500 * time.Millisecond
	maxBackoff := 5 * time.Second

	cmd := &cobra.Command{
		Use:   "github-mcp [owner/repo]",
		Short: "Ingest a GitHub repository via the MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoSpec := args[0]
			parts := strings.Split(repoSpec, "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid repository specification %q (expected owner/repo)", repoSpec)
			}
			owner, repo := parts[0], parts[1]

			headerMap, err := parseHeaderPairs(headers)
			if err != nil {
				return err
			}

			overrides := mcppkg.Config{
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
			cfg, err := mcppkg.ResolveConfig("github", overrides)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Using GitHub MCP endpoint %s\n", cfg.Endpoint)

			result, err := jobs.RunGitHubMCP(cmd.Context(), cmd.OutOrStdout(), cfg, jobs.GitHubOptions{
				Owner: owner,
				Repo:  repo,
			})
			if err != nil {
				return fmt.Errorf("failed to ingest GitHub repository: %w", err)
			}

			issues := result.Details["issues"]
			issueComments := result.Details["issueComments"]
			prs := result.Details["pullRequests"]
			prComments := result.Details["prComments"]

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"Ingested %d issues (%d comments) and %d pull requests (%d comments)\n",
				issues,
				issueComments,
				prs,
				prComments,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&endpoint, "endpoint", "", "MCP SSE endpoint URL")
	cmd.Flags().StringVar(&token, "token", "", "Bearer token for MCP server")
	cmd.Flags().StringSliceVar(&headers, "header", nil, "Additional HTTP header key=value (repeatable)")
	cmd.Flags().DurationVar(&timeout, "timeout", timeout, "HTTP timeout (supports Go duration syntax)")
	cmd.Flags().IntVar(&maxAttempts, "retry-max-attempts", maxAttempts, "Maximum connection attempts before failing")
	cmd.Flags().DurationVar(&initialBackoff, "retry-initial-backoff", initialBackoff, "Initial retry backoff duration")
	cmd.Flags().DurationVar(&maxBackoff, "retry-max-backoff", maxBackoff, "Maximum retry backoff duration")

	return cmd
}

func newLinearCmd() *cobra.Command {
	var endpoint string
	var token string
	var headers []string
	timeout := 30 * time.Second
	maxAttempts := 3
	initialBackoff := 500 * time.Millisecond
	maxBackoff := 5 * time.Second

	cmd := &cobra.Command{
		Use:   "linear",
		Short: "Ingest Linear issues via MCP",
		Long:  `Connects to the Linear MCP server over SSE and stores issues in the ingest database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			headerMap, err := parseHeaderPairs(headers)
			if err != nil {
				return err
			}

			overrides := mcppkg.Config{
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
			cfg, err := mcppkg.ResolveConfig("linear", overrides)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Using Linear MCP endpoint %s\n", cfg.Endpoint)
			result, err := jobs.RunLinearMCP(cmd.Context(), cmd.OutOrStdout(), cfg)
			if err != nil {
				return fmt.Errorf("failed to ingest Linear issues: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Ingested %d Linear issues\n", result.ItemCount)
			return nil
		},
	}

	cmd.Flags().StringVar(&endpoint, "endpoint", "", "MCP SSE endpoint URL")
	cmd.Flags().StringVar(&token, "token", "", "Bearer token for MCP server")
	cmd.Flags().StringSliceVar(&headers, "header", nil, "Additional HTTP header key=value (repeatable)")
	cmd.Flags().DurationVar(&timeout, "timeout", timeout, "HTTP timeout (supports Go duration syntax)")
	cmd.Flags().IntVar(&maxAttempts, "retry-max-attempts", maxAttempts, "Maximum connection attempts before failing")
	cmd.Flags().DurationVar(&initialBackoff, "retry-initial-backoff", initialBackoff, "Initial retry backoff duration")
	cmd.Flags().DurationVar(&maxBackoff, "retry-max-backoff", maxBackoff, "Maximum retry backoff duration")

	return cmd
}

func newRunConfigCmd() *cobra.Command {
	var configPath string
	var parallelism int

	cmd := &cobra.Command{
		Use:   "run-config",
		Short: "Run jobs defined in ingest.config.toml",
		Long:  "Execute ingestion jobs described in a TOML configuration file. Jobs run in parallel and execution continues even if some jobs fail.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runconfig.LoadFile(configPath)
			if err != nil {
				return err
			}
			if parallelism > 0 {
				cfg.Parallelism = parallelism
			}

			results, execErr := runconfig.Execute(cmd.Context(), cmd.OutOrStdout(), cfg)

			fmt.Fprintln(cmd.OutOrStdout(), "\nRun summary:")
			successes := 0
			failures := 0
			for _, res := range results {
				status := "ok"
				if res.Err != nil {
					status = fmt.Sprintf("error: %v", res.Err)
					failures++
				} else {
					successes++
				}
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"- %s (%s) -> %s in %s\n",
					res.Job.DisplayName(),
					res.Job.Type,
					status,
					res.Duration.Round(10*time.Millisecond),
				)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nCompleted %d jobs (%d ok, %d failed)\n", len(results), successes, failures)

			return execErr
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "ingest.config.toml", "Path to run configuration file")
	cmd.Flags().IntVar(&parallelism, "parallelism", 0, "Maximum number of jobs to run concurrently (0 = auto)")

	return cmd
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect ingest configuration files",
	}

	cmd.AddCommand(newConfigValidateCmd())
	return cmd
}

func newConfigValidateCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate an ingest TOML configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runconfig.LoadFile(configPath)
			if err != nil {
				return err
			}

			warnings := runconfig.ConfigWarnings(cfg)

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"Configuration %s is valid (%d job(s)).\n",
				configPath,
				len(cfg.Jobs),
			)

			for idx, job := range cfg.Jobs {
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"- job %d: %s (%s)\n",
					idx+1,
					job.DisplayName(),
					job.Type,
				)
			}

			if len(warnings) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "\nWarnings:")
				for _, warning := range warnings {
					fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", warning)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "ingest.config.toml", "Path to run configuration file")

	return cmd
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

	helpText := fmt.Sprintf(`Connect to an MCP server over SSE. Flags override environment variables.

Environment variables:
  INGEST_MCP_ENDPOINT, INGEST_<PROVIDER>_MCP_ENDPOINT
  INGEST_MCP_TOKEN,    INGEST_<PROVIDER>_MCP_TOKEN
  INGEST_MCP_HEADERS,  INGEST_<PROVIDER>_MCP_HEADERS (comma-separated key=value pairs)
  INGEST_MCP_TIMEOUT, INGEST_MCP_RETRY_MAX_ATTEMPTS, INGEST_MCP_RETRY_INITIAL_BACKOFF, INGEST_MCP_RETRY_MAX_BACKOFF

Built-in providers:
%s`, mcppkg.ProviderHelp())

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Interact with Model Context Protocol servers",
		Long:  helpText,
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
