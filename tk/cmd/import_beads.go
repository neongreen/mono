package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

// BeadsDependency represents a dependency relationship in beads format
type BeadsDependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
}

// BeadsIssue represents an issue from beads JSONL format
type BeadsIssue struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       string            `json:"status"` // open, in_progress, closed
	Priority     int               `json:"priority"`
	Type         string            `json:"type"` // bug, feature, task, epic, chore
	Labels       []string          `json:"labels"`
	Assignee     string            `json:"assignee"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
	Dependencies []BeadsDependency `json:"dependencies"`
}

var importBeadsCmd = &cobra.Command{
	Use:   "import-beads [path]",
	Short: "Import issues from beads JSONL format",
	Long: `Import issues from a beads .beads/issues.jsonl file into tk.

This command reads a beads issues.jsonl file and converts each issue into
a tk task, preserving:
- Titles and descriptions
- Status (open, in_progress, closed)
- Priority (0-4) as metadata
- Labels as metadata
- Dependencies (blocks, parent-child, related, discovered-from)

Examples:
  tk import-beads .beads/issues.jsonl    # Import from specific file
  tk import-beads /path/to/project       # Import from project (auto-finds .beads/issues.jsonl)
  tk import-beads                        # Import from current directory's .beads/issues.jsonl
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName, _ := cmd.Flags().GetString("project")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Determine the path to the beads file
		var beadsPath string
		if len(args) == 0 {
			// Default to current directory
			beadsPath = ".beads/issues.jsonl"
		} else {
			arg := args[0]
			// Check if it's a file or directory
			info, err := os.Stat(arg)
			if err != nil {
				return fmt.Errorf("path not found: %w", err)
			}
			if info.IsDir() {
				beadsPath = filepath.Join(arg, ".beads", "issues.jsonl")
			} else {
				beadsPath = arg
			}
		}

		// Check if file exists
		if _, err := os.Stat(beadsPath); os.IsNotExist(err) {
			return fmt.Errorf("beads file not found: %s", beadsPath)
		}

		// Read and parse beads file
		issues, err := readBeadsFile(beadsPath)
		if err != nil {
			return fmt.Errorf("failed to read beads file: %w", err)
		}

		if len(issues) == 0 {
			fmt.Println("No issues found in beads file")
			return nil
		}

		fmt.Printf("Found %d issues in beads file\n", len(issues))

		if dryRun {
			fmt.Println("\nDry run mode - no changes will be made")
			for _, issue := range issues {
				fmt.Printf("  %s: %s (status: %s, type: %s)\n",
					issue.ID, issue.Title, issue.Status, issue.Type)
			}
			return nil
		}

		// Open database
		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Get or create project
		projectUID, err := getOrCreateProjectForImport(db, projectName)
		if err != nil {
			return fmt.Errorf("failed to get/create project: %w", err)
		}

		// Import each issue
		imported := 0
		skipped := 0
		issueMap := make(map[string]string) // beads ID -> tk task UID

		for _, issue := range issues {
			taskUID, err := importBeadsIssue(db, issue, projectUID)
			if err != nil {
				fmt.Printf("Warning: failed to import %s: %v\n", issue.ID, err)
				skipped++
				continue
			}
			issueMap[issue.ID] = taskUID
			imported++
		}

		fmt.Printf("\nImported %d issues (%d skipped)\n", imported, skipped)

		// Second pass: import relationships
		if imported > 0 {
			fmt.Println("\nImporting relationships...")
			relImported := 0
			for _, issue := range issues {
				if taskUID, ok := issueMap[issue.ID]; ok {
					count, err := importBeadsRelationships(db, issue, taskUID, issueMap)
					if err != nil {
						fmt.Printf("Warning: failed to import relationships for %s: %v\n", issue.ID, err)
					}
					relImported += count
				}
			}
			fmt.Printf("Imported %d relationships\n", relImported)
		}

		return nil
	},
}

// readBeadsFile reads and parses a beads JSONL file
func readBeadsFile(path string) ([]BeadsIssue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var issues []BeadsIssue
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		var issue BeadsIssue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		issues = append(issues, issue)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return issues, nil
}

// importBeadsIssue imports a single beads issue as a tk task
func importBeadsIssue(db *database.DB, issue BeadsIssue, projectUID string) (string, error) {
	// Get node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", err
	}

	// Generate task UID
	taskUID, err := utils.GenerateTaskUUID()
	if err != nil {
		return "", err
	}

	// Compute next task number (max + 1)
	var maxNumber int64
	err = db.Db.QueryRow(`
		SELECT COALESCE(MAX(number), 0) FROM task_numbers
		WHERE project_uid = ?
	`, projectUID).Scan(&maxNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get max number: %w", err)
	}
	proposedNumber := maxNumber + 1

	// Parse created_at time if available
	var createdAt time.Time
	if issue.CreatedAt != "" {
		createdAt, err = time.Parse(time.RFC3339, issue.CreatedAt)
		if err != nil {
			// Try other common formats
			createdAt, err = time.Parse("2006-01-02T15:04:05", issue.CreatedAt)
			if err != nil {
				createdAt = time.Now()
			}
		}
	} else {
		createdAt = time.Now()
	}

	// Get actor
	actor := "importer"
	if issue.Assignee != "" {
		actor = issue.Assignee
	}

	// Create task.created event
	payload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: proposedNumber,
		Title:          issue.Title,
		CreatedNode:    nodeID,
		CreatedBy:      actor,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return "", err
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return "", err
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: createdAt,
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return "", err
	}

	if err := db.ProjectTaskCreatedEvent(event); err != nil {
		return "", err
	}

	// Create task.number.set event
	numberPayload := types.TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     proposedNumber,
		Reason:     "imported",
	}

	numberPayloadJSON, err := json.Marshal(numberPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal number payload: %w", err)
	}

	numberEventID, err := database.GenerateEventID(db)
	if err != nil {
		return "", fmt.Errorf("failed to generate event ID: %w", err)
	}

	numberTS, err := db.GetNextLamportTS()
	if err != nil {
		return "", fmt.Errorf("failed to get next lamport timestamp: %w", err)
	}

	numberEvent := types.Event{
		ID:        numberEventID,
		TS:        numberTS,
		CreatedAt: createdAt,
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   numberPayloadJSON,
	}

	if err := db.InsertEvent(numberEvent); err != nil {
		return "", fmt.Errorf("failed to insert number event: %w", err)
	}

	if err := db.ProjectTaskNumberSetEvent(numberEvent); err != nil {
		return "", fmt.Errorf("failed to project task number: %w", err)
	}

	// Add description as a note if present
	if issue.Description != "" {
		notePayload := types.TaskNoteAddPayload{
			TaskUUID: taskUID,
			TaskID:   "", // Will be filled by the database
			Markdown: issue.Description,
		}

		notePayloadJSON, err := json.Marshal(notePayload)
		if err != nil {
			return "", err
		}

		noteEventID, err := database.GenerateEventID(db)
		if err != nil {
			return "", err
		}

		noteTS, err := db.GetNextLamportTS()
		if err != nil {
			return "", err
		}

		noteEvent := types.Event{
			ID:        noteEventID,
			TS:        noteTS,
			CreatedAt: createdAt,
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindTaskNoteAdd),
			Payload:   notePayloadJSON,
		}

		if err := db.InsertEvent(noteEvent); err != nil {
			return "", err
		}
	}

	// Set status if not open
	if issue.Status != "" && issue.Status != "open" {
		// Map beads status to tk status
		tkStatus := mapBeadsStatus(issue.Status)

		statusPayload := types.TaskStatusSetPayload{
			TaskUUID: taskUID,
			TaskID:   "", // Will be filled by the database
			Axis:     "generic",
			State:    tkStatus,
			Role:     "human",
		}

		statusPayloadJSON, err := json.Marshal(statusPayload)
		if err != nil {
			return "", err
		}

		statusEventID, err := database.GenerateEventID(db)
		if err != nil {
			return "", err
		}

		statusTS, err := db.GetNextLamportTS()
		if err != nil {
			return "", err
		}

		statusEvent := types.Event{
			ID:        statusEventID,
			TS:        statusTS,
			CreatedAt: createdAt,
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindTaskStatusSet),
			Payload:   statusPayloadJSON,
		}

		if err := db.InsertEvent(statusEvent); err != nil {
			return "", err
		}
	}

	// Import metadata: priority
	if issue.Priority >= 0 && issue.Priority <= 4 {
		if err := createMetadataEvent(db, taskUID, "priority", json.RawMessage(fmt.Sprintf("%d", issue.Priority)), actor, createdAt); err != nil {
			return "", fmt.Errorf("failed to create priority metadata: %w", err)
		}
	}

	// Import metadata: labels
	if len(issue.Labels) > 0 {
		labelsJSON, err := json.Marshal(issue.Labels)
		if err != nil {
			return "", fmt.Errorf("failed to marshal labels: %w", err)
		}
		if err := createMetadataEvent(db, taskUID, "labels", json.RawMessage(labelsJSON), actor, createdAt); err != nil {
			return "", fmt.Errorf("failed to create labels metadata: %w", err)
		}
	}

	return taskUID, nil
}

// createMetadataEvent creates a task.meta.set event
func createMetadataEvent(db *database.DB, taskUID string, key string, value json.RawMessage, actor string, createdAt time.Time) error {
	payload := types.TaskMetaSetPayload{
		TaskUUID: taskUID,
		TaskID:   "",
		Key:      key,
		Value:    value,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata payload: %w", err)
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: createdAt,
		Actor:     actor,
		Role:      "human", // Import acts with human authority
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert metadata event: %w", err)
	}

	return nil
}

// importBeadsRelationships imports relationships for a beads issue
func importBeadsRelationships(db *database.DB, issue BeadsIssue, taskUID string, issueMap map[string]string) (int, error) {
	count := 0

	if issue.Dependencies == nil {
		return 0, nil
	}

	// Process dependency array
	for _, dep := range issue.Dependencies {
		// Only process dependencies where this issue is the source
		if dep.IssueID != issue.ID {
			continue
		}

		// Look up the target task UID
		targetUID, ok := issueMap[dep.DependsOnID]
		if !ok {
			// Target issue not imported, skip
			continue
		}

		// Map beads relationship type to tk type
		var tkRelType string
		switch dep.Type {
		case "parent-child":
			tkRelType = "parent"
		case "blocks":
			tkRelType = "blocks"
		case "related":
			tkRelType = "related"
		case "discovered-from":
			tkRelType = "related" // Map discovered-from to related
		default:
			// Unknown type, skip
			continue
		}

		// Create the relationship
		if err := createRelation(db, taskUID, targetUID, tkRelType); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// createRelation creates a relationship between two tasks
func createRelation(db *database.DB, fromUID, toUID, relType string) error {
	payload := types.RelationAddPayload{
		Src:  fromUID,
		Type: relType,
		Dst:  toUID,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return err
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "importer",
		Role:      "human",
		Kind:      string(types.EventKindRelationAdd),
		Payload:   payloadJSON,
	}

	return db.InsertEvent(event)
}

// mapBeadsStatus maps beads status to tk status
func mapBeadsStatus(beadsStatus string) string {
	switch beadsStatus {
	case "in_progress":
		return "in-progress"
	case "closed":
		return "done"
	case "open":
		return "todo"
	default:
		return beadsStatus
	}
}

// getOrCreateProjectForImport gets an existing project or creates a new one
func getOrCreateProjectForImport(db *database.DB, projectName string) (string, error) {
	// Get current user and node
	actor, err := getCurrentUser()
	if err != nil {
		return "", err
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", err
	}

	// Try to resolve existing project
	ref := types.NewProjectRef(projectName)
	projectUID, err := database.ResolveProjectRef(db, ref)
	if err == nil {
		// Project exists
		return projectUID.String(), nil
	}

	// Create new project
	projectUID = types.NewProjectUID()

	payload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID.String(),
		Type:        "local",
		Name:        projectName,
		Description: "Imported from beads",
		CreatedBy:   actor,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return "", err
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return "", err
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return "", err
	}

	if err := db.ProjectProjectCreatedEvent(event); err != nil {
		return "", err
	}

	// Add alias for the project
	aliasPayload := types.ProjectAliasAddPayload{
		ProjectUID: projectUID.String(),
		Alias:      projectName,
		Node:       nodeID,
		AddedBy:    actor,
	}

	aliasPayloadJSON, err := json.Marshal(aliasPayload)
	if err != nil {
		return "", err
	}

	aliasEventID, err := database.GenerateEventID(db)
	if err != nil {
		return "", err
	}

	aliasTS, err := db.GetNextLamportTS()
	if err != nil {
		return "", err
	}

	aliasEvent := types.Event{
		ID:        aliasEventID,
		TS:        aliasTS,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindProjectAliasAdd),
		Payload:   aliasPayloadJSON,
	}

	if err := db.InsertEvent(aliasEvent); err != nil {
		return "", err
	}

	if err := db.ProjectProjectAliasAddEvent(aliasEvent); err != nil {
		return "", err
	}

	return projectUID.String(), nil
}

func init() {
	importBeadsCmd.Flags().StringP("project", "p", "beads-import", "Project to import issues into")
	importBeadsCmd.Flags().Bool("dry-run", false, "Preview import without making changes")
}
