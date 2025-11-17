package beads

// BeadsDependency represents a dependency relationship in beads format
type BeadsDependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
}

// BeadsIssue represents an issue from beads JSONL format
type BeadsIssue struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       string            `json:"status"` // open, in_progress, closed
	Priority     int               `json:"priority"`
	Type         string            `json:"type"` // bug, feature, task, epic, chore
	Labels       []string          `json:"labels"`
	Assignee     string            `json:"assignee"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
	Dependencies []BeadsDependency `json:"dependencies"`
}

// ImportOptions contains options for importing beads issues
type ImportOptions struct {
	BeadsPath   string
	AliasPrefix string
	DryRun      bool
}

// ImportResult contains the results of an import operation
type ImportResult struct {
	TotalImported       int
	TotalSkipped        int
	RelationsImported   int
	RenumberedIssues    []string
	ProjectsCreated     map[string]string // prefix -> project UID
	FailedNotes         []string          // Failed renumber notes
	FailedRelationships []string          // Failed relationship imports
}
