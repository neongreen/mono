// Package v4_to_v5 migrates the database from version 4 to version 5.
//
// # Changes in v5
//
// Schema changes:
//   - Adds is_synthetic column to projects table (INTEGER DEFAULT 0)
//
// Projection changes:
//   - ProjectTaskRelocateEvent now creates synthetic projects for unresolvable project references
//   - Synthetic projects use the literal corrupt value as project_uid (e.g., "abc")
//   - Synthetic projects are marked with is_synthetic=1
//
// # Background
//
// This migration addresses tk-281: task.relocate events were accepting project names
// instead of UUIDs. When these events were created with invalid project references
// (e.g., to_project_uid="abc" instead of "prj_..."), tasks became orphaned when
// the referenced project was deleted.
//
// The v5 migration makes this corrupt data visible and operable by creating "synthetic
// projects" during projection. These synthetic projects:
//   - Have no corresponding project.created event
//   - Use the literal corrupt value as their project_uid
//   - Are marked with is_synthetic=1 in the database
//   - Make orphaned tasks immediately accessible after upgrade
//
// # Design Principles (from tk-283)
//
//  1. Upgrades never require immediate action - data is immediately visible
//  2. Never withhold data from the user - corrupt state is operable
//  3. Weird state should be first-class - synthetic projects work like normal projects
//
// # Deprecation Timeline
//
// Synthetic projects are a supported long-term feature but will eventually be deprecated:
//   - v5-v9: Synthetic projects fully supported
//   - v9: Deprecation warnings added
//   - v10: Cleanup required before upgrade (estimated 2027+)
//
// # Known Edge Case
//
// If there were multiple deleted projects with the same name (e.g., two different
// "abc" projects at different times), corrupt events referencing both will
// collapse into a single synthetic project. We lose the distinction, but this is
// acceptable given the data is already corrupt.
//
// Users can manually fix this using 'tk migrate fix-relocate-bug' (to be implemented)
// or 'tk mv' to move tasks to correct projects.
//
// # Related Tasks
//
//   - tk-281: task.relocate events accept project names instead of UUIDs
//   - tk-282: tk debug doctor should validate event payload references
//   - tk-283: Document design principles for migrations and conflict resolution
package v4_to_v5
