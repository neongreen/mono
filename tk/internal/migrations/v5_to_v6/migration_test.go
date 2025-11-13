package v5_to_v6

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

	// Create initial schema (v5)
	_, err = db.Exec(`
		CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT);
		INSERT INTO metadata (key, value) VALUES ('db_version', '5');

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
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	mock := &mockDB{db: db, version: 5}

	// Run migration
	if err := Migrate(mock); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Verify container_kinds table was created
	var tableCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='container_kinds'
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("failed to check container_kinds table: %v", err)
	}
	if tableCount != 1 {
		t.Errorf("container_kinds table not created")
	}

	// Verify containers table was created
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='containers'
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("failed to check containers table: %v", err)
	}
	if tableCount != 1 {
		t.Errorf("containers table not created")
	}

	// Verify container_members table was created
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='container_members'
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("failed to check container_members table: %v", err)
	}
	if tableCount != 1 {
		t.Errorf("container_members table not created")
	}

	// Verify indexes were created
	var indexCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='index' AND name LIKE 'idx_container%'
	`).Scan(&indexCount)
	if err != nil {
		t.Fatalf("failed to check indexes: %v", err)
	}
	if indexCount != 4 {
		t.Errorf("expected 4 container indexes, got %d", indexCount)
	}

	// Verify version was updated
	if mock.version != 6 {
		t.Errorf("database version = %d, want 6", mock.version)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	// Create temp database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create initial schema (v5)
	_, err = db.Exec(`
		CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT);
		INSERT INTO metadata (key, value) VALUES ('db_version', '5');

		CREATE TABLE projects (
			project_uid TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL,
			is_synthetic INTEGER DEFAULT 0
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	mock := &mockDB{db: db, version: 5}

	// Run migration twice - should be idempotent
	if err := Migrate(mock); err != nil {
		t.Fatalf("first Migrate() failed: %v", err)
	}

	if err := Migrate(mock); err != nil {
		t.Fatalf("second Migrate() failed (not idempotent): %v", err)
	}

	// Verify still at version 6
	if mock.version != 6 {
		t.Errorf("database version after second migration = %d, want 6", mock.version)
	}
}

func TestMigrateTableSchema(t *testing.T) {
	// Create temp database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create minimal schema
	_, err = db.Exec(`
		CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT);
		INSERT INTO metadata (key, value) VALUES ('db_version', '5');
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	mock := &mockDB{db: db, version: 5}

	// Run migration
	if err := Migrate(mock); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Verify container_kinds table structure
	rows, err := db.Query(`PRAGMA table_info(container_kinds)`)
	if err != nil {
		t.Fatalf("failed to query container_kinds table info: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]string{
		"name":        "TEXT",
		"primitive":   "TEXT",
		"description": "TEXT",
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
			t.Errorf("container_kinds missing column %q", col)
		} else if actualType != expectedType {
			t.Errorf("container_kinds column %q has type %q, want %q", col, actualType, expectedType)
		}
	}

	// Verify container_members table has position column
	rows, err = db.Query(`PRAGMA table_info(container_members)`)
	if err != nil {
		t.Fatalf("failed to query container_members table info: %v", err)
	}
	defer rows.Close()

	foundPosition := false
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

		if name == "position" {
			foundPosition = true
			if colType != "INTEGER" {
				t.Errorf("position column type = %q, want INTEGER", colType)
			}
		}
	}

	if !foundPosition {
		t.Errorf("container_members table missing position column")
	}
}
