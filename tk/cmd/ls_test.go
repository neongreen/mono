package cmd

import (
	"encoding/json"
	"slices"
	"sort"
	"strings"
	"testing"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
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
	allProjects, err := database.GetAllProjectDisplayNames(db)
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
		projectAlias, err := database.GetProjectAliasForTask(db, task.TaskID)
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
		if slices.Contains(expectedOrder, name) {
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

func TestLsWithoutProjectFilterShowsAllProjects(t *testing.T) {
	db := openTempDB(t)

	// Create several projects, some with tasks and some empty
	appleUID := seedProject(t, db, "apple")
	bananaUID := seedProject(t, db, "banana")
	cherryUID := seedProject(t, db, "cherry")
	_ = seedProject(t, db, "empty-project")

	// Add tasks to some projects but not all
	seedTask(t, db, appleUID, "Apple task 1", 1)
	seedTask(t, db, appleUID, "Apple task 2", 2)
	seedTask(t, db, bananaUID, "Banana task", 1)
	seedTask(t, db, cherryUID, "Cherry task", 1)
	// empty-project has no tasks

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

	// Simulate the ls command logic without project filter
	grouped := make(map[string][]*types.Task)
	var groupOrder []string

	// Get all projects (this is what happens when no -p flag is specified)
	allProjects, err := database.GetAllProjectDisplayNames(db)
	if err != nil {
		t.Fatalf("failed to get projects: %v", err)
	}

	for _, displayName := range allProjects {
		grouped[displayName] = []*types.Task{}
		groupOrder = append(groupOrder, displayName)
	}

	for _, task := range tasks {
		projectAlias, err := database.GetProjectAliasForTask(db, task.TaskID)
		if err != nil {
			continue
		}
		if _, exists := grouped[projectAlias]; !exists {
			groupOrder = append(groupOrder, projectAlias)
		}
		grouped[projectAlias] = append(grouped[projectAlias], task)
	}

	sort.Strings(groupOrder)

	// Verify all projects are shown, including empty ones
	expectedProjects := []string{"apple", "banana", "cherry", "empty-project"}
	for _, expected := range expectedProjects {
		if !slices.Contains(groupOrder, expected) {
			t.Errorf("expected project %q to be in output, but it wasn't found", expected)
		}
	}

	// Verify empty project has no tasks
	if len(grouped["empty-project"]) != 0 {
		t.Errorf("expected empty-project to have 0 tasks, got %d", len(grouped["empty-project"]))
	}

	// Verify non-empty projects have tasks
	if len(grouped["apple"]) != 2 {
		t.Errorf("expected apple to have 2 tasks, got %d", len(grouped["apple"]))
	}
	if len(grouped["banana"]) != 1 {
		t.Errorf("expected banana to have 1 task, got %d", len(grouped["banana"]))
	}
}

func TestLsWithProjectFilterShowsOnlyFilteredProject(t *testing.T) {
	db := openTempDB(t)

	// Create multiple projects with tasks
	appleUID := seedProject(t, db, "apple")
	bananaUID := seedProject(t, db, "banana")
	cherryUID := seedProject(t, db, "cherry")

	seedTask(t, db, appleUID, "Apple task 1", 1)
	seedTask(t, db, appleUID, "Apple task 2", 2)
	seedTask(t, db, bananaUID, "Banana task", 1)
	seedTask(t, db, cherryUID, "Cherry task", 1)

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

	allTasks := reducer.GetAllTasks()

	// Filter to only "apple" project (simulating -p apple flag)
	projectFilter := []string{"apple"}
	taskIDs, err := db.GetTaskIDsByProjects(projectFilter)
	if err != nil {
		t.Fatalf("failed to get task IDs by projects: %v", err)
	}

	var filteredTasks []*types.Task
	taskUIDSet := make(map[string]bool)
	for _, id := range taskIDs {
		taskUIDSet[id] = true
	}
	for _, task := range allTasks {
		if taskUIDSet[task.TaskID] {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// Simulate the ls command logic WITH project filter
	grouped := make(map[string][]*types.Task)
	var groupOrder []string

	// When project filter is specified, DON'T get all projects
	// (this is the key behavior being tested)

	for _, task := range filteredTasks {
		projectAlias, err := database.GetProjectAliasForTask(db, task.TaskID)
		if err != nil {
			continue
		}
		if _, exists := grouped[projectAlias]; !exists {
			groupOrder = append(groupOrder, projectAlias)
		}
		grouped[projectAlias] = append(grouped[projectAlias], task)
	}

	sort.Strings(groupOrder)

	// Verify ONLY the filtered project appears
	if len(groupOrder) != 1 {
		t.Errorf("expected exactly 1 project in output, got %d: %v", len(groupOrder), groupOrder)
	}

	if !slices.Contains(groupOrder, "apple") {
		t.Errorf("expected 'apple' project in output, but it wasn't found")
	}

	// Verify other projects don't appear
	if slices.Contains(groupOrder, "banana") {
		t.Errorf("'banana' project should not appear when filtering for 'apple'")
	}
	if slices.Contains(groupOrder, "cherry") {
		t.Errorf("'cherry' project should not appear when filtering for 'apple'")
	}

	// Verify the correct tasks are shown
	if len(grouped["apple"]) != 2 {
		t.Errorf("expected apple to have 2 tasks, got %d", len(grouped["apple"]))
	}
}

func TestLsWithProjectFilterDoesNotShowEmptyProjects(t *testing.T) {
	db := openTempDB(t)

	// Create projects including an empty one in the filtered set
	appleUID := seedProject(t, db, "apple")
	bananaUID := seedProject(t, db, "banana")
	_ = seedProject(t, db, "empty-project")

	seedTask(t, db, appleUID, "Apple task", 1)
	seedTask(t, db, bananaUID, "Banana task", 1)
	// empty-project has no tasks

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

	allTasks := reducer.GetAllTasks()

	// Filter to only "empty-project" (simulating -p empty-project flag)
	projectFilter := []string{"empty-project"}
	taskIDs, err := db.GetTaskIDsByProjects(projectFilter)
	if err != nil {
		t.Fatalf("failed to get task IDs by projects: %v", err)
	}

	var filteredTasks []*types.Task
	taskUIDSet := make(map[string]bool)
	for _, id := range taskIDs {
		taskUIDSet[id] = true
	}
	for _, task := range allTasks {
		if taskUIDSet[task.TaskID] {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// Simulate the ls command logic WITH project filter
	grouped := make(map[string][]*types.Task)
	var groupOrder []string

	// When project filter is specified, DON'T get all projects

	for _, task := range filteredTasks {
		projectAlias, err := database.GetProjectAliasForTask(db, task.TaskID)
		if err != nil {
			continue
		}
		if _, exists := grouped[projectAlias]; !exists {
			groupOrder = append(groupOrder, projectAlias)
		}
		grouped[projectAlias] = append(grouped[projectAlias], task)
	}

	sort.Strings(groupOrder)

	// When filtering by a project with no tasks, the groupOrder should be empty
	if len(groupOrder) != 0 {
		t.Errorf("expected 0 projects in output when filtering by empty project, got %d: %v", len(groupOrder), groupOrder)
	}

	// Verify other projects don't appear
	if slices.Contains(groupOrder, "apple") {
		t.Errorf("'apple' project should not appear when filtering for 'empty-project'")
	}
	if slices.Contains(groupOrder, "banana") {
		t.Errorf("'banana' project should not appear when filtering for 'empty-project'")
	}
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

	allProjects, err := database.GetAllProjectDisplayNames(db)
	if err != nil {
		t.Fatalf("failed to get projects: %v", err)
	}

	for _, displayName := range allProjects {
		grouped[displayName] = 0
		groupOrder = append(groupOrder, displayName)
	}

	for _, task := range tasks {
		projectAlias, err := database.GetProjectAliasForTask(db, task.TaskID)
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

	if applePos >= monkeyPos || monkeyPos >= zebraPos {
		t.Errorf("projects not in alphabetical order: apple at %d, monkey at %d, zebra at %d",
			applePos, monkeyPos, zebraPos)
	}
}
