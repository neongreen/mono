package v4_to_v5

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

	// Create initial schema (v4)
	_, err = db.Exec(`
		CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT);
		INSERT INTO metadata (key, value) VALUES ('db_version', '4');

		CREATE TABLE projects (
			project_uid TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	mock := &mockDB{db: db, version: 4}

	// Run migration
	if err := Migrate(mock); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Verify is_synthetic column was added
	rows, err := db.Query(`PRAGMA table_info(projects)`)
	if err != nil {
		t.Fatalf("failed to query table info: %v", err)
	}
	defer rows.Close()

	foundSyntheticColumn := false
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

		if name == "is_synthetic" {
			foundSyntheticColumn = true
			if colType != "INTEGER" {
				t.Errorf("is_synthetic column type = %q, want INTEGER", colType)
			}
		}
	}

	if !foundSyntheticColumn {
		t.Errorf("is_synthetic column not found in projects table")
	}

	// Verify version was updated
	if mock.version != 5 {
		t.Errorf("database version = %d, want 5", mock.version)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	// Create temp database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create initial schema (v4)
	_, err = db.Exec(`
		CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT);
		INSERT INTO metadata (key, value) VALUES ('db_version', '4');

		CREATE TABLE projects (
			project_uid TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	mock := &mockDB{db: db, version: 4}

	// Run migration twice - should be idempotent
	if err := Migrate(mock); err != nil {
		t.Fatalf("first Migrate() failed: %v", err)
	}

	if err := Migrate(mock); err != nil {
		t.Fatalf("second Migrate() failed (not idempotent): %v", err)
	}

	// Verify still at version 5
	if mock.version != 5 {
		t.Errorf("database version after second migration = %d, want 5", mock.version)
	}
}
