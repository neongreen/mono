package v6_to_v7

import (
	"database/sql"
	"fmt"
)

// DB interface defines the minimal database operations needed for migration.
// This avoids an import cycle with the database package.
type DB interface {
	Exec(query string, args ...any) (sql.Result, error)
	SetDBVersion(version int) error
}

// Migrate runs the v6 to v7 migration.
//
// This migration:
//  1. Creates item_kinds table
//  2. Adds builtin item kinds (task, bug, idea, goal, decision, requirement, constraint, wish, question, hypothesis, experiment, observation, research, doubt, assumption, resource, specification, definition, techdebt, checklist, discussion, feedback)
//  3. Adds item_kind column to tasks table (defaults to 'task')
//  4. Creates index on item_kind for performance
//  5. Updates db_version to 7
//
// This adds support for custom item kinds that can be defined at runtime via events.
//
// This migration is safe to run multiple times (idempotent).
func Migrate(db DB) error {
	// Create item_kinds table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS item_kinds (
			name TEXT PRIMARY KEY,
			description TEXT,
			llm_hint TEXT,
			builtin INTEGER NOT NULL DEFAULT 0,
			deprecated INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create item_kinds table: %w", err)
	}

	// Insert builtin item kinds
	builtinKinds := []struct {
		name        string
		description string
		llmHint     string
	}{
		{"task", "Actionable work to be done", "Use for concrete work items that need to be completed"},
		{"bug", "Defect or error to fix", "Use for software bugs, errors, or defects that need fixing"},
		{"idea", "Unrefined concept to explore", "Use for raw ideas that haven't been developed into actionable items yet"},
		{"goal", "Desired outcome to achieve", "Use for high-level objectives or targets you want to reach"},
		{"decision", "Choice to be made", "Use for decisions that need to be made or have been made"},
		{"requirement", "Must-have feature or capability", "Use for definite requirements that must be satisfied"},
		{"constraint", "Limitation or boundary condition", "Use for restrictions, limits, or conditions that must be respected"},
		{"wish", "Nice-to-have desire", "Use for aspirational features or improvements that aren't requirements"},
		{"question", "Something needing an answer", "Use for questions that need to be answered or clarified"},
		{"hypothesis", "Testable proposition", "Use for theories or assumptions that need to be validated through testing"},
		{"experiment", "Test to validate hypothesis", "Use for experiments or trials designed to test hypotheses"},
		{"observation", "Recorded fact or finding", "Use for observations, data points, or findings from experiments"},
		{"research", "Investigation topic", "Use for topics that need investigation or information gathering"},
		{"doubt", "Uncertainty to resolve", "Use for concerns, uncertainties, or risks that need addressing"},
		{"assumption", "Accepted premise", "Use for assumptions that underlie decisions or plans"},
		{"resource", "Reference or material", "Use for documentation, tools, links, or materials to reference"},
		{"specification", "Detailed technical definition", "Use for detailed technical specs or implementation details"},
		{"definition", "Meaning or explanation", "Use for defining terms, concepts, or establishing shared understanding"},
		{"techdebt", "Technical debt to address", "Use for technical debt, code quality issues, or refactoring needs"},
		{"checklist", "List of items to complete", "Use for checklists or sequences of steps to follow"},
		{"discussion", "Topic for conversation", "Use for topics that need discussion or conversation among stakeholders"},
		{"feedback", "Input or critique received", "Use for feedback, comments, or suggestions from users or stakeholders"},
	}

	for _, kind := range builtinKinds {
		_, err = db.Exec(`
			INSERT OR IGNORE INTO item_kinds (name, description, llm_hint, builtin, deprecated, created_at, created_by)
			VALUES (?, ?, ?, 1, 0, 0, 'system')
		`, kind.name, kind.description, kind.llmHint)
		if err != nil {
			return fmt.Errorf("failed to insert builtin %s kind: %w", kind.name, err)
		}
	}

	// Add item_kind column to tasks table (defaults to 'task')
	// This will fail gracefully if column already exists (idempotent)
	_, err = db.Exec(`
		ALTER TABLE tasks ADD COLUMN item_kind TEXT NOT NULL DEFAULT 'task'
	`)
	if err != nil {
		// SQLite returns error if column already exists - this is OK for idempotent migration
		// We just continue - the important thing is the column exists after this runs
	}

	// Create index on item_kind for performance
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tasks_item_kind ON tasks(item_kind)
	`)
	if err != nil {
		return fmt.Errorf("failed to create item_kind index: %w", err)
	}

	// Update database version
	if err := db.SetDBVersion(7); err != nil {
		return fmt.Errorf("failed to set DB version: %w", err)
	}

	return nil
}
