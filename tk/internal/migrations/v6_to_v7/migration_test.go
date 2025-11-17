package v6_to_v7

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// mockDB implements the DB interface for testing
type mockDB struct {
	db      *sql.DB
	version int
}

func (m *mockDB) Exec(query string, args ...any) (sql.Result, error) {
	return m.db.Exec(query, args...)
}

func (m *mockDB) SetDBVersion(version int) error {
	m.version = version
	_, err := m.db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('db_version', ?)`, version)
	return err
}

func TestMigrate(t *testing.T) {
	// Create temp database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create initial schema (v6) - including container tables from v5_to_v6
	_, err = db.Exec(`
		CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT);
		INSERT INTO metadata (key, value) VALUES ('db_version', '6');

		CREATE TABLE projects (
			project_uid TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL,
			is_synthetic INTEGER DEFAULT 0
		);

		CREATE TABLE tasks (
			task_uid TEXT PRIMARY KEY,
			project_uid TEXT NOT NULL,
			created_node TEXT NOT NULL,
			title TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL
		);

		CREATE TABLE container_kinds (
			name TEXT PRIMARY KEY,
			primitive TEXT NOT NULL,
			description TEXT,
			deprecated INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Insert a test task to verify migration preserves data
	_, err = db.Exec(`
		INSERT INTO projects (project_uid, type, name, description, created_at, created_by)
		VALUES ('prj_test', 'local', 'test', 'test project', 0, 'test');

		INSERT INTO tasks (task_uid, project_uid, created_node, title, created_at, created_by)
		VALUES ('tsk_test', 'prj_test', 'node1', 'Test task', 0, 'test');
	`)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	mock := &mockDB{db: db, version: 6}

	// Run migration
	if err := Migrate(mock); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Verify item_kinds table was created
	var tableCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='item_kinds'
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("failed to check item_kinds table: %v", err)
	}
	if tableCount != 1 {
		t.Errorf("item_kinds table not created")
	}

	// Verify all builtin item kinds were inserted
	// Check the count matches expected number (24 kinds)
	var kindCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM item_kinds WHERE builtin = 1`).Scan(&kindCount)
	if err != nil {
		t.Fatalf("failed to count builtin kinds: %v", err)
	}
	if kindCount != 24 {
		t.Errorf("builtin item kind count = %d, want 24", kindCount)
	}

	// Verify a sample of kinds exist and are marked as builtin
	sampleKinds := []string{"task", "bug", "promise", "regret"}
	for _, kind := range sampleKinds {
		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM item_kinds WHERE name = ? AND builtin = 1`, kind).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check %s kind: %v", kind, err)
		}
		if count != 1 {
			t.Errorf("'%s' builtin item kind not found", kind)
		}
	}

	// Verify item_kind column was added to tasks table
	var columnCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks')
		WHERE name = 'item_kind'
	`).Scan(&columnCount)
	if err != nil {
		t.Fatalf("failed to check item_kind column: %v", err)
	}
	if columnCount != 1 {
		t.Errorf("item_kind column not added to tasks table")
	}

	// Verify existing task has item_kind='task'
	var itemKind string
	err = db.QueryRow(`SELECT item_kind FROM tasks WHERE task_uid = 'tsk_test'`).Scan(&itemKind)
	if err != nil {
		t.Fatalf("failed to query task item_kind: %v", err)
	}
	if itemKind != "task" {
		t.Errorf("existing task item_kind = %q, want 'task'", itemKind)
	}

	// Verify index was created
	var idxCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='index' AND name='idx_tasks_item_kind'
	`).Scan(&idxCount)
	if err != nil {
		t.Fatalf("failed to check index: %v", err)
	}
	if idxCount != 1 {
		t.Errorf("idx_tasks_item_kind index not created")
	}

	// Verify version was updated
	if mock.version != 7 {
		t.Errorf("database version = %d, want 7", mock.version)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	// Create temp database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create initial schema (v6)
	_, err = db.Exec(`
		CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT);
		INSERT INTO metadata (key, value) VALUES ('db_version', '6');

		CREATE TABLE tasks (
			task_uid TEXT PRIMARY KEY,
			project_uid TEXT NOT NULL,
			created_node TEXT NOT NULL,
			title TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	mock := &mockDB{db: db, version: 6}

	// Run migration twice - should be idempotent
	if err := Migrate(mock); err != nil {
		t.Fatalf("first Migrate() failed: %v", err)
	}

	if err := Migrate(mock); err != nil {
		t.Fatalf("second Migrate() failed (not idempotent): %v", err)
	}

	// Verify still at version 7
	if mock.version != 7 {
		t.Errorf("database version after second migration = %d, want 7", mock.version)
	}

	// Verify all builtin kinds exist only once (not duplicated)
	var kindCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM item_kinds WHERE builtin = 1`).Scan(&kindCount)
	if err != nil {
		t.Fatalf("failed to count builtin kinds: %v", err)
	}
	if kindCount != 24 {
		t.Errorf("builtin item kind count after second migration = %d, want 24 (idempotency check)", kindCount)
	}

	// Verify sample kinds exist exactly once
	sampleKinds := []string{"task", "bug", "promise", "regret"}
	for _, kind := range sampleKinds {
		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM item_kinds WHERE name = ?`, kind).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check %s kind: %v", kind, err)
		}
		if count != 1 {
			t.Errorf("'%s' item kind count = %d, want 1 (idempotency check)", kind, count)
		}
	}
}

func TestMigrateTableSchema(t *testing.T) {
	// Create temp database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create minimal schema (v6)
	_, err = db.Exec(`
		CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT);
		INSERT INTO metadata (key, value) VALUES ('db_version', '6');

		CREATE TABLE tasks (
			task_uid TEXT PRIMARY KEY,
			project_uid TEXT NOT NULL,
			created_node TEXT NOT NULL,
			title TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	mock := &mockDB{db: db, version: 6}

	// Run migration
	if err := Migrate(mock); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Verify item_kinds table structure
	rows, err := db.Query(`PRAGMA table_info(item_kinds)`)
	if err != nil {
		t.Fatalf("failed to query item_kinds table info: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]string{
		"name":        "TEXT",
		"description": "TEXT",
		"llm_hint":    "TEXT",
		"builtin":     "INTEGER",
		"deprecated":  "INTEGER",
		"created_at":  "INTEGER",
		"created_by":  "TEXT",
	}

	foundColumns := make(map[string]string)
	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var dfltValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("failed to scan column info: %v", err)
		}
		foundColumns[name] = colType
	}

	for col, expectedType := range expectedColumns {
		if actualType, ok := foundColumns[col]; !ok {
			t.Errorf("item_kinds missing column %q", col)
		} else if actualType != expectedType {
			t.Errorf("item_kinds column %q has type %q, want %q", col, actualType, expectedType)
		}
	}

	// Verify tasks table has item_kind column with correct default
	rows, err = db.Query(`PRAGMA table_info(tasks)`)
	if err != nil {
		t.Fatalf("failed to query tasks table info: %v", err)
	}
	defer rows.Close()

	foundItemKind := false
	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var dfltValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("failed to scan column info: %v", err)
		}

		if name == "item_kind" {
			foundItemKind = true
			if colType != "TEXT" {
				t.Errorf("item_kind column type = %q, want TEXT", colType)
			}
			if !dfltValue.Valid || dfltValue.String != "'task'" {
				t.Errorf("item_kind default value = %q, want 'task'", dfltValue.String)
			}
		}
	}

	if !foundItemKind {
		t.Errorf("tasks table missing item_kind column")
	}
}
