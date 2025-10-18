package jobs

// Result captures the outcome of a single ingestion run.
type Result struct {
	RunID     int64
	ItemCount int
	Details   map[string]int
}

// GitOptions configures a git ingestion job.
type GitOptions struct {
	Path             string
	RespectGitignore bool
}

// FSOptions configures a filesystem ingestion job.
type FSOptions struct {
	Path             string
	RespectGitignore bool
}

// CommandOptions configures a shell command ingestion job.
type CommandOptions struct {
	Command string
}

// GitHubOptions configures the GitHub MCP ingestion job.
type GitHubOptions struct {
	Owner string
	Repo  string
}
