package extractor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractTypeScript(t *testing.T) {
	// Skip if TypeScript helper is not built
	if _, err := findTSHelper(); err != nil {
		t.Skip("TypeScript helper not built - run 'npm install && npm run build' in lion/ts-helper")
	}

	tmpDir := t.TempDir()

	// Create test TypeScript file
	testFile := filepath.Join(tmpDir, "test.ts")
	content := `/**
 * @lion intro
 * This is the module documentation.
 * It describes what the module does.
 */

/**
 * @lion api
 * The Config interface holds settings.
 */
interface Config {
  port: number;
}

/**
 * @lion api section="Initialize"
 * Initialize creates a new Config.
 */
function initialize(): Config {
  return { port: 8080 };
}

/**
 * @lion implementation
 * Internal helper function.
 */
function helper() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Extract documentation
	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	// Verify topics
	expectedTopics := []string{"intro", "api", "implementation"}
	if len(docs) != len(expectedTopics) {
		t.Errorf("expected %d topics, got %d", len(expectedTopics), len(docs))
	}

	for _, topic := range expectedTopics {
		if _, exists := docs[topic]; !exists {
			t.Errorf("expected topic %q not found", topic)
		}
	}

	// Verify intro entries
	introEntries := docs["intro"]
	if len(introEntries) != 1 {
		t.Errorf("expected 1 intro entry, got %d", len(introEntries))
	}

	// Verify api entries
	apiEntries := docs["api"]
	if len(apiEntries) != 2 {
		t.Errorf("expected 2 api entries, got %d", len(apiEntries))
	}

	// Check entities
	if len(apiEntries) >= 2 {
		if apiEntries[0].Entity != "Config" {
			t.Errorf("expected entity 'Config', got %q", apiEntries[0].Entity)
		}
		if apiEntries[1].Entity != "initialize" {
			t.Errorf("expected entity 'initialize', got %q", apiEntries[1].Entity)
		}
	}

	// Verify implementation entries
	implEntries := docs["implementation"]
	if len(implEntries) != 1 {
		t.Errorf("expected 1 implementation entry, got %d", len(implEntries))
	}
}

func TestExtractTypeScriptNoFiles(t *testing.T) {
	// Skip if TypeScript helper is not built
	if _, err := findTSHelper(); err != nil {
		t.Skip("TypeScript helper not built - run 'npm install && npm run build' in lion/ts-helper")
	}

	tmpDir := t.TempDir()

	// Create a Go file only
	goFile := filepath.Join(tmpDir, "main.go")
	content := `package main
func main() {}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create Go file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	if docs != nil && len(docs) > 0 {
		t.Errorf("expected no TypeScript docs, got %d topics", len(docs))
	}
}

func TestExtractTypeScriptWithMetadata(t *testing.T) {
	// Skip if TypeScript helper is not built
	if _, err := findTSHelper(); err != nil {
		t.Skip("TypeScript helper not built - run 'npm install && npm run build' in lion/ts-helper")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "meta.ts")
	content := `/**
 * @lion api title="API Reference" section="User Interface"
 * The User interface defines user data.
 */
interface User {
  name: string;
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	entries := docs["api"]
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if !entry.HasTopicTitle || entry.TopicTitle != "API Reference" {
		t.Errorf("expected topic title 'API Reference', got %q (has=%v)", entry.TopicTitle, entry.HasTopicTitle)
	}
	if !entry.HasSection || entry.SectionTitle != "User Interface" {
		t.Errorf("expected section 'User Interface', got %q (has=%v)", entry.SectionTitle, entry.HasSection)
	}
}

func TestExtractTypeScriptSkipsTestFiles(t *testing.T) {
	// Skip if TypeScript helper is not built
	if _, err := findTSHelper(); err != nil {
		t.Skip("TypeScript helper not built - run 'npm install && npm run build' in lion/ts-helper")
	}

	tmpDir := t.TempDir()

	// Create test file (should be skipped)
	testFile := filepath.Join(tmpDir, "example.test.ts")
	content := `/**
 * @lion testing
 * This should be skipped.
 */
function testSomething() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create spec file (should be skipped)
	specFile := filepath.Join(tmpDir, "example.spec.ts")
	specContent := `/**
 * @lion testing
 * This should also be skipped.
 */
function specSomething() {}
`

	if err := os.WriteFile(specFile, []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to create spec file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	if docs != nil && len(docs) > 0 {
		t.Errorf("expected no documentation (test files should be skipped), got %d topics", len(docs))
	}
}

func TestExtractMixedGoAndTypeScript(t *testing.T) {
	// Skip if TypeScript helper is not built
	if _, err := findTSHelper(); err != nil {
		t.Skip("TypeScript helper not built - run 'npm install && npm run build' in lion/ts-helper")
	}

	tmpDir := t.TempDir()

	// Create Go file
	goFile := filepath.Join(tmpDir, "main.go")
	goContent := `//lion:shared Go implementation
package main

//lion:api The Config struct holds settings.
type Config struct {
	Port int
}
`

	// Create TypeScript file
	tsFile := filepath.Join(tmpDir, "main.ts")
	tsContent := `/**
 * @lion shared
 * TypeScript implementation
 */

/**
 * @lion api
 * The Config interface holds settings.
 */
interface Config {
  port: number;
}
`

	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatalf("failed to create Go file: %v", err)
	}
	if err := os.WriteFile(tsFile, []byte(tsContent), 0644); err != nil {
		t.Fatalf("failed to create TypeScript file: %v", err)
	}

	docs, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should have 2 topics
	if len(docs) != 2 {
		t.Errorf("expected 2 topics, got %d", len(docs))
	}

	// shared topic should have 2 entries (one from each file)
	sharedEntries := docs["shared"]
	if len(sharedEntries) != 2 {
		t.Errorf("expected 2 shared entries, got %d", len(sharedEntries))
	}

	// api topic should have 2 entries (one from each file)
	apiEntries := docs["api"]
	if len(apiEntries) != 2 {
		t.Errorf("expected 2 api entries, got %d", len(apiEntries))
	}

	// Check that we have both Go and TypeScript files
	goFound := false
	tsFound := false
	for _, entry := range sharedEntries {
		if strings.HasSuffix(entry.File, ".go") {
			goFound = true
		}
		if strings.HasSuffix(entry.File, ".ts") {
			tsFound = true
		}
	}
	if !goFound {
		t.Error("expected Go file entry in shared topic")
	}
	if !tsFound {
		t.Error("expected TypeScript file entry in shared topic")
	}
}

func TestExtractTypeScriptEntities(t *testing.T) {
	// Skip if TypeScript helper is not built
	if _, err := findTSHelper(); err != nil {
		t.Skip("TypeScript helper not built - run 'npm install && npm run build' in lion/ts-helper")
	}

	tmpDir := t.TempDir()

	// Create TypeScript file with various entity types
	testFile := filepath.Join(tmpDir, "entities.ts")
	content := `/**
 * @lion entities
 * Interface declaration
 */
interface User {
  name: string;
}

/**
 * @lion entities
 * Type alias
 */
type UserId = string;

/**
 * @lion entities
 * Class declaration
 */
class UserService {
  constructor() {}
}

/**
 * @lion entities
 * Regular function
 */
function createUser(name: string) {
  return { name };
}

/**
 * @lion entities
 * Const variable
 */
const API_URL = "https://api.example.com";

/**
 * @lion entities
 * Enum declaration
 */
enum Status {
  Active,
  Inactive
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	entries := docs["entities"]
	if len(entries) != 6 {
		t.Errorf("expected 6 entity entries, got %d", len(entries))
	}

	// Verify entity names are extracted correctly
	expectedEntities := map[string]bool{
		"User":        false,
		"UserId":      false,
		"UserService": false,
		"createUser":  false,
		"API_URL":     false,
		"Status":      false,
	}

	for _, entry := range entries {
		if entry.Entity != "" {
			if _, exists := expectedEntities[entry.Entity]; exists {
				expectedEntities[entry.Entity] = true
			}
		}
	}

	for entity, found := range expectedEntities {
		if !found {
			t.Errorf("expected entity %q was not found", entity)
		}
	}
}
