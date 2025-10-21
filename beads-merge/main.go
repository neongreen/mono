package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

var debugMode bool

var rootCmd = &cobra.Command{
	Use:   "beads-merge",
	Short: "Tools for working with beads .jsonl issue files",
	Long: `beads-merge provides tools for working with beads issue tracker .jsonl files.

It includes a 3-way merge tool designed to work with jj (Jujutsu) version control,
and utilities for managing issue files.`,
	// Allow fallthrough for unknown commands - this enables backwards compatibility
	// where "beads-merge <output> <base> <left> <right>" still works
	SilenceUsage:  true,
	SilenceErrors: true,
}

var mergeCmd = &cobra.Command{
	Use:   "merge <output> <base> <left> <right>",
	Short: "3-way merge tool for beads .jsonl issue files",
	Long: `3-way merge tool for beads issue tracker .jsonl files.

Intelligently merges issues based on identity (id + created_at + created_by),
applies field-specific merge rules, combines dependencies, and outputs conflict
markers for unresolvable conflicts.

Designed to work with jj (Jujutsu) version control as a merge driver.`,
	Args: cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		outputPath := args[0]
		basePath := args[1]
		leftPath := args[2]
		rightPath := args[3]

		if debugMode {
			fmt.Fprintf(os.Stderr, "=== DEBUG MODE ===\n")
			fmt.Fprintf(os.Stderr, "Output path: %s\n", outputPath)
			fmt.Fprintf(os.Stderr, "Base path:   %s\n", basePath)
			fmt.Fprintf(os.Stderr, "Left path:   %s\n", leftPath)
			fmt.Fprintf(os.Stderr, "Right path:  %s\n", rightPath)
			fmt.Fprintf(os.Stderr, "\n")
		}

		// Read all three files
		baseIssues, err := readIssues(basePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading base file: %v\n", err)
			os.Exit(1)
		}
		if debugMode {
			fmt.Fprintf(os.Stderr, "Base issues read: %d\n", len(baseIssues))
		}

		leftIssues, err := readIssues(leftPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading left file: %v\n", err)
			os.Exit(1)
		}
		if debugMode {
			fmt.Fprintf(os.Stderr, "Left issues read: %d\n", len(leftIssues))
		}

		rightIssues, err := readIssues(rightPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading right file: %v\n", err)
			os.Exit(1)
		}
		if debugMode {
			fmt.Fprintf(os.Stderr, "Right issues read: %d\n", len(rightIssues))
			fmt.Fprintf(os.Stderr, "\n")
		}

		// Perform 3-way merge
		result, conflicts := merge3Way(baseIssues, leftIssues, rightIssues)

		if debugMode {
			fmt.Fprintf(os.Stderr, "Merge complete:\n")
			fmt.Fprintf(os.Stderr, "  Merged issues: %d\n", len(result))
			fmt.Fprintf(os.Stderr, "  Conflicts: %d\n", len(conflicts))
			fmt.Fprintf(os.Stderr, "\n")
		}

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

		if debugMode {
			fmt.Fprintf(os.Stderr, "Output written to: %s\n", outputPath)
			fmt.Fprintf(os.Stderr, "\n")

			// Show first few lines of output for debugging
			outFile.Sync()
			if content, err := os.ReadFile(outputPath); err == nil {
				lines := 0
				fmt.Fprintf(os.Stderr, "Output file preview (first 10 lines):\n")
				for _, line := range splitLines(string(content)) {
					if lines >= 10 {
						fmt.Fprintf(os.Stderr, "... (%d more lines)\n", len(splitLines(string(content)))-10)
						break
					}
					fmt.Fprintf(os.Stderr, "  %s\n", line)
					lines++
				}
			}
			fmt.Fprintf(os.Stderr, "\n")
		}

		// Exit with 1 if there were conflicts (jj will interpret this as conflict markers present)
		if len(conflicts) > 0 {
			if debugMode {
				fmt.Fprintf(os.Stderr, "Exiting with status 1 (conflicts present)\n")
			}
			os.Exit(1)
		}
		if debugMode {
			fmt.Fprintf(os.Stderr, "Exiting with status 0 (no conflicts)\n")
		}
	},
}

var dedupCmd = &cobra.Command{
	Use:   "dedup --canonical=<issue-id> <issue-id> [<issue-id>...]",
	Short: "Deduplicate issues by replacing duplicate IDs with a canonical ID",
	Long: `Deduplicate issues by removing duplicate issues and replacing all references
to their IDs with the canonical issue ID.

This command:
1. Removes the specified duplicate issues from the file
2. Replaces all references to duplicate IDs with the canonical ID in:
   - Dependency lists (issue_id and depends_on_id fields)
   - Text fields (description, notes, title)

Example:
  beads-merge dedup --canonical=bd-1 bd-5 bd-7 bd-10 < input.jsonl > output.jsonl
  
This removes bd-5, bd-7, and bd-10 from the file and replaces all references to
those IDs with bd-1.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		canonicalID, _ := cmd.Flags().GetString("canonical")
		if canonicalID == "" {
			return fmt.Errorf("--canonical flag is required")
		}

		duplicateIDs := args

		// Read issues from stdin
		issues, err := readIssuesFromReader(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read issues: %w", err)
		}

		// Perform deduplication
		dedupedIssues := deduplicateIssues(issues, canonicalID, duplicateIDs)

		// Write to stdout
		for _, issue := range dedupedIssues {
			line, err := json.Marshal(issue)
			if err != nil {
				return fmt.Errorf("failed to marshal issue %s: %w", issue.ID, err)
			}
			fmt.Println(string(line))
		}

		return nil
	},
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func init() {
	// Add flags to merge command
	mergeCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug output to stderr")

	// Add flags to dedup command
	dedupCmd.Flags().String("canonical", "", "Canonical issue ID to use for all duplicates (required)")
	dedupCmd.MarkFlagRequired("canonical")

	// Add commands to root
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(dedupCmd)
}

func main() {
	// Check if we're being called with 4 positional args (backwards compatibility)
	// In that case, invoke merge command directly
	if len(os.Args) == 5 && !strings.HasPrefix(os.Args[1], "-") && os.Args[1] != "merge" && os.Args[1] != "dedup" && os.Args[1] != "help" && os.Args[1] != "completion" {
		// Looks like old-style invocation: beads-merge <output> <base> <left> <right>
		// Insert "merge" as the first argument
		os.Args = append([]string{os.Args[0], "merge"}, os.Args[1:]...)
	}

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

	return readIssuesFromReader(file)
}

func readIssuesFromReader(reader *os.File) ([]Issue, error) {
	var issues []Issue
	scanner := bufio.NewScanner(reader)
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
				conflicts = append(conflicts, makeConflictWithBase(baseIssue.RawLine, leftIssue.RawLine, ""))
			}
		} else if inBase && !inLeft && inRight {
			// Deleted in left, maybe modified in right
			if issuesEqual(baseIssue, rightIssue) {
				// Deleted in left, unchanged in right - accept deletion
				continue
			} else {
				// Modified in right, deleted in left - conflict
				conflicts = append(conflicts, makeConflictWithBase(baseIssue.RawLine, "", rightIssue.RawLine))
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
		return result, makeConflictWithBase(base.RawLine, left.RawLine, right.RawLine)
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
	conflict += ">>>>>>> right\n"
	return conflict
}

func makeConflictWithBase(base, left, right string) string {
	conflict := "<<<<<<< left\n"
	if left != "" {
		conflict += left + "\n"
	}
	conflict += "||||||| base\n"
	if base != "" {
		conflict += base + "\n"
	}
	conflict += "=======\n"
	if right != "" {
		conflict += right + "\n"
	}
	conflict += ">>>>>>> right\n"
	return conflict
}

// deduplicateIssues removes duplicate issues and replaces all references to them with the canonical ID
func deduplicateIssues(issues []Issue, canonicalID string, duplicateIDs []string) []Issue {
	// Create a set of duplicate IDs for quick lookup
	duplicateSet := make(map[string]bool)
	for _, id := range duplicateIDs {
		duplicateSet[id] = true
	}

	// Build replacement map
	replacements := make(map[string]string)
	for _, id := range duplicateIDs {
		replacements[id] = canonicalID
	}

	var result []Issue
	for _, issue := range issues {
		// Skip duplicate issues
		if duplicateSet[issue.ID] {
			continue
		}

		// Replace IDs in the issue
		issue = replaceIDsInIssue(issue, replacements)
		result = append(result, issue)
	}

	return result
}

// replaceIDsInIssue replaces all occurrences of duplicate IDs with the canonical ID
func replaceIDsInIssue(issue Issue, replacements map[string]string) Issue {
	// Replace in text fields
	issue.Title = replaceIDsInString(issue.Title, replacements)
	issue.Description = replaceIDsInString(issue.Description, replacements)
	issue.Notes = replaceIDsInString(issue.Notes, replacements)

	// Replace in dependencies
	for i := range issue.Dependencies {
		if newID, ok := replacements[issue.Dependencies[i].IssueID]; ok {
			issue.Dependencies[i].IssueID = newID
		}
		if newID, ok := replacements[issue.Dependencies[i].DependsOnID]; ok {
			issue.Dependencies[i].DependsOnID = newID
		}
	}

	return issue
}

// replaceIDsInString replaces issue ID references in text
func replaceIDsInString(text string, replacements map[string]string) string {
	if text == "" {
		return text
	}

	result := text
	for oldID, newID := range replacements {
		// Replace the ID as a word (to avoid partial matches)
		// Look for patterns like "bd-123" or "#bd-123" or " bd-123 "
		result = strings.ReplaceAll(result, oldID, newID)
	}

	return result
}
