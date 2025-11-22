package cmd

// seeAlsoRegistry maps command paths to related commands that should appear
// in the "See Also" section of their help text.
//
// This registry improves command discoverability, especially for AI agents
// who might not guess at less-obvious command names.
//
// IMPORTANT: When you add a new command, consider adding it here to help users
// discover it! Think about:
// - What commands naturally lead to using this new command?
// - What other commands would users want to know about after using this command?
// - Are there similar/alternative commands that should cross-reference each other?
//
// The registry is validated by tests in help_test.go - if you reference a command
// that doesn't exist, the tests will fail.
var seeAlsoRegistry = map[string][]string{
	// Core task management
	// Commands for creating, viewing, editing, and organizing tasks
	"new":      {"show", "mark", "edit", "describe", "ls"},
	"show":     {"edit", "describe", "note", "history", "relate ls", "attach"},
	"ls":       {"show", "mark", "new", "project ls", "blocked"},
	"mark":     {"show", "ls", "blockers"},
	"edit":     {"describe", "show", "note", "mv"},
	"describe": {"edit", "show", "note"},
	"note":     {"show", "describe", "history"},
	"attach":   {"show", "note"},
	"mv":       {"edit", "project ls", "show"},
	"rm":       {"show", "ls"},

	// Relations & dependencies
	// Commands for linking tasks together and understanding dependencies
	"relate":        {"relate add", "relate ls", "dup", "blockers", "graph"},
	"relate add":    {"relate ls", "relate remove", "dup", "graph"},
	"relate remove": {"relate ls", "relate add"},
	"relate ls":     {"relate add", "graph", "show", "blockers"},
	"dup":           {"relate add", "relate ls"},
	"blockers":      {"blocked", "relate add", "mark", "graph"},
	"blocked":       {"blockers", "mark", "relate ls"},
	"graph":         {"relate ls", "blockers", "show"},

	// Project management
	// Commands for organizing tasks into projects
	"project":        {"project create", "project ls", "project rename"},
	"project create": {"new", "mv", "project ls"},
	"project ls":     {"project create", "ls", "mv"},
	"project rm":     {"project ls", "mv"},
	"project rename": {"project ls"},

	// Container management - queues
	// FIFO containers for task organization
	"queue":        {"queue create", "queue push", "queue pop", "queue list", "stack", "group"},
	"queue create": {"queue push", "queue list", "new"},
	"queue push":   {"queue pop", "queue show", "ls"},
	"queue pop":    {"queue push", "queue show", "mark"},
	"queue list":   {"queue show", "queue create", "stack list", "group list"},
	"queue show":   {"queue push", "queue pop", "ls"},
	"queue rename": {"queue list"},
	"queue rm":     {"queue list"},

	// Container management - stacks
	// LIFO containers for task organization
	"stack":        {"stack create", "stack push", "stack pop", "stack list", "queue", "group"},
	"stack create": {"stack push", "stack list", "new"},
	"stack push":   {"stack pop", "stack show", "ls"},
	"stack pop":    {"stack push", "stack show", "mark"},
	"stack list":   {"stack show", "stack create", "queue list", "group list"},
	"stack show":   {"stack push", "stack pop", "ls"},
	"stack rename": {"stack list"},
	"stack rm":     {"stack list"},

	// Container management - groups
	// Unordered containers for task organization
	"group":        {"group create", "group add", "group list", "queue", "stack"},
	"group create": {"group add", "group list"},
	"group add":    {"group show", "group remove", "ls"},
	"group remove": {"group add", "group show"},
	"group list":   {"group show", "group create", "queue list", "stack list"},
	"group show":   {"group add", "ls"},
	"group rename": {"group list"},
	"group rm":     {"group list"},

	// Schema & metadata
	// Commands for custom fields and metadata
	"schema":        {"schema add", "schema list", "schema export", "meta"},
	"schema add":    {"schema list", "meta set"},
	"schema list":   {"schema add", "schema export", "meta list"},
	"schema export": {"schema list", "schema add"},
	"meta":          {"meta set", "meta get", "meta list", "schema"},
	"meta set":      {"meta get", "meta list", "show"},
	"meta get":      {"meta set", "meta list", "show"},
	"meta list":     {"meta get", "meta claims", "schema list"},
	"meta claims":   {"meta list", "meta get"},

	// Sync & remote
	// Commands for syncing tasks across machines
	"remote":      {"remote add", "remote ls", "push", "pull", "sync"},
	"remote add":  {"remote ls", "sync", "push"},
	"remote ls":   {"remote add", "remote rm", "status sync"},
	"remote rm":   {"remote ls"},
	"push":        {"pull", "sync", "remote ls", "status sync"},
	"pull":        {"push", "sync", "ingest"},
	"sync":        {"push", "pull", "status sync", "remote ls"},
	"ingest":      {"pull", "sync"},
	"status sync": {"sync", "remote ls", "push", "pull"},

	// Debugging & maintenance
	// Commands for troubleshooting and database maintenance
	"debug":              {"debug doctor", "debug repair", "debug rebuild"},
	"debug doctor":       {"debug repair", "conflicts"},
	"debug repair":       {"debug doctor", "debug rebuild"},
	"debug rebuild":      {"debug repair", "ingest"},
	"debug events":       {"debug events list", "debug events show", "debug events stats", "log query"},
	"debug events list":  {"debug events show", "debug events stats"},
	"debug events show":  {"debug events list", "history"},
	"debug events stats": {"debug events list"},
	"debug node":         {"debug node show", "debug node regen", "remote ls"},
	"debug node show":    {"debug node regen"},
	"debug node regen":   {"debug node show"},
	"conflicts":          {"conflicts numbers", "debug doctor", "edit"},
	"conflicts numbers":  {"edit"},
	"history":            {"show", "log query", "debug events show"},
	"log":                {"log query", "log search"},
	"log query":          {"log search", "debug events list", "history"},
	"log search":         {"log query"},

	// Migration & setup
	// Commands for initializing and migrating databases
	"init":                    {"new", "project create", "remote add"},
	"migrate":                 {"migrate scan-deprecated", "debug doctor"},
	"migrate scan-deprecated": {"debug doctor"},
	"version":                 {"debug doctor"},

	// Database
	// Commands for database operations
	"db":      {"db path"},
	"db path": {"init"},

	// Status
	// Commands for viewing system status
	"status":     {"status sync"},
	"statusline": {"status sync"},
}
