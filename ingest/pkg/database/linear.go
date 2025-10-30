package database

import "fmt"

// CreateLinearIssue stores a Linear issue for the given run.
func (d *Database) CreateLinearIssue(runID int64, issue LinearIssue) error {
	_, err := d.db.Exec(
		`INSERT INTO linear_issues (
			run_id,
			issue_id,
			identifier,
			title,
			description,
			priority,
			status,
			assignee,
			team,
			url,
			raw_data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		issue.IssueID,
		issue.Identifier,
		issue.Title,
		nullableString(issue.Description),
		nullableInt(issue.Priority),
		nullableString(issue.Status),
		nullableString(issue.Assignee),
		nullableString(issue.Team),
		nullableString(issue.URL),
		nullableString(issue.RawData),
	)
	if err != nil {
		return fmt.Errorf("failed to create linear issue %s: %w", issue.Identifier, err)
	}
	return nil
}
