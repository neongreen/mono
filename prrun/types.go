package main

// PRInfo contains parsed PR information
type PRInfo struct {
	Owner   string
	Repo    string
	PRNum   int
	Project string
}

// GitHubWorkflowRun represents a GitHub Actions workflow run
type GitHubWorkflowRun struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// GitHubWorkflowRunsResponse represents the response from the workflow runs API
type GitHubWorkflowRunsResponse struct {
	WorkflowRuns []GitHubWorkflowRun `json:"workflow_runs"`
}
