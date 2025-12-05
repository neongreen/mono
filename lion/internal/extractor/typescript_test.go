package extractor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// PARITY TESTS: These tests mirror Go extraction tests to ensure TypeScript
// extraction has equivalent functionality.
// =============================================================================

// TestExtractTypeScript mirrors TestExtract for Go
// Tests basic extraction of topics and entities from TypeScript files.
func TestExtractTypeScript(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	// Create test TypeScript file - mirrors Go test structure
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
 * @lion api
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

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	// Verify topics (same as Go test)
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
	if len(introEntries) > 0 {
		if introEntries[0].Topic != "intro" {
			t.Errorf("expected topic 'intro', got %q", introEntries[0].Topic)
		}
	}

	// Verify api entries (2 entries like Go test)
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

// TestExtractTypeScriptNoComments mirrors TestExtractNoComments for Go
func TestExtractTypeScriptNoComments(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	// Create TypeScript file without lion comments
	testFile := filepath.Join(tmpDir, "empty.ts")
	content := `
function normalFunction() {}

interface Empty {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	if docs != nil && len(docs) != 0 {
		t.Errorf("expected no documentation, got %d topics", len(docs))
	}
}

// TestExtractTypeScriptMultipleFiles mirrors TestExtractMultipleFiles for Go
func TestExtractTypeScriptMultipleFiles(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	// Create first file
	file1 := filepath.Join(tmpDir, "file1.ts")
	content1 := `/**
 * @lion overview
 * Part one of the overview.
 */

/**
 * @lion api
 * Function in file1.
 */
function function1() {}
`

	// Create second file
	file2 := filepath.Join(tmpDir, "file2.ts")
	content2 := `/**
 * @lion overview
 * Part two of the overview.
 */

/**
 * @lion api
 * Function in file2.
 */
function function2() {}
`

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	// Should have 2 topics: overview and api
	if len(docs) != 2 {
		t.Errorf("expected 2 topics, got %d", len(docs))
	}

	// Overview should have 2 entries (one from each file)
	overviewEntries := docs["overview"]
	if len(overviewEntries) != 2 {
		t.Errorf("expected 2 overview entries, got %d", len(overviewEntries))
	}

	// API should have 2 entries (one from each file)
	apiEntries := docs["api"]
	if len(apiEntries) != 2 {
		t.Errorf("expected 2 api entries, got %d", len(apiEntries))
	}
}

func TestExtractTypeScriptNoFiles(t *testing.T) {
	// Skip if TypeScript helper is not built
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
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
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
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
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
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
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
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
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
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

// =============================================================================
// JSDOC FORMAT TESTS: Tests various JSDoc comment formats
// =============================================================================

// TestTypeScriptJSDocFormats tests various JSDoc comment formats
// Note: TypeScript only supports multi-line JSDoc comments (/** ... */), not single-line.
// This is a documented difference from Go extraction.
func TestTypeScriptJSDocFormats(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	// Test various JSDoc formats (all multi-line, as TypeScript requires)
	testFile := filepath.Join(tmpDir, "jsdoc.ts")
	content := `/**
 * @lion format1
 * Standard multi-line JSDoc with asterisk on each line.
 * This is the most common format.
 */
function format1() {}

/**
 * @lion format2
 * Another standard format
 */
function format2() {}

/**
 * @lion format3
 * JSDoc with blank lines
 *
 * And continuation after blank line.
 */
function format3() {}

/**
 * Mixed content with other JSDoc tags
 * @lion format4
 * The lion tag can appear after other content.
 * @param x - Some parameter (should be ignored in lion content)
 * @returns Something (should be ignored in lion content)
 */
function format4(x: number): number { return x; }

/**
 * @lion format5 Content on same line as tag
 * Followed by more content on subsequent lines.
 * Multiple lines should be captured.
 */
function format5() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	// Should have 5 different format topics
	expectedFormats := []string{"format1", "format2", "format3", "format4", "format5"}
	for _, format := range expectedFormats {
		if _, exists := docs[format]; !exists {
			t.Errorf("expected topic %q not found", format)
		}
	}

	// Verify format1 content (standard multi-line)
	if entries := docs["format1"]; len(entries) == 1 {
		content := entries[0].Content
		if !strings.Contains(content, "Standard multi-line") {
			t.Errorf("format1 missing expected content, got: %q", content)
		}
		if !strings.Contains(content, "most common format") {
			t.Errorf("format1 should include continuation, got: %q", content)
		}
	}

	// Verify format2 content
	if entries := docs["format2"]; len(entries) == 1 {
		content := entries[0].Content
		if !strings.Contains(content, "Another standard format") {
			t.Errorf("format2 missing expected content, got: %q", content)
		}
	}

	// Verify format5 content (content on tag line)
	if entries := docs["format5"]; len(entries) == 1 {
		content := entries[0].Content
		if !strings.Contains(content, "Content on same line") {
			t.Errorf("format5 missing expected content, got: %q", content)
		}
	}
}

// TestTypeScriptClassMembers tests extraction from class methods and properties
func TestTypeScriptClassMembers(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "class.ts")
	content := `/**
 * @lion api
 * The UserService class handles user operations.
 */
class UserService {
  /**
   * @lion api section="Properties"
   * The base URL for API calls.
   */
  private baseUrl: string = "";

  /**
   * @lion api section="Methods"
   * Creates a new user.
   */
  createUser(name: string) {
    return { name };
  }

  /**
   * @lion api section="Methods"
   * Deletes a user by ID.
   */
  deleteUser(id: string) {}
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
	if len(entries) < 3 {
		t.Errorf("expected at least 3 api entries (class + 2 methods), got %d", len(entries))
	}

	// Verify we have entries for class and members
	entities := make(map[string]bool)
	for _, entry := range entries {
		entities[entry.Entity] = true
	}

	expectedEntities := []string{"UserService", "createUser", "deleteUser"}
	for _, entity := range expectedEntities {
		if !entities[entity] {
			t.Errorf("expected entity %q not found", entity)
		}
	}
}

// =============================================================================
// MIXED LANGUAGE TESTS: Tests Go and TypeScript in the same project
// =============================================================================

// TestMixedProjectUnifiedOutput tests that mixed Go/TS projects produce unified docs
func TestMixedProjectUnifiedOutput(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	// Create subdirectories to simulate a real project structure
	goDir := filepath.Join(tmpDir, "server")
	tsDir := filepath.Join(tmpDir, "client")
	os.MkdirAll(goDir, 0755)
	os.MkdirAll(tsDir, 0755)

	// Go files in server/
	goFile := filepath.Join(goDir, "server.go")
	goContent := `//lion:architecture section="Server"
// The server handles HTTP requests and routing.
package server

//lion:api section="Server Config"
// ServerConfig holds server configuration.
type ServerConfig struct {
	Port int
}

//lion:api section="Server Functions"
// StartServer initializes and starts the HTTP server.
func StartServer() {}
`

	// TypeScript files in client/
	tsFile := filepath.Join(tsDir, "client.ts")
	tsContent := `/**
 * @lion architecture section="Client"
 * The client handles UI rendering and user interactions.
 */

/**
 * @lion api section="Client Config"
 * ClientConfig holds client-side configuration.
 */
interface ClientConfig {
  apiUrl: string;
}

/**
 * @lion api section="Client Functions"
 * initClient initializes the client application.
 */
function initClient() {}
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

	// Should have unified topics from both languages
	if len(docs) != 2 {
		t.Errorf("expected 2 topics (architecture, api), got %d", len(docs))
	}

	// Architecture topic should have entries from both Go and TypeScript
	archEntries := docs["architecture"]
	if len(archEntries) < 2 {
		t.Errorf("expected at least 2 architecture entries, got %d", len(archEntries))
	}

	// Verify both file types are present in architecture
	hasGo := false
	hasTS := false
	for _, entry := range archEntries {
		if strings.HasSuffix(entry.File, ".go") {
			hasGo = true
			if !strings.Contains(entry.Content, "server handles HTTP") {
				t.Errorf("Go architecture entry missing expected content")
			}
		}
		if strings.HasSuffix(entry.File, ".ts") {
			hasTS = true
			if !strings.Contains(entry.Content, "client handles UI") {
				t.Errorf("TypeScript architecture entry missing expected content")
			}
		}
	}
	if !hasGo {
		t.Error("expected Go file in architecture topic")
	}
	if !hasTS {
		t.Error("expected TypeScript file in architecture topic")
	}

	// API topic should have entries from both languages
	apiEntries := docs["api"]
	if len(apiEntries) < 4 {
		t.Errorf("expected at least 4 api entries (2 Go + 2 TS), got %d", len(apiEntries))
	}

	// Check sections are preserved
	sections := make(map[string]int)
	for _, entry := range apiEntries {
		if entry.HasSection {
			sections[entry.SectionTitle]++
		}
	}

	expectedSections := []string{"Server Config", "Server Functions", "Client Config", "Client Functions"}
	for _, sec := range expectedSections {
		if sections[sec] == 0 {
			t.Errorf("expected section %q not found", sec)
		}
	}
}

// TestMixedProjectWithSameTopicDifferentEntities tests combining entries from different languages
func TestMixedProjectWithSameTopicDifferentEntities(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	// Both files contribute to the same 'models' topic
	goFile := filepath.Join(tmpDir, "models.go")
	goContent := `//lion:models The User struct in Go
package models

//lion:models User represents a user in the system.
type User struct {
	ID   string
	Name string
}

//lion:models Product represents a product.
type Product struct {
	ID    string
	Price float64
}
`

	tsFile := filepath.Join(tmpDir, "models.ts")
	tsContent := `/**
 * @lion models
 * The Order interface in TypeScript
 */
interface Order {
  id: string;
  userId: string;
  products: string[];
}

/**
 * @lion models
 * The Cart interface for shopping cart.
 */
interface Cart {
  userId: string;
  items: string[];
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

	// Should have 1 unified 'models' topic
	if len(docs) != 1 {
		t.Errorf("expected 1 topic (models), got %d", len(docs))
	}

	modelEntries := docs["models"]
	// Go: 3 entries (package doc + User + Product)
	// TypeScript: 2 entries (Order + Cart)
	// Total: at least 4 entries (package doc may be combined)
	if len(modelEntries) < 4 {
		t.Errorf("expected at least 4 model entries, got %d", len(modelEntries))
	}

	// Verify entities from both languages
	entities := make(map[string]bool)
	for _, entry := range modelEntries {
		if entry.Entity != "" {
			entities[entry.Entity] = true
		}
	}

	// Check Go entities
	if !entities["User"] {
		t.Error("expected Go entity 'User' not found")
	}
	if !entities["Product"] {
		t.Error("expected Go entity 'Product' not found")
	}

	// Check TypeScript entities
	if !entities["Order"] {
		t.Error("expected TypeScript entity 'Order' not found")
	}
	if !entities["Cart"] {
		t.Error("expected TypeScript entity 'Cart' not found")
	}
}

// =============================================================================
// CONTENT VERIFICATION TESTS
// =============================================================================

// TestTypeScriptContentExtraction verifies content is properly extracted
func TestTypeScriptContentExtraction(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "content.ts")
	content := `/**
 * @lion content-test
 * This is the first line of content.
 * This is the second line.
 * 
 * This is after a blank line.
 * Final line of content.
 */
function testContent() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	entries := docs["content-test"]
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	content = entries[0].Content
	if !strings.Contains(content, "first line of content") {
		t.Errorf("missing first line, got: %q", content)
	}
	if !strings.Contains(content, "second line") {
		t.Errorf("missing second line, got: %q", content)
	}
	if !strings.Contains(content, "Final line") {
		t.Errorf("missing final line, got: %q", content)
	}
}

// TestTypeScriptFileAndLineTracking verifies file paths and line numbers are tracked
func TestTypeScriptFileAndLineTracking(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "tracking.ts")
	content := `/**
 * @lion tracking
 * First function.
 */
function first() {}

/**
 * @lion tracking
 * Second function.
 */
function second() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	entries := docs["tracking"]
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Check file paths
	for _, entry := range entries {
		if !strings.HasSuffix(entry.File, "tracking.ts") {
			t.Errorf("expected file to end with 'tracking.ts', got: %q", entry.File)
		}
	}

	// Check line numbers are different (second should be after first)
	if entries[0].Line >= entries[1].Line {
		t.Errorf("expected second entry to have higher line number, got: first=%d, second=%d",
			entries[0].Line, entries[1].Line)
	}
}

// =============================================================================
// INLINE COMMENT TESTS: Tests for inline comments inside function bodies
// =============================================================================

// TestTypeScriptInlineComments tests extraction of inline comments inside function bodies
// This mirrors Go's support for inline comments
func TestTypeScriptInlineComments(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "inline.ts")
	content := `/**
 * @lion api
 * Main function documentation
 */
function main() {
  // @lion implementation This is an inline comment using @lion syntax
  doSomething();
  
  //lion:implementation Another inline using lion: syntax
  doSomethingElse();
}

function doSomething() {}
function doSomethingElse() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	// Should have 2 topics: api and implementation
	if len(docs) != 2 {
		t.Errorf("expected 2 topics, got %d", len(docs))
	}

	// API topic should have 1 entry (the function JSDoc)
	apiEntries := docs["api"]
	if len(apiEntries) != 1 {
		t.Errorf("expected 1 api entry, got %d", len(apiEntries))
	}

	// Implementation topic should have 2 entries (inline comments)
	implEntries := docs["implementation"]
	if len(implEntries) != 2 {
		t.Errorf("expected 2 implementation entries, got %d", len(implEntries))
	}

	// Verify inline comment content and entity names
	for _, entry := range implEntries {
		// Inline comments should be associated with containing function
		if entry.Entity != "main" {
			t.Errorf("expected entity 'main', got %q", entry.Entity)
		}
	}

	// Check content of inline comments
	if len(implEntries) >= 1 {
		if !strings.Contains(implEntries[0].Content, "inline comment") {
			t.Errorf("first inline comment missing expected content, got: %q", implEntries[0].Content)
		}
	}
	if len(implEntries) >= 2 {
		if !strings.Contains(implEntries[1].Content, "Another inline") {
			t.Errorf("second inline comment missing expected content, got: %q", implEntries[1].Content)
		}
	}
}

// TestTypeScriptInlineCommentsWithMetadata tests inline comments with title/section metadata
func TestTypeScriptInlineCommentsWithMetadata(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc not available")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "inline-meta.ts")
	content := `function process() {
  // @lion steps section="Step 1" Initialize the system
  init();
  
  //lion:steps section="Step 2" Process the data
  processData();
}

function init() {}
function processData() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := ExtractTypeScript(tmpDir)
	if err != nil {
		t.Fatalf("ExtractTypeScript failed: %v", err)
	}

	entries := docs["steps"]
	if len(entries) != 2 {
		t.Fatalf("expected 2 steps entries, got %d", len(entries))
	}

	// Check sections are preserved
	sections := make(map[string]bool)
	for _, entry := range entries {
		if entry.HasSection {
			sections[entry.SectionTitle] = true
		}
	}

	if !sections["Step 1"] {
		t.Error("expected section 'Step 1' not found")
	}
	if !sections["Step 2"] {
		t.Error("expected section 'Step 2' not found")
	}
}
