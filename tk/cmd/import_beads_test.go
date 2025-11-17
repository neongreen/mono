package cmd

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/import/beads"
	"github.com/stretchr/testify/assert"
)

func TestExtractPrefixesFromBeads(t *testing.T) {
	issues := []beads.BeadsIssue{
		{ID: "mono-1", Title: "Task 1"},
		{ID: "mono-2", Title: "Task 2"},
		{ID: "foo-1", Title: "Foo task"},
		{ID: "foo-2", Title: "Another foo"},
		{ID: "bar-5", Title: "Bar task"},
	}

	grouped := beads.ExtractPrefixesFromBeads(issues)

	assert.Len(t, grouped, 3, "should have 3 prefixes")
	assert.Len(t, grouped["mono"], 2, "mono should have 2 issues")
	assert.Len(t, grouped["foo"], 2, "foo should have 2 issues")
	assert.Len(t, grouped["bar"], 1, "bar should have 1 issue")
}

func TestExtractPrefixesFromBeads_SinglePrefix(t *testing.T) {
	issues := []beads.BeadsIssue{
		{ID: "mono-1", Title: "Task 1"},
		{ID: "mono-2", Title: "Task 2"},
	}

	grouped := beads.ExtractPrefixesFromBeads(issues)

	assert.Len(t, grouped, 1, "should have 1 prefix")
	assert.Len(t, grouped["mono"], 2, "mono should have 2 issues")
}

func TestExtractPrefixesFromBeads_SkipsMalformed(t *testing.T) {
	issues := []beads.BeadsIssue{
		{ID: "mono-1", Title: "Valid"},
		{ID: "invalid", Title: "Malformed"},
		{ID: "mono-2", Title: "Valid"},
	}

	grouped := beads.ExtractPrefixesFromBeads(issues)

	assert.Len(t, grouped, 1, "should have 1 prefix (malformed skipped)")
	assert.Len(t, grouped["mono"], 2, "mono should have 2 issues")
}

func TestParseBeadsNumber(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    int64
		wantErr bool
	}{
		{
			name: "standard ID",
			id:   "mono-123",
			want: 123,
		},
		{
			name: "single digit",
			id:   "foo-1",
			want: 1,
		},
		{
			name: "large number",
			id:   "bar-9999",
			want: 9999,
		},
		{
			name:    "no dash",
			id:      "invalid",
			wantErr: true,
		},
		{
			name:    "non-numeric",
			id:      "mono-abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := beads.ParseBeadsNumber(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestImportBeads_PreservesNumbering(t *testing.T) {
	db := openTempDB(t)

	// Create project with alias
	projectUID, err := beads.CreateProjectForImport(db, "test", "bd-test", "test-actor")
	assert.NoError(t, err)

	// Import issue with specific number
	issue := beads.BeadsIssue{
		ID:       "test-123",
		Title:    "Test task",
		Priority: 1,
		Status:   "open",
	}

	taskUID, err := beads.ImportBeadsIssue(db, issue, projectUID, 123)
	assert.NoError(t, err)

	// Verify task number is 123
	var number int64
	err = db.Db.QueryRow("SELECT number FROM task_numbers WHERE task_uid = ?", taskUID).Scan(&number)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), number)
}

func TestImportBeads_AlwaysCreatesNewProject(t *testing.T) {
	db := openTempDB(t)

	// Create first project
	uid1, err := beads.CreateProjectForImport(db, "mono", "bd-mono", "test-actor")
	assert.NoError(t, err)

	// Create second project (same name, should create new project)
	uid2, err := beads.CreateProjectForImport(db, "mono", "beads-mono", "test-actor")
	assert.NoError(t, err)

	// Should be different UIDs
	assert.NotEqual(t, uid1, uid2, "should create different projects")

	// Verify both projects exist
	var count int
	err = db.Db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count, "should have 2 projects")
}

func TestImportBeads_CreatesProject(t *testing.T) {
	db := openTempDB(t)

	projectUID, err := beads.CreateProjectForImport(db, "test", "bd-test", "test-actor")
	assert.NoError(t, err)

	// Verify project was created
	var name string
	err = db.Db.QueryRow("SELECT name FROM projects WHERE project_uid = ?", projectUID).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "test", name, "should have correct project name")
}
