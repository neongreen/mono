package extractor

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create test Go file
	testFile := filepath.Join(tmpDir, "test.go")
	content := `//lion:intro This is the package documentation.
//lion:intro It describes what the package does.
package test

//lion:api The Config struct holds settings.
type Config struct {
	Port int
}

//lion:api Initialize creates a new Config.
func Initialize() *Config {
	return &Config{Port: 8080}
}

//lion:implementation Internal helper function.
func helper() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Extract documentation
	docs, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
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
	if len(introEntries) > 0 {
		entry := introEntries[0]
		if entry.Entity != "package test" {
			t.Errorf("expected entity 'package test', got %q", entry.Entity)
		}
		if entry.Topic != "intro" {
			t.Errorf("expected topic 'intro', got %q", entry.Topic)
		}
	}

	// Verify api entries
	apiEntries := docs["api"]
	if len(apiEntries) != 2 {
		t.Errorf("expected 2 api entries, got %d", len(apiEntries))
	}

	// Verify implementation entries
	implEntries := docs["implementation"]
	if len(implEntries) != 1 {
		t.Errorf("expected 1 implementation entry, got %d", len(implEntries))
	}
}

func TestExtractNoComments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go file without lion comments
	testFile := filepath.Join(tmpDir, "empty.go")
	content := `package empty

func Normal() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(docs) != 0 {
		t.Errorf("expected no documentation, got %d topics", len(docs))
	}
}

func TestExtractMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first file
	file1 := filepath.Join(tmpDir, "file1.go")
	content1 := `//lion:overview Part one of the overview.
package test

//lion:api Function in file1.
func Function1() {}
`

	// Create second file
	file2 := filepath.Join(tmpDir, "file2.go")
	content2 := `package test

//lion:overview Part two of the overview.
//lion:api Function in file2.
func Function2() {}
`

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	docs, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
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

func TestExtractSkipsTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file (should be skipped)
	testFile := filepath.Join(tmpDir, "example_test.go")
	content := `package test

//lion:testing This should be skipped.
func TestSomething() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(docs) != 0 {
		t.Errorf("expected no documentation (test files should be skipped), got %d topics", len(docs))
	}
}

func TestParseLionCommentLine(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedTopic   string
		expectedContent string
	}{
		{
			name:            "topic with content",
			input:           "lion:api This is the content",
			expectedTopic:   "api",
			expectedContent: "This is the content",
		},
		{
			name:            "topic without content",
			input:           "lion:overview",
			expectedTopic:   "overview",
			expectedContent: "",
		},
		{
			name:            "topic with hyphen",
			input:           "lion:getting-started Welcome message",
			expectedTopic:   "getting-started",
			expectedContent: "Welcome message",
		},
		{
			name:            "empty after lion:",
			input:           "lion:",
			expectedTopic:   "",
			expectedContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic, content, _, err := parseLionCommentLine(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if topic != tt.expectedTopic {
				t.Errorf("expected topic %q, got %q", tt.expectedTopic, topic)
			}
			if content != tt.expectedContent {
				t.Errorf("expected content %q, got %q", tt.expectedContent, content)
			}
		})
	}
}

func TestParseLionBlockComment(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedTopic   string
		expectedContent string
	}{
		{
			name:            "single line block",
			input:           "lion:api This is a single line block comment",
			expectedTopic:   "api",
			expectedContent: "This is a single line block comment",
		},
		{
			name: "multi-line block",
			input: `lion:architecture
This is the first line
This is the second line
And a third line`,
			expectedTopic:   "architecture",
			expectedContent: "This is the first line\nThis is the second line\nAnd a third line",
		},
		{
			name: "multi-line with content on first line",
			input: `lion:cli The root command
provides the main entry point
for the lion CLI`,
			expectedTopic:   "cli",
			expectedContent: "The root command\nprovides the main entry point\nfor the lion CLI",
		},
		{
			name:            "topic only",
			input:           "lion:overview",
			expectedTopic:   "overview",
			expectedContent: "",
		},
		{
			name: "multi-line with empty lines",
			input: `lion:implementation
First paragraph

Second paragraph with gap`,
			expectedTopic:   "implementation",
			expectedContent: "First paragraph\nSecond paragraph with gap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic, content, _, err := parseLionBlockComment(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if topic != tt.expectedTopic {
				t.Errorf("expected topic %q, got %q", tt.expectedTopic, topic)
			}
			if content != tt.expectedContent {
				t.Errorf("expected content %q, got %q", tt.expectedContent, content)
			}
		})
	}
}

func TestParseLionCommentLineWithMetadata(t *testing.T) {
	topic, content, meta, err := parseLionCommentLine(`lion:topic title="Custom Title" section="Section Heading" Remaining content`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topic != "topic" {
		t.Fatalf("expected topic %q, got %q", "topic", topic)
	}
	if !meta.hasTopicTitle || meta.topicTitle != "Custom Title" {
		t.Fatalf("expected topic title %q, got %+v", "Custom Title", meta)
	}
	if !meta.hasSection || meta.sectionTitle != "Section Heading" {
		t.Fatalf("expected section title %q, got %+v", "Section Heading", meta)
	}
	if content != "Remaining content" {
		t.Fatalf("expected content %q, got %q", "Remaining content", content)
	}
}

func TestParseLionBlockCommentEntrySuppression(t *testing.T) {
	topic, content, meta, err := parseLionBlockComment(`lion:topic section=""
Body line`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topic != "topic" {
		t.Fatalf("expected topic %q, got %q", "topic", topic)
	}
	if !meta.hasSection || meta.sectionTitle != "" {
		t.Fatalf("expected empty section title suppression, got %+v", meta)
	}
	if content != "Body line" {
		t.Fatalf("expected content %q, got %q", "Body line", content)
	}
}

func TestExtractBlockComments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test Go file with block comments
	testFile := filepath.Join(tmpDir, "test.go")
	content := `/*lion:overview
This is a multi-line overview.
It describes what the package does.
No need to repeat lion:overview on each line.
*/
package test

/*lion:architecture
The main function initializes the application.
It handles setup and configuration.
*/
func main() {}

/*lion:api
The Config struct holds settings.
Fields can be loaded from various sources.
*/
type Config struct {
	Port int
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Extract documentation
	docs, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Verify topics
	expectedTopics := []string{"overview", "architecture", "api"}
	if len(docs) != len(expectedTopics) {
		t.Errorf("expected %d topics, got %d", len(expectedTopics), len(docs))
	}

	// Verify overview content
	overviewEntries := docs["overview"]
	if len(overviewEntries) != 1 {
		t.Errorf("expected 1 overview entry, got %d", len(overviewEntries))
	}
	if len(overviewEntries) > 0 {
		content := overviewEntries[0].Content
		if !strings.Contains(content, "multi-line overview") {
			t.Errorf("overview content missing expected text, got: %q", content)
		}
		if !strings.Contains(content, "No need to repeat") {
			t.Errorf("overview content missing expected text, got: %q", content)
		}
	}

	// Verify architecture content
	archEntries := docs["architecture"]
	if len(archEntries) != 1 {
		t.Errorf("expected 1 architecture entry, got %d", len(archEntries))
	}

	// Verify api content
	apiEntries := docs["api"]
	if len(apiEntries) != 1 {
		t.Errorf("expected 1 api entry, got %d", len(apiEntries))
	}
}

func TestExtractConflictingTitlesFails(t *testing.T) {
	fset := token.NewFileSet()
	comments := []*ast.Comment{
		{Slash: token.Pos(1), Text: "//lion:topic title=\"One\""},
		{Slash: token.Pos(10), Text: "//lion:topic title=\"Two\""},
	}
	cg := &ast.CommentGroup{List: comments}
	docs := make(map[string][]DocEntry)
	err := extractFromCommentGroup(fset, cg, "test.go", "Func", docs)
	if err == nil {
		t.Fatalf("expected conflict error, got none")
	}
	if !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("expected conflict in error, got %v", err)
	}
}

func TestExtractMarkerAtTopWithFollowingComments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test Go file with marker-at-top format that gathers following lines
	testFile := filepath.Join(tmpDir, "test.go")
content := `//lion:overview section="Overview section"
//
// This is a multi-line comment.
// It describes the functionality.
// The lion marker is at the top and collects following lines.
package test


//lion:api
// The Config struct holds settings.
// Fields can be configured via files or environment.
// This demonstrates marker-at-top format.
type Config struct {
	Port int
}

//lion:api section="Init"
//
// Initialize creates a new instance.
// It sets up default values.
// Additional info can go on the following lines.
func Initialize() *Config {
	return &Config{Port: 8080}
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Extract documentation
	docs, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Verify topics
	if len(docs) != 2 {
		t.Errorf("expected 2 topics, got %d", len(docs))
	}

	// Verify overview content
	overviewEntries := docs["overview"]
	if len(overviewEntries) != 1 {
		t.Errorf("expected 1 overview entry, got %d", len(overviewEntries))
	}
	if len(overviewEntries) > 0 {
		content := overviewEntries[0].Content
		if !strings.Contains(content, "multi-line comment") {
			t.Errorf("overview content missing expected text, got: %q", content)
		}
		if !strings.Contains(content, "marker is at the top") {
			t.Errorf("overview content missing expected text, got: %q", content)
		}
		// Should NOT contain the lion marker itself
		if strings.Contains(content, "lion:overview") {
			t.Errorf("overview content should not contain lion marker, got: %q", content)
		}
	}

	// Verify api content (should have 2 entries)
	apiEntries := docs["api"]
	if len(apiEntries) != 2 {
		t.Errorf("expected 2 api entries, got %d", len(apiEntries))
	}

	// First api entry (Config struct)
	if len(apiEntries) > 0 {
		content := apiEntries[0].Content
		if !strings.Contains(content, "Config struct holds settings") {
			t.Errorf("api content missing expected text, got: %q", content)
		}
	}

	// Second api entry (Initialize function) - should include marker line content
	if len(apiEntries) > 1 {
		content := apiEntries[1].Content
		if !strings.Contains(content, "Initialize creates a new instance") {
			t.Errorf("api content missing expected text, got: %q", content)
		}
		if !strings.Contains(content, "Additional info can go on the following lines") {
			t.Errorf("api content should include following lines, got: %q", content)
		}
	}
}

func TestMixedCommentFormats(t *testing.T) {
	tmpDir := t.TempDir()

	// Test file that mixes all three formats
	testFile := filepath.Join(tmpDir, "test.go")
	content := `//lion:format1 section="Format 1"
// Format 1: marker first with following content.
package test

//lion:format2 Format 2: Single-line with content on same line
func format2example() {}

/*lion:format3
Format 3: Block comment
with multiple lines
*/
func format3example() {}

//lion:format1 section="Format 1 again"
// Format 1 again: Another marker-first example
// With multiple lines of documentation.
func example() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should have 3 topics
	if len(docs) != 3 {
		t.Errorf("expected 3 topics, got %d", len(docs))
	}

	// Verify format1 has 2 entries
	if len(docs["format1"]) != 2 {
		t.Errorf("expected 2 format1 entries, got %d", len(docs["format1"]))
	}

	// Verify format2 has 1 entry
	if len(docs["format2"]) != 1 {
		t.Errorf("expected 1 format2 entry, got %d", len(docs["format2"]))
	}

	// Verify format3 has 1 entry
	if len(docs["format3"]) != 1 {
		t.Errorf("expected 1 format3 entry, got %d", len(docs["format3"]))
	}
}
