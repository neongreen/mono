package runconfig

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pelletier/go-toml/v2"

	"ingest/pkg/jobs"
	mcppkg "ingest/pkg/mcp"
)

// Config represents the contents of ingest.config.toml.
type Config struct {
	Parallelism int         `toml:"parallelism"`
	Jobs        []JobConfig `toml:"job"`
}

// JobConfig describes a single ingestion job.
type JobConfig struct {
	Name             string        `toml:"name"`
	Type             string        `toml:"type"`
	Path             string        `toml:"path"`
	RespectGitignore *bool         `toml:"respect_gitignore"`
	Command          string        `toml:"command"`
	Owner            string        `toml:"owner"`
	Repo             string        `toml:"repo"`
	MCP              *MCPOverrides `toml:"mcp"`
}

// MCPOverrides configures MCP-backed jobs.
type MCPOverrides struct {
	Provider string            `toml:"provider"`
	Endpoint string            `toml:"endpoint"`
	Token    string            `toml:"token"`
	Headers  map[string]string `toml:"headers"`
	Timeout  Duration          `toml:"timeout"`
	Retry    RetryOverrides    `toml:"retry"`
}

// RetryOverrides customises retry behaviour for MCP jobs.
type RetryOverrides struct {
	MaxAttempts    int      `toml:"max_attempts"`
	InitialBackoff Duration `toml:"initial_backoff"`
	MaxBackoff     Duration `toml:"max_backoff"`
}

// Duration wraps time.Duration to support TOML marshaling.
type Duration struct {
	time.Duration
}

// UnmarshalText parses Go duration strings.
func (d *Duration) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		d.Duration = 0
		return nil
	}
	value, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", text, err)
	}
	d.Duration = value
	return nil
}

// LoadFile loads a configuration from disk.
func LoadFile(path string) (Config, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config %s: %w", path, err)
	}
	return Parse(data)
}

// Parse decodes configuration data.
func Parse(data []byte) (Config, error) {
	var cfg Config
	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse configuration: %w", err)
	}
	if len(cfg.Jobs) == 0 {
		return Config{}, errors.New("run-config: no jobs defined")
	}
	for i := range cfg.Jobs {
		if err := cfg.Jobs[i].validate(); err != nil {
			return Config{}, fmt.Errorf("job %d: %w", i+1, err)
		}
	}
	return cfg, nil
}

func (job *JobConfig) validate() error {
	job.Type = strings.ToLower(strings.TrimSpace(job.Type))
	if job.Type == "" {
		return errors.New("type is required")
	}

	switch job.Type {
	case "git", "fs":
		if job.Path == "" {
			return fmt.Errorf("%s job requires 'path'", job.Type)
		}
	case "command":
		if strings.TrimSpace(job.Command) == "" {
			return errors.New("command job requires 'command'")
		}
	case "github", "github_mcp":
		if job.Owner == "" || job.Repo == "" {
			return errors.New("github job requires 'owner' and 'repo'")
		}
	case "linear_mcp":
		// no extra validation here
	default:
		return fmt.Errorf("unknown job type %q", job.Type)
	}
	return nil
}

// ExecutionResult captures the outcome of a job execution.
type ExecutionResult struct {
	Job      JobConfig
	Result   jobs.Result
	Err      error
	Duration time.Duration
}

// Execute runs all jobs with optional parallelism. The returned error aggregates
// individual job failures while ensuring every job runs.
func Execute(ctx context.Context, out io.Writer, cfg Config) ([]ExecutionResult, error) {
	count := len(cfg.Jobs)
	if count == 0 {
		return nil, errors.New("run-config: no jobs to execute")
	}

	if err := preflightJobs(cfg); err != nil {
		return nil, err
	}

	parallelism := cfg.Parallelism
	if parallelism <= 0 || parallelism > count {
		parallelism = count
	}

	results := make([]ExecutionResult, count)
	var wg sync.WaitGroup
	sem := make(chan struct{}, parallelism)

	for i := range cfg.Jobs {
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			start := time.Now()
			res, err := runJob(ctx, out, cfg.Jobs[index])
			results[index] = ExecutionResult{
				Job:      cfg.Jobs[index],
				Result:   res,
				Err:      err,
				Duration: time.Since(start),
			}
		}()
	}

	wg.Wait()

	var errs []error
	for _, res := range results {
		if res.Err != nil {
			errs = append(errs, fmt.Errorf("job %q failed: %w", jobDisplayName(res.Job), res.Err))
		}
	}

	return results, errors.Join(errs...)
}

func runJob(ctx context.Context, out io.Writer, job JobConfig) (jobs.Result, error) {
	switch job.Type {
	case "git":
		return jobs.RunGit(ctx, out, jobs.GitOptions{
			Path: job.Path,
		})
	case "fs":
		respect := true
		if job.RespectGitignore != nil {
			respect = *job.RespectGitignore
		}
		return jobs.RunFS(ctx, out, jobs.FSOptions{
			Path:             job.Path,
			RespectGitignore: respect,
		})
	case "command":
		return jobs.RunCommand(ctx, out, jobs.CommandOptions{
			Command: job.Command,
		})
	case "github":
		return jobs.RunGitHub(ctx, out, jobs.GitHubOptions{
			Owner: job.Owner,
			Repo:  job.Repo,
		})
	case "github_mcp":
		mcpCfg, err := resolveMCPConfig(job, "github")
		if err != nil {
			return jobs.Result{}, err
		}
		return jobs.RunGitHubMCP(ctx, out, mcpCfg, jobs.GitHubOptions{
			Owner: job.Owner,
			Repo:  job.Repo,
		})
	case "linear_mcp":
		mcpCfg, err := resolveMCPConfig(job, "linear")
		if err != nil {
			return jobs.Result{}, err
		}
		return jobs.RunLinearMCP(ctx, out, mcpCfg)
	default:
		return jobs.Result{}, fmt.Errorf("unsupported job type %q", job.Type)
	}
}

func resolveMCPConfig(job JobConfig, defaultProvider string) (mcppkg.Config, error) {
	overrides := mcppkg.Config{}
	provider := defaultProvider

	if job.MCP != nil {
		if job.MCP.Provider != "" {
			provider = job.MCP.Provider
		}
		overrides.Endpoint = job.MCP.Endpoint
		overrides.AuthToken = job.MCP.Token
		if len(job.MCP.Headers) > 0 {
			overrides.Headers = maps.Clone(job.MCP.Headers)
		}
		if job.MCP.Timeout.Duration > 0 {
			overrides.Timeout = job.MCP.Timeout.Duration
		}
		overrides.Retry = mcppkg.RetryConfig{
			MaxAttempts:    job.MCP.Retry.MaxAttempts,
			InitialBackoff: job.MCP.Retry.InitialBackoff.Duration,
			MaxBackoff:     job.MCP.Retry.MaxBackoff.Duration,
		}
	}

	return mcppkg.ResolveConfig(provider, overrides)
}

func preflightJobs(cfg Config) error {
	for idx, job := range cfg.Jobs {
		switch job.Type {
		case "github_mcp", "linear_mcp":
			resolved, err := resolveMCPConfig(job, jobTypeDefaultProvider(job.Type))
			if err != nil {
				return fmt.Errorf("job %d (%s): %w", idx+1, jobDisplayName(job), err)
			}
			if err := ensureMCPAuth(job, resolved, idx); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureMCPAuth(job JobConfig, cfg mcppkg.Config, idx int) error {
	if cfg.AuthToken != "" {
		return nil
	}
	return fmt.Errorf("job %d (%s): no MCP token resolved; set job.mcp.token or %s", idx+1, jobDisplayName(job), tokenEnvHint(providerForHints(job)))
}

func providerForHints(job JobConfig) string {
	if job.MCP != nil && strings.TrimSpace(job.MCP.Provider) != "" {
		return strings.TrimSpace(job.MCP.Provider)
	}
	return jobTypeDefaultProvider(job.Type)
}

func tokenEnvHint(provider string) string {
	providerEnv := strings.ToUpper(strings.TrimSpace(provider))
	var hints []string
	if providerEnv != "" {
		hints = append(hints, fmt.Sprintf("INGEST_%s_MCP_TOKEN", providerEnv))
	}
	hints = append(hints, "INGEST_MCP_TOKEN")
	if providerEnv == "GITHUB" {
		hints = append(hints, "MISE_GITHUB_TOKEN", "GITHUB_TOKEN")
	}
	return strings.Join(hints, " or ")
}

func endpointEnvHint(provider string) string {
	providerEnv := strings.ToUpper(strings.TrimSpace(provider))
	var hints []string
	if providerEnv != "" {
		hints = append(hints, fmt.Sprintf("INGEST_%s_MCP_ENDPOINT", providerEnv))
	}
	hints = append(hints, "INGEST_MCP_ENDPOINT")
	return strings.Join(hints, " or ")
}

// ConfigWarnings returns advisory messages for potential MCP misconfiguration.
func ConfigWarnings(cfg Config) []string {
	var warnings []string
	for idx, job := range cfg.Jobs {
		switch job.Type {
		case "github_mcp", "linear_mcp":
			resolved, err := resolveMCPConfig(job, jobTypeDefaultProvider(job.Type))
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("job %d (%s): %v (set job.mcp.endpoint or %s)", idx+1, jobDisplayName(job), err, endpointEnvHint(providerForHints(job))))
				continue
			}
			if resolved.AuthToken == "" {
				warnings = append(warnings, fmt.Sprintf("job %d (%s): no MCP token resolved; set job.mcp.token or %s", idx+1, jobDisplayName(job), tokenEnvHint(providerForHints(job))))
			}
		}
	}
	return warnings
}

func jobTypeDefaultProvider(jobType string) string {
	switch jobType {
	case "github_mcp":
		return "github"
	case "linear_mcp":
		return "linear"
	default:
		return ""
	}
}

// DisplayName returns a human-friendly identifier for the job.
func (job JobConfig) DisplayName() string {
	return jobDisplayName(job)
}

func jobDisplayName(job JobConfig) string {
	if strings.TrimSpace(job.Name) != "" {
		return job.Name
	}
	switch job.Type {
	case "git", "fs":
		if job.Path != "" {
			return fmt.Sprintf("%s:%s", job.Type, job.Path)
		}
	case "github", "github_mcp":
		if job.Owner != "" && job.Repo != "" {
			return fmt.Sprintf("%s:%s/%s", job.Type, job.Owner, job.Repo)
		}
	}
	return job.Type
}
