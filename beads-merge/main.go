package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

// Issue represents a beads issue with all possible fields
type Issue struct {
	ID           string       `json:"id"`
	Title        string       `json:"title,omitempty"`
	Description  string       `json:"description,omitempty"`
	Notes        string       `json:"notes,omitempty"`
	Status       string       `json:"status,omitempty"`
	Priority     int          `json:"priority,omitempty"`
	IssueType    string       `json:"issue_type,omitempty"`
	CreatedAt    string       `json:"created_at,omitempty"`
	UpdatedAt    string       `json:"updated_at,omitempty"`
	ClosedAt     string       `json:"closed_at,omitempty"`
	CreatedBy    string       `json:"created_by,omitempty"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
	RawLine      string       `json:"-"` // Store original line for conflict output
}

// Dependency represents an issue dependency
type Dependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
}

// IssueKey uniquely identifies an issue for matching
type IssueKey struct {
	ID        string
	CreatedAt string
	CreatedBy string
}

var rootCmd = &cobra.Command{
	Use:   "beads-merge <output> <base> <left> <right>",
	Short: "3-way merge tool for beads .jsonl issue files",
	Long: `beads-merge is a 3-way merge tool for beads issue tracker .jsonl files.

It intelligently merges issues based on identity (id + created_at + created_by),
applies field-specific merge rules, combines dependencies, and outputs conflict
markers for unresolvable conflicts.

Designed to work with jj (Jujutsu) version control as a merge driver.`,
	Args: cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		outputPath := args[0]
		basePath := args[1]
		leftPath := args[2]
		rightPath := args[3]

		// Read all three files
		baseIssues, err := readIssues(basePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading base file: %v\n", err)
			os.Exit(1)
		}

		leftIssues, err := readIssues(leftPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading left file: %v\n", err)
			os.Exit(1)
		}

		rightIssues, err := readIssues(rightPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading right file: %v\n", err)
			os.Exit(1)
		}

		// Perform 3-way merge
		result, conflicts := merge3Way(baseIssues, leftIssues, rightIssues)

		// Open output file for writing
		outFile, err := os.Create(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer outFile.Close()

		// Write merged result to output file
		for _, issue := range result {
			line, err := json.Marshal(issue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling issue %s: %v\n", issue.ID, err)
				os.Exit(1)
			}
			fmt.Fprintln(outFile, string(line))
		}

		// Write conflicts to output file
		for _, conflict := range conflicts {
			fmt.Fprintln(outFile, conflict)
		}

		// Exit with 1 if there were conflicts (jj will interpret this as conflict markers present)
		if len(conflicts) > 0 {
			os.Exit(1)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func readIssues(path string) ([]Issue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var issues []Issue
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var issue Issue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}
		issue.RawLine = line
		issues = append(issues, issue)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return issues, nil
}

func makeKey(issue Issue) IssueKey {
	return IssueKey{
		ID:        issue.ID,
		CreatedAt: issue.CreatedAt,
		CreatedBy: issue.CreatedBy,
	}
}

func merge3Way(base, left, right []Issue) ([]Issue, []string) {
	// Build maps for quick lookup
	baseMap := make(map[IssueKey]Issue)
	for _, issue := range base {
		baseMap[makeKey(issue)] = issue
	}

	leftMap := make(map[IssueKey]Issue)
	for _, issue := range left {
		leftMap[makeKey(issue)] = issue
	}

	rightMap := make(map[IssueKey]Issue)
	for _, issue := range right {
		rightMap[makeKey(issue)] = issue
	}

	// Track which issues we've processed
	processed := make(map[IssueKey]bool)
	var result []Issue
	var conflicts []string

	// Process all unique keys
	allKeys := make(map[IssueKey]bool)
	for k := range baseMap {
		allKeys[k] = true
	}
	for k := range leftMap {
		allKeys[k] = true
	}
	for k := range rightMap {
		allKeys[k] = true
	}

	for key := range allKeys {
		if processed[key] {
			continue
		}
		processed[key] = true

		baseIssue, inBase := baseMap[key]
		leftIssue, inLeft := leftMap[key]
		rightIssue, inRight := rightMap[key]

		// Handle different scenarios
		if inBase && inLeft && inRight {
			// All three present - merge
			merged, conflict := mergeIssue(baseIssue, leftIssue, rightIssue)
			if conflict != "" {
				conflicts = append(conflicts, conflict)
			} else {
				result = append(result, merged)
			}
		} else if !inBase && inLeft && inRight {
			// Added in both - check if identical
			if issuesEqual(leftIssue, rightIssue) {
				result = append(result, leftIssue)
			} else {
				conflicts = append(conflicts, makeConflict(leftIssue.RawLine, rightIssue.RawLine))
			}
		} else if inBase && inLeft && !inRight {
			// Deleted in right, maybe modified in left
			if issuesEqual(baseIssue, leftIssue) {
				// Deleted in right, unchanged in left - accept deletion
				continue
			} else {
				// Modified in left, deleted in right - conflict
				conflicts = append(conflicts, makeConflict(leftIssue.RawLine, ""))
			}
		} else if inBase && !inLeft && inRight {
			// Deleted in left, maybe modified in right
			if issuesEqual(baseIssue, rightIssue) {
				// Deleted in left, unchanged in right - accept deletion
				continue
			} else {
				// Modified in right, deleted in left - conflict
				conflicts = append(conflicts, makeConflict("", rightIssue.RawLine))
			}
		} else if !inBase && inLeft && !inRight {
			// Added only in left
			result = append(result, leftIssue)
		} else if !inBase && !inLeft && inRight {
			// Added only in right
			result = append(result, rightIssue)
		}
	}

	return result, conflicts
}

func mergeIssue(base, left, right Issue) (Issue, string) {
	result := Issue{
		ID:        base.ID,
		CreatedAt: base.CreatedAt,
		CreatedBy: base.CreatedBy,
	}

	// Merge title
	result.Title = mergeField(base.Title, left.Title, right.Title)

	// Merge description
	result.Description = mergeField(base.Description, left.Description, right.Description)

	// Merge notes
	result.Notes = mergeField(base.Notes, left.Notes, right.Notes)

	// Merge status
	result.Status = mergeField(base.Status, left.Status, right.Status)

	// Merge priority (as int)
	if base.Priority == left.Priority && base.Priority != right.Priority {
		result.Priority = right.Priority
	} else if base.Priority == right.Priority && base.Priority != left.Priority {
		result.Priority = left.Priority
	} else if left.Priority == right.Priority {
		result.Priority = left.Priority
	} else {
		// Conflict - take left for now
		result.Priority = left.Priority
	}

	// Merge issue_type
	result.IssueType = mergeField(base.IssueType, left.IssueType, right.IssueType)

	// Merge updated_at - take the max
	result.UpdatedAt = maxTime(left.UpdatedAt, right.UpdatedAt)

	// Merge closed_at - take the max
	result.ClosedAt = maxTime(left.ClosedAt, right.ClosedAt)

	// Merge dependencies - combine and deduplicate
	result.Dependencies = mergeDependencies(left.Dependencies, right.Dependencies)

	// Check if we have a real conflict
	if hasConflict(base, left, right, result) {
		return result, makeConflict(left.RawLine, right.RawLine)
	}

	return result, ""
}

func mergeField(base, left, right string) string {
	if base == left && base != right {
		return right
	}
	if base == right && base != left {
		return left
	}
	// Both changed to same value or no change
	return left
}

func maxTime(t1, t2 string) string {
	if t1 == "" && t2 == "" {
		return ""
	}
	if t1 == "" {
		return t2
	}
	if t2 == "" {
		return t1
	}

	// Try RFC3339Nano first (supports fractional seconds), fall back to RFC3339
	time1, err1 := time.Parse(time.RFC3339Nano, t1)
	if err1 != nil {
		time1, err1 = time.Parse(time.RFC3339, t1)
	}

	time2, err2 := time.Parse(time.RFC3339Nano, t2)
	if err2 != nil {
		time2, err2 = time.Parse(time.RFC3339, t2)
	}

	// If both fail to parse, return t2 as fallback
	if err1 != nil && err2 != nil {
		return t2
	}
	// If only t1 failed to parse, return t2
	if err1 != nil {
		return t2
	}
	// If only t2 failed to parse, return t1
	if err2 != nil {
		return t1
	}

	if time1.After(time2) {
		return t1
	}
	return t2
}

func mergeDependencies(left, right []Dependency) []Dependency {
	seen := make(map[string]bool)
	var result []Dependency

	for _, dep := range left {
		key := fmt.Sprintf("%s:%s:%s", dep.IssueID, dep.DependsOnID, dep.Type)
		if !seen[key] {
			seen[key] = true
			result = append(result, dep)
		}
	}

	for _, dep := range right {
		key := fmt.Sprintf("%s:%s:%s", dep.IssueID, dep.DependsOnID, dep.Type)
		if !seen[key] {
			seen[key] = true
			result = append(result, dep)
		}
	}

	return result
}

func hasConflict(base, left, right, merged Issue) bool {
	// Check if any field has conflicting changes
	if base.Title != left.Title && base.Title != right.Title && left.Title != right.Title {
		return true
	}
	if base.Description != left.Description && base.Description != right.Description && left.Description != right.Description {
		return true
	}
	if base.Notes != left.Notes && base.Notes != right.Notes && left.Notes != right.Notes {
		return true
	}
	if base.Status != left.Status && base.Status != right.Status && left.Status != right.Status {
		return true
	}
	if base.Priority != left.Priority && base.Priority != right.Priority && left.Priority != right.Priority {
		return true
	}
	if base.IssueType != left.IssueType && base.IssueType != right.IssueType && left.IssueType != right.IssueType {
		return true
	}
	return false
}

func issuesEqual(a, b Issue) bool {
	// Use go-cmp for deep equality comparison, ignoring RawLine field
	return cmp.Equal(a, b, cmp.FilterPath(func(p cmp.Path) bool {
		return p.String() == "RawLine"
	}, cmp.Ignore()))
}

func makeConflict(left, right string) string {
	conflict := "<<<<<<< left\n"
	if left != "" {
		conflict += left + "\n"
	}
	conflict += "=======\n"
	if right != "" {
		conflict += right + "\n"
	}
	conflict += ">>>>>>> right"
	return conflict
}
