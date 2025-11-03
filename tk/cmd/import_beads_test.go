package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractPrefixesFromBeads(t *testing.T) {
	issues := []BeadsIssue{
		{ID: "mono-1", Title: "Task 1"},
		{ID: "mono-2", Title: "Task 2"},
		{ID: "foo-1", Title: "Foo task"},
		{ID: "foo-2", Title: "Another foo"},
		{ID: "bar-5", Title: "Bar task"},
	}

	grouped := extractPrefixesFromBeads(issues)

	assert.Len(t, grouped, 3, "should have 3 prefixes")
	assert.Len(t, grouped["mono"], 2, "mono should have 2 issues")
	assert.Len(t, grouped["foo"], 2, "foo should have 2 issues")
	assert.Len(t, grouped["bar"], 1, "bar should have 1 issue")
}

func TestExtractPrefixesFromBeads_SinglePrefix(t *testing.T) {
	issues := []BeadsIssue{
		{ID: "mono-1", Title: "Task 1"},
		{ID: "mono-2", Title: "Task 2"},
	}

	grouped := extractPrefixesFromBeads(issues)

	assert.Len(t, grouped, 1, "should have 1 prefix")
	assert.Len(t, grouped["mono"], 2, "mono should have 2 issues")
}

func TestExtractPrefixesFromBeads_SkipsMalformed(t *testing.T) {
	issues := []BeadsIssue{
		{ID: "mono-1", Title: "Valid"},
		{ID: "invalid", Title: "Malformed"},
		{ID: "mono-2", Title: "Valid"},
	}

	grouped := extractPrefixesFromBeads(issues)

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
			got, err := parseBeadsNumber(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestImportBeads_PreservesNumbering(t *testing.T) {
	db := openTempDB(t)

	// Create project with alias
	projectUID, err := createProjectForImport(db, "test", "bd-test")
	require.NoError(t, err)

	// Import issue with specific number
	issue := BeadsIssue{
		ID:       "test-123",
		Title:    "Test task",
		Priority: 1,
		Status:   "open",
	}

	taskUID, err := importBeadsIssue(db, issue, projectUID, 123)
	require.NoError(t, err)

	// Verify task number is 123
	var number int64
	err = db.Db.QueryRow("SELECT number FROM task_numbers WHERE task_uid = ?", taskUID).Scan(&number)
	require.NoError(t, err)
	assert.Equal(t, int64(123), number)
}

func TestImportBeads_AlwaysCreatesNewProject(t *testing.T) {
	db := openTempDB(t)

	// Create first project with alias "bd-mono"
	uid1, err := createProjectForImport(db, "mono", "bd-mono")
	require.NoError(t, err)

	// Create second project - different alias required (same node can't have duplicate)
	uid2, err := createProjectForImport(db, "mono", "beads-mono")
	require.NoError(t, err)

	// Should be different UIDs
	assert.NotEqual(t, uid1, uid2, "should create different projects")

	// Should have different aliases
	var count int
	err = db.Db.QueryRow("SELECT COUNT(*) FROM project_aliases WHERE alias = 'bd-mono'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should have 1 project with 'bd-mono' alias")

	err = db.Db.QueryRow("SELECT COUNT(*) FROM project_aliases WHERE alias = 'beads-mono'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should have 1 project with 'beads-mono' alias")
}

func TestImportBeads_CreatesSingleAlias(t *testing.T) {
	db := openTempDB(t)

	projectUID, err := createProjectForImport(db, "test", "bd-test")
	require.NoError(t, err)

	// Verify only one alias created
	var aliases []string
	rows, err := db.Db.Query("SELECT alias FROM project_aliases WHERE project_uid = ?", projectUID)
	require.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var alias string
		rows.Scan(&alias)
		aliases = append(aliases, alias)
	}

	assert.Len(t, aliases, 1, "should have 1 alias")
	assert.Contains(t, aliases, "bd-test", "should have the specified alias")
}
