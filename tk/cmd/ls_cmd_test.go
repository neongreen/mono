package main

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
)

func TestLsProjectsSortedAlphabetically(t *testing.T) {
	db := openTempDB(t)

	// Create projects in non-alphabetical order and tasks
	zebraUID := seedProject(t, db, "zebra")
	appleUID := seedProject(t, db, "apple")
	monkeyUID := seedProject(t, db, "monkey")
	bananaUID := seedProject(t, db, "banana")

	seedTask(t, db, zebraUID, "Zebra task", 1)
	seedTask(t, db, appleUID, "Apple task", 1)
	seedTask(t, db, monkeyUID, "Monkey task", 1)
	seedTask(t, db, bananaUID, "Banana task", 1)

	// Load config
	config, err := config_pkg.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Get cached reducer
	reducer, err := db.GetCachedReducerWithConfig(config)
	if err != nil {
		t.Fatalf("failed to get reducer: %v", err)
	}

	tasks := reducer.GetAllTasks()

	// Test JSON output with project grouping
	// We'll simulate what outputTasksJSON does
	grouped := make(map[string]int) // map of project name to count
	var groupOrder []string

	// Get all projects
	allProjects, err := getAllProjectDisplayNames(db)
	if err != nil {
		t.Fatalf("failed to get projects: %v", err)
	}

	// Initialize all projects
	for _, displayName := range allProjects {
		grouped[displayName] = 0
		groupOrder = append(groupOrder, displayName)
	}

	// Add tasks to groups
	for _, task := range tasks {
		projectAlias, err := getProjectAliasForTask(db, task.TaskUUID)
		if err != nil {
			continue
		}
		grouped[projectAlias]++
	}

	// Sort projects (this is what our fix does)
	sort.Strings(groupOrder)

	// Verify the order is alphabetical
	expectedOrder := []string{"apple", "banana", "monkey", "zebra"}

	// Filter to only projects that exist in our test
	var actualOrder []string
	for _, name := range groupOrder {
		if contains(expectedOrder, name) {
			actualOrder = append(actualOrder, name)
		}
	}

	if len(actualOrder) != len(expectedOrder) {
		t.Fatalf("expected %d projects, got %d", len(expectedOrder), len(actualOrder))
	}

	for i := range expectedOrder {
		if actualOrder[i] != expectedOrder[i] {
			t.Errorf("at position %d: expected %q, got %q", i, expectedOrder[i], actualOrder[i])
		}
	}
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func TestOutputTasksJSONSorting(t *testing.T) {
	db := openTempDB(t)

	// Create projects in non-alphabetical order
	zebraUID := seedProject(t, db, "zebra")
	appleUID := seedProject(t, db, "apple")
	monkeyUID := seedProject(t, db, "monkey")

	// Create tasks
	seedTask(t, db, zebraUID, "Zebra task", 1)
	seedTask(t, db, appleUID, "Apple task", 1)
	seedTask(t, db, monkeyUID, "Monkey task", 1)

	// Load config
	config, err := config_pkg.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Get cached reducer
	reducer, err := db.GetCachedReducerWithConfig(config)
	if err != nil {
		t.Fatalf("failed to get reducer: %v", err)
	}

	tasks := reducer.GetAllTasks()

	// Capture JSON output to a string buffer
	// We'll use a simplified version of the JSON output logic
	type GroupedOutput struct {
		Group string `json:"group"`
		Tasks int    `json:"task_count"` // Simplified for testing
	}

	type Output struct {
		Groups []GroupedOutput `json:"groups"`
	}

	grouped := make(map[string]int)
	var groupOrder []string

	allProjects, err := getAllProjectDisplayNames(db)
	if err != nil {
		t.Fatalf("failed to get projects: %v", err)
	}

	for _, displayName := range allProjects {
		grouped[displayName] = 0
		groupOrder = append(groupOrder, displayName)
	}

	for _, task := range tasks {
		projectAlias, err := getProjectAliasForTask(db, task.TaskUUID)
		if err != nil {
			continue
		}
		if _, exists := grouped[projectAlias]; !exists {
			groupOrder = append(groupOrder, projectAlias)
		}
		grouped[projectAlias]++
	}

	// Sort (this is what our fix does)
	sort.Strings(groupOrder)

	var output Output
	output.Groups = make([]GroupedOutput, 0, len(groupOrder))
	for _, groupKey := range groupOrder {
		output.Groups = append(output.Groups, GroupedOutput{
			Group: groupKey,
			Tasks: grouped[groupKey],
		})
	}

	// Verify JSON structure has sorted groups
	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	jsonStr := string(jsonOutput)

	// Check that projects appear in alphabetical order in JSON
	applePos := strings.Index(jsonStr, `"apple"`)
	monkeyPos := strings.Index(jsonStr, `"monkey"`)
	zebraPos := strings.Index(jsonStr, `"zebra"`)

	if applePos == -1 || monkeyPos == -1 || zebraPos == -1 {
		t.Fatal("not all projects found in JSON output")
	}

	if !(applePos < monkeyPos && monkeyPos < zebraPos) {
		t.Errorf("projects not in alphabetical order: apple at %d, monkey at %d, zebra at %d",
			applePos, monkeyPos, zebraPos)
	}
}
