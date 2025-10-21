package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadIssues(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name:    "empty file",
			content: "",
			want:    0,
			wantErr: false,
		},
		{
			name: "single issue",
			content: `{"id":"bd-1","title":"Test","created_at":"2025-10-16T20:51:29.646949+02:00"}
`,
			want:    1,
			wantErr: false,
		},
		{
			name: "multiple issues",
			content: `{"id":"bd-1","title":"Test 1","created_at":"2025-10-16T20:51:29.646949+02:00"}
{"id":"bd-2","title":"Test 2","created_at":"2025-10-16T20:51:30.646949+02:00"}
{"id":"bd-3","title":"Test 3","created_at":"2025-10-16T20:51:31.646949+02:00"}
`,
			want:    3,
			wantErr: false,
		},
		{
			name:    "invalid json",
			content: `{invalid json}`,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.jsonl")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			issues, err := readIssues(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("readIssues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(issues) != tt.want {
				t.Errorf("readIssues() got %d issues, want %d", len(issues), tt.want)
			}
		})
	}
}

func TestMergeField(t *testing.T) {
	tests := []struct {
		name  string
		base  string
		left  string
		right string
		want  string
	}{
		{
			name:  "no change",
			base:  "original",
			left:  "original",
			right: "original",
			want:  "original",
		},
		{
			name:  "only left changed",
			base:  "original",
			left:  "changed",
			right: "original",
			want:  "changed",
		},
		{
			name:  "only right changed",
			base:  "original",
			left:  "original",
			right: "changed",
			want:  "changed",
		},
		{
			name:  "both changed to same",
			base:  "original",
			left:  "changed",
			right: "changed",
			want:  "changed",
		},
		{
			name:  "both changed to different",
			base:  "original",
			left:  "left-change",
			right: "right-change",
			want:  "left-change", // Takes left when conflicting
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeField(tt.base, tt.left, tt.right)
			if got != tt.want {
				t.Errorf("mergeField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxTime(t *testing.T) {
	tests := []struct {
		name string
		t1   string
		t2   string
		want string
	}{
		{
			name: "both empty",
			t1:   "",
			t2:   "",
			want: "",
		},
		{
			name: "t1 empty",
			t1:   "",
			t2:   "2025-10-16T20:51:29.646949+02:00",
			want: "2025-10-16T20:51:29.646949+02:00",
		},
		{
			name: "t2 empty",
			t1:   "2025-10-16T20:51:29.646949+02:00",
			t2:   "",
			want: "2025-10-16T20:51:29.646949+02:00",
		},
		{
			name: "t1 later with fractional seconds",
			t1:   "2025-10-17T20:51:29.646949+02:00",
			t2:   "2025-10-16T20:51:29.646949+02:00",
			want: "2025-10-17T20:51:29.646949+02:00",
		},
		{
			name: "t2 later with fractional seconds",
			t1:   "2025-10-16T20:51:29.646949+02:00",
			t2:   "2025-10-17T20:51:29.646949+02:00",
			want: "2025-10-17T20:51:29.646949+02:00",
		},
		{
			name: "t1 later without fractional seconds (RFC3339)",
			t1:   "2025-10-17T20:51:29+02:00",
			t2:   "2025-10-16T20:51:29+02:00",
			want: "2025-10-17T20:51:29+02:00",
		},
		{
			name: "t2 later without fractional seconds (RFC3339)",
			t1:   "2025-10-16T20:51:29+02:00",
			t2:   "2025-10-17T20:51:29+02:00",
			want: "2025-10-17T20:51:29+02:00",
		},
		{
			name: "mixed formats - t1 with fractions, t2 without",
			t1:   "2025-10-17T20:51:29.646949+02:00",
			t2:   "2025-10-16T20:51:29+02:00",
			want: "2025-10-17T20:51:29.646949+02:00",
		},
		{
			name: "mixed formats - t1 without fractions, t2 with",
			t1:   "2025-10-16T20:51:29+02:00",
			t2:   "2025-10-17T20:51:29.646949+02:00",
			want: "2025-10-17T20:51:29.646949+02:00",
		},
		{
			name: "very close timestamps differing only in microseconds",
			t1:   "2025-10-16T20:51:29.646949+02:00",
			t2:   "2025-10-16T20:51:29.646950+02:00",
			want: "2025-10-16T20:51:29.646950+02:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxTime(tt.t1, tt.t2)
			if got != tt.want {
				t.Errorf("maxTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeDependencies(t *testing.T) {
	tests := []struct {
		name  string
		left  []Dependency
		right []Dependency
		want  int // Number of unique dependencies
	}{
		{
			name:  "both empty",
			left:  []Dependency{},
			right: []Dependency{},
			want:  0,
		},
		{
			name: "only left",
			left: []Dependency{
				{IssueID: "bd-1", DependsOnID: "bd-2", Type: "blocks"},
			},
			right: []Dependency{},
			want:  1,
		},
		{
			name: "only right",
			left: []Dependency{},
			right: []Dependency{
				{IssueID: "bd-1", DependsOnID: "bd-2", Type: "blocks"},
			},
			want: 1,
		},
		{
			name: "duplicates removed",
			left: []Dependency{
				{IssueID: "bd-1", DependsOnID: "bd-2", Type: "blocks"},
			},
			right: []Dependency{
				{IssueID: "bd-1", DependsOnID: "bd-2", Type: "blocks"},
			},
			want: 1,
		},
		{
			name: "unique merged",
			left: []Dependency{
				{IssueID: "bd-1", DependsOnID: "bd-2", Type: "blocks"},
			},
			right: []Dependency{
				{IssueID: "bd-1", DependsOnID: "bd-3", Type: "blocks"},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeDependencies(tt.left, tt.right)
			if len(got) != tt.want {
				t.Errorf("mergeDependencies() got %d dependencies, want %d", len(got), tt.want)
			}
		})
	}
}

func TestMerge3Way_SimpleCase(t *testing.T) {
	// Create base, left, and right issues
	base := []Issue{
		{
			ID:          "bd-1",
			Title:       "Original title",
			Description: "Original description",
			Status:      "open",
			Priority:    1,
			CreatedAt:   "2025-10-16T20:51:29.646949+02:00",
			CreatedBy:   "user1",
		},
	}

	left := []Issue{
		{
			ID:          "bd-1",
			Title:       "Updated title",
			Description: "Original description",
			Status:      "open",
			Priority:    1,
			CreatedAt:   "2025-10-16T20:51:29.646949+02:00",
			CreatedBy:   "user1",
			UpdatedAt:   "2025-10-17T10:00:00.000000+02:00",
		},
	}

	right := []Issue{
		{
			ID:          "bd-1",
			Title:       "Original title",
			Description: "Updated description",
			Status:      "open",
			Priority:    1,
			CreatedAt:   "2025-10-16T20:51:29.646949+02:00",
			CreatedBy:   "user1",
			UpdatedAt:   "2025-10-17T11:00:00.000000+02:00",
		},
	}

	result, conflicts := merge3Way(base, left, right)

	if len(conflicts) != 0 {
		t.Errorf("merge3Way() got %d conflicts, want 0", len(conflicts))
	}

	if len(result) != 1 {
		t.Fatalf("merge3Way() got %d results, want 1", len(result))
	}

	merged := result[0]
	if merged.Title != "Updated title" {
		t.Errorf("merged.Title = %v, want %v", merged.Title, "Updated title")
	}
	if merged.Description != "Updated description" {
		t.Errorf("merged.Description = %v, want %v", merged.Description, "Updated description")
	}
	// UpdatedAt should be the max
	if merged.UpdatedAt != "2025-10-17T11:00:00.000000+02:00" {
		t.Errorf("merged.UpdatedAt = %v, want %v", merged.UpdatedAt, "2025-10-17T11:00:00.000000+02:00")
	}
}

func TestMerge3Way_ConflictingChanges(t *testing.T) {
	base := []Issue{
		{
			ID:        "bd-1",
			Title:     "Original title",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
			RawLine:   `{"id":"bd-1","title":"Original title","status":"open"}`,
		},
	}

	left := []Issue{
		{
			ID:        "bd-1",
			Title:     "Left title",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
			RawLine:   `{"id":"bd-1","title":"Left title","status":"open"}`,
		},
	}

	right := []Issue{
		{
			ID:        "bd-1",
			Title:     "Right title",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
			RawLine:   `{"id":"bd-1","title":"Right title","status":"open"}`,
		},
	}

	result, conflicts := merge3Way(base, left, right)

	if len(conflicts) != 1 {
		t.Errorf("merge3Way() got %d conflicts, want 1", len(conflicts))
	}

	if len(result) != 0 {
		t.Errorf("merge3Way() got %d results, want 0 (all conflicts)", len(result))
	}

	if len(conflicts) > 0 {
		if !strings.Contains(conflicts[0], "<<<<<<< left") {
			t.Errorf("conflict block missing left marker")
		}
		if !strings.Contains(conflicts[0], ">>>>>>> right") {
			t.Errorf("conflict block missing right marker")
		}
	}
}

func TestMerge3Way_AddedInBoth(t *testing.T) {
	base := []Issue{}

	left := []Issue{
		{
			ID:        "bd-1",
			Title:     "New issue",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
		},
	}

	right := []Issue{
		{
			ID:        "bd-1",
			Title:     "New issue",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
		},
	}

	result, conflicts := merge3Way(base, left, right)

	if len(conflicts) != 0 {
		t.Errorf("merge3Way() got %d conflicts, want 0 (identical additions)", len(conflicts))
	}

	if len(result) != 1 {
		t.Errorf("merge3Way() got %d results, want 1", len(result))
	}
}

func TestMerge3Way_DeletedInBoth(t *testing.T) {
	base := []Issue{
		{
			ID:        "bd-1",
			Title:     "To be deleted",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
		},
	}

	left := []Issue{}
	right := []Issue{}

	result, conflicts := merge3Way(base, left, right)

	if len(conflicts) != 0 {
		t.Errorf("merge3Way() got %d conflicts, want 0", len(conflicts))
	}

	if len(result) != 0 {
		t.Errorf("merge3Way() got %d results, want 0 (deleted)", len(result))
	}
}

func TestMerge3Way_ModifiedAndDeleted(t *testing.T) {
	base := []Issue{
		{
			ID:        "bd-1",
			Title:     "Original",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
			RawLine:   `{"id":"bd-1","title":"Original","status":"open"}`,
		},
	}

	left := []Issue{
		{
			ID:        "bd-1",
			Title:     "Modified",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
			RawLine:   `{"id":"bd-1","title":"Modified","status":"open"}`,
		},
	}

	right := []Issue{} // Deleted

	result, conflicts := merge3Way(base, left, right)

	if len(conflicts) != 1 {
		t.Errorf("merge3Way() got %d conflicts, want 1 (modified vs deleted)", len(conflicts))
	}

	if len(result) != 0 {
		t.Errorf("merge3Way() got %d results, want 0", len(result))
	}
}

func TestMerge3Way_DependencyMerge(t *testing.T) {
	base := []Issue{
		{
			ID:        "bd-1",
			Title:     "Issue with deps",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
			Dependencies: []Dependency{
				{IssueID: "bd-1", DependsOnID: "bd-2", Type: "blocks"},
			},
		},
	}

	left := []Issue{
		{
			ID:        "bd-1",
			Title:     "Issue with deps",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
			Dependencies: []Dependency{
				{IssueID: "bd-1", DependsOnID: "bd-2", Type: "blocks"},
				{IssueID: "bd-1", DependsOnID: "bd-3", Type: "related"},
			},
		},
	}

	right := []Issue{
		{
			ID:        "bd-1",
			Title:     "Issue with deps",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
			Dependencies: []Dependency{
				{IssueID: "bd-1", DependsOnID: "bd-2", Type: "blocks"},
				{IssueID: "bd-1", DependsOnID: "bd-4", Type: "parent-child"},
			},
		},
	}

	result, conflicts := merge3Way(base, left, right)

	if len(conflicts) != 0 {
		t.Errorf("merge3Way() got %d conflicts, want 0", len(conflicts))
	}

	if len(result) != 1 {
		t.Fatalf("merge3Way() got %d results, want 1", len(result))
	}

	merged := result[0]
	// Should have 3 unique dependencies: bd-2, bd-3, bd-4
	if len(merged.Dependencies) != 3 {
		t.Errorf("merged dependencies count = %d, want 3", len(merged.Dependencies))
	}
}

func TestMerge3Way_StatusChange(t *testing.T) {
	base := []Issue{
		{
			ID:        "bd-1",
			Title:     "Test issue",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
		},
	}

	left := []Issue{
		{
			ID:        "bd-1",
			Title:     "Test issue",
			Status:    "open",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
		},
	}

	right := []Issue{
		{
			ID:        "bd-1",
			Title:     "Test issue",
			Status:    "closed",
			ClosedAt:  "2025-10-17T10:00:00.000000+02:00",
			CreatedAt: "2025-10-16T20:51:29.646949+02:00",
			CreatedBy: "user1",
		},
	}

	result, conflicts := merge3Way(base, left, right)

	if len(conflicts) != 0 {
		t.Errorf("merge3Way() got %d conflicts, want 0", len(conflicts))
	}

	if len(result) != 1 {
		t.Fatalf("merge3Way() got %d results, want 1", len(result))
	}

	merged := result[0]
	if merged.Status != "closed" {
		t.Errorf("merged.Status = %v, want closed", merged.Status)
	}
	if merged.ClosedAt != "2025-10-17T10:00:00.000000+02:00" {
		t.Errorf("merged.ClosedAt = %v, want %v", merged.ClosedAt, "2025-10-17T10:00:00.000000+02:00")
	}
}

func TestMakeKey(t *testing.T) {
	issue := Issue{
		ID:        "bd-1",
		CreatedAt: "2025-10-16T20:51:29.646949+02:00",
		CreatedBy: "user1",
		Title:     "Test",
	}

	key := makeKey(issue)

	if key.ID != "bd-1" {
		t.Errorf("key.ID = %v, want bd-1", key.ID)
	}
	if key.CreatedAt != "2025-10-16T20:51:29.646949+02:00" {
		t.Errorf("key.CreatedAt = %v", key.CreatedAt)
	}
	if key.CreatedBy != "user1" {
		t.Errorf("key.CreatedBy = %v, want user1", key.CreatedBy)
	}
}

func TestIssuesEqual(t *testing.T) {
	issue1 := Issue{
		ID:        "bd-1",
		Title:     "Test",
		CreatedAt: "2025-10-16T20:51:29.646949+02:00",
	}

	issue2 := Issue{
		ID:        "bd-1",
		Title:     "Test",
		CreatedAt: "2025-10-16T20:51:29.646949+02:00",
	}

	issue3 := Issue{
		ID:        "bd-1",
		Title:     "Different",
		CreatedAt: "2025-10-16T20:51:29.646949+02:00",
	}

	if !issuesEqual(issue1, issue2) {
		t.Error("issuesEqual() should be true for identical issues")
	}

	if issuesEqual(issue1, issue3) {
		t.Error("issuesEqual() should be false for different issues")
	}
}

func TestMakeConflict(t *testing.T) {
	left := `{"id":"bd-1","title":"Left"}`
	right := `{"id":"bd-1","title":"Right"}`

	conflict := makeConflict(left, right)

	if !strings.Contains(conflict, "<<<<<<< left") {
		t.Error("conflict missing left marker")
	}
	if !strings.Contains(conflict, "=======") {
		t.Error("conflict missing separator")
	}
	if !strings.Contains(conflict, ">>>>>>> right") {
		t.Error("conflict missing right marker")
	}
	if !strings.Contains(conflict, left) {
		t.Error("conflict missing left content")
	}
	if !strings.Contains(conflict, right) {
		t.Error("conflict missing right content")
	}
}

func TestIntegration_FullMerge(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	baseContent := `{"id":"bd-1","title":"Issue 1","status":"open","priority":1,"created_at":"2025-10-16T20:51:29.646949+02:00","created_by":"user1"}
{"id":"bd-2","title":"Issue 2","status":"open","priority":2,"created_at":"2025-10-16T20:51:30.646949+02:00","created_by":"user1"}
`
	leftContent := `{"id":"bd-1","title":"Issue 1 - Updated","status":"open","priority":1,"created_at":"2025-10-16T20:51:29.646949+02:00","created_by":"user1","updated_at":"2025-10-17T10:00:00.000000+02:00"}
{"id":"bd-2","title":"Issue 2","status":"closed","priority":2,"created_at":"2025-10-16T20:51:30.646949+02:00","created_by":"user1","closed_at":"2025-10-17T09:00:00.000000+02:00"}
{"id":"bd-3","title":"New issue","status":"open","priority":1,"created_at":"2025-10-17T10:00:00.000000+02:00","created_by":"user1"}
`
	rightContent := `{"id":"bd-1","title":"Issue 1","description":"Added description","status":"open","priority":1,"created_at":"2025-10-16T20:51:29.646949+02:00","created_by":"user1","updated_at":"2025-10-17T11:00:00.000000+02:00"}
{"id":"bd-2","title":"Issue 2","status":"open","priority":2,"created_at":"2025-10-16T20:51:30.646949+02:00","created_by":"user1"}
`

	basePath := filepath.Join(tmpDir, "base.jsonl")
	leftPath := filepath.Join(tmpDir, "left.jsonl")
	rightPath := filepath.Join(tmpDir, "right.jsonl")

	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write base file: %v", err)
	}
	if err := os.WriteFile(leftPath, []byte(leftContent), 0644); err != nil {
		t.Fatalf("Failed to write left file: %v", err)
	}
	if err := os.WriteFile(rightPath, []byte(rightContent), 0644); err != nil {
		t.Fatalf("Failed to write right file: %v", err)
	}

	// Read files
	base, err := readIssues(basePath)
	if err != nil {
		t.Fatalf("Failed to read base: %v", err)
	}

	left, err := readIssues(leftPath)
	if err != nil {
		t.Fatalf("Failed to read left: %v", err)
	}

	right, err := readIssues(rightPath)
	if err != nil {
		t.Fatalf("Failed to read right: %v", err)
	}

	// Perform merge
	result, conflicts := merge3Way(base, left, right)

	// Validate results
	if len(conflicts) != 0 {
		t.Errorf("Expected no conflicts, got %d: %v", len(conflicts), conflicts)
	}

	// Should have 3 issues: bd-1 (merged), bd-2 (conflict on status change), bd-3 (added in left)
	// Actually bd-2 should merge cleanly since left closed it and right didn't change it
	if len(result) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(result))
	}

	// Find bd-1 in results
	var bd1 *Issue
	for i := range result {
		if result[i].ID == "bd-1" {
			bd1 = &result[i]
			break
		}
	}

	if bd1 == nil {
		t.Fatal("bd-1 not found in results")
	}

	// Check bd-1 merge
	if bd1.Title != "Issue 1 - Updated" {
		t.Errorf("bd-1 title = %v, want 'Issue 1 - Updated'", bd1.Title)
	}
	if bd1.Description != "Added description" {
		t.Errorf("bd-1 description = %v, want 'Added description'", bd1.Description)
	}
	// UpdatedAt should be the max
	if bd1.UpdatedAt != "2025-10-17T11:00:00.000000+02:00" {
		t.Errorf("bd-1 updated_at = %v, want latest", bd1.UpdatedAt)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	original := Issue{
		ID:          "bd-1",
		Title:       "Test Issue",
		Description: "Test description",
		Status:      "open",
		Priority:    2,
		IssueType:   "task",
		CreatedAt:   "2025-10-16T20:51:29.646949+02:00",
		UpdatedAt:   "2025-10-17T10:00:00.000000+02:00",
		CreatedBy:   "user1",
		Dependencies: []Dependency{
			{IssueID: "bd-1", DependsOnID: "bd-2", Type: "blocks", CreatedAt: "2025-10-17T10:00:00.000000+02:00", CreatedBy: "user1"},
		},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded Issue
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: %v != %v", decoded.ID, original.ID)
	}
	if decoded.Title != original.Title {
		t.Errorf("Title mismatch: %v != %v", decoded.Title, original.Title)
	}
	if len(decoded.Dependencies) != len(original.Dependencies) {
		t.Errorf("Dependencies count mismatch: %v != %v", len(decoded.Dependencies), len(original.Dependencies))
	}
}

func TestReplaceIDsInString(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		replacements map[string]string
		want         string
	}{
		{
			name:         "empty string",
			text:         "",
			replacements: map[string]string{"bd-5": "bd-1"},
			want:         "",
		},
		{
			name:         "no matches",
			text:         "This is a test with bd-2",
			replacements: map[string]string{"bd-5": "bd-1"},
			want:         "This is a test with bd-2",
		},
		{
			name:         "single replacement",
			text:         "This relates to bd-5",
			replacements: map[string]string{"bd-5": "bd-1"},
			want:         "This relates to bd-1",
		},
		{
			name:         "multiple replacements",
			text:         "This relates to bd-5 and bd-7",
			replacements: map[string]string{"bd-5": "bd-1", "bd-7": "bd-1"},
			want:         "This relates to bd-1 and bd-1",
		},
		{
			name:         "replacement in middle of text",
			text:         "See bd-5 for details",
			replacements: map[string]string{"bd-5": "bd-1"},
			want:         "See bd-1 for details",
		},
		{
			name:         "multiple occurrences",
			text:         "bd-5 is related to bd-5 again",
			replacements: map[string]string{"bd-5": "bd-1"},
			want:         "bd-1 is related to bd-1 again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceIDsInString(tt.text, tt.replacements)
			if got != tt.want {
				t.Errorf("replaceIDsInString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceIDsInIssue(t *testing.T) {
	replacements := map[string]string{
		"bd-5": "bd-1",
		"bd-7": "bd-1",
	}

	issue := Issue{
		ID:          "bd-2",
		Title:       "Issue referencing bd-5",
		Description: "This is related to bd-5 and bd-7",
		Notes:       "See bd-5",
		Status:      "open",
		CreatedAt:   "2025-10-16T20:51:29.646949+02:00",
		CreatedBy:   "user1",
		Dependencies: []Dependency{
			{IssueID: "bd-2", DependsOnID: "bd-5", Type: "blocks"},
			{IssueID: "bd-7", DependsOnID: "bd-3", Type: "related"},
		},
	}

	result := replaceIDsInIssue(issue, replacements)

	if result.Title != "Issue referencing bd-1" {
		t.Errorf("Title not replaced: %v", result.Title)
	}
	if result.Description != "This is related to bd-1 and bd-1" {
		t.Errorf("Description not replaced: %v", result.Description)
	}
	if result.Notes != "See bd-1" {
		t.Errorf("Notes not replaced: %v", result.Notes)
	}

	// Check dependencies
	if result.Dependencies[0].DependsOnID != "bd-1" {
		t.Errorf("Dependency depends_on_id not replaced: %v", result.Dependencies[0].DependsOnID)
	}
	if result.Dependencies[1].IssueID != "bd-1" {
		t.Errorf("Dependency issue_id not replaced: %v", result.Dependencies[1].IssueID)
	}
}

func TestDeduplicateIssues(t *testing.T) {
	issues := []Issue{
		{
			ID:          "bd-1",
			Title:       "Canonical issue",
			Description: "Main issue",
			Status:      "open",
			CreatedAt:   "2025-10-16T20:51:29.646949+02:00",
			CreatedBy:   "user1",
		},
		{
			ID:          "bd-2",
			Title:       "Another issue",
			Description: "References bd-5",
			Status:      "open",
			CreatedAt:   "2025-10-16T20:51:30.646949+02:00",
			CreatedBy:   "user1",
			Dependencies: []Dependency{
				{IssueID: "bd-2", DependsOnID: "bd-5", Type: "blocks"},
			},
		},
		{
			ID:          "bd-5",
			Title:       "Duplicate issue",
			Description: "Should be removed",
			Status:      "open",
			CreatedAt:   "2025-10-16T20:51:31.646949+02:00",
			CreatedBy:   "user1",
		},
		{
			ID:          "bd-7",
			Title:       "Another duplicate",
			Description: "Also should be removed",
			Status:      "open",
			CreatedAt:   "2025-10-16T20:51:32.646949+02:00",
			CreatedBy:   "user1",
		},
		{
			ID:          "bd-10",
			Title:       "Issue with dependency",
			Description: "Depends on bd-7",
			Status:      "open",
			CreatedAt:   "2025-10-16T20:51:33.646949+02:00",
			CreatedBy:   "user1",
			Dependencies: []Dependency{
				{IssueID: "bd-10", DependsOnID: "bd-7", Type: "blocks"},
			},
		},
	}

	canonicalID := "bd-1"
	duplicateIDs := []string{"bd-5", "bd-7"}

	result := deduplicateIssues(issues, canonicalID, duplicateIDs)

	// Should have 3 issues left (bd-1, bd-2, bd-10)
	if len(result) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(result))
	}

	// Find bd-2 and check its dependency
	var bd2 *Issue
	for i := range result {
		if result[i].ID == "bd-2" {
			bd2 = &result[i]
			break
		}
	}
	if bd2 == nil {
		t.Fatal("bd-2 not found in result")
	}
	if bd2.Dependencies[0].DependsOnID != "bd-1" {
		t.Errorf("bd-2 dependency not replaced: %v", bd2.Dependencies[0].DependsOnID)
	}
	if bd2.Description != "References bd-1" {
		t.Errorf("bd-2 description not replaced: %v", bd2.Description)
	}

	// Find bd-10 and check its dependency
	var bd10 *Issue
	for i := range result {
		if result[i].ID == "bd-10" {
			bd10 = &result[i]
			break
		}
	}
	if bd10 == nil {
		t.Fatal("bd-10 not found in result")
	}
	if bd10.Dependencies[0].DependsOnID != "bd-1" {
		t.Errorf("bd-10 dependency not replaced: %v", bd10.Dependencies[0].DependsOnID)
	}
	if bd10.Description != "Depends on bd-1" {
		t.Errorf("bd-10 description not replaced: %v", bd10.Description)
	}

	// Verify bd-5 and bd-7 are not in result
	for _, issue := range result {
		if issue.ID == "bd-5" || issue.ID == "bd-7" {
			t.Errorf("Duplicate issue %s should have been removed", issue.ID)
		}
	}
}
