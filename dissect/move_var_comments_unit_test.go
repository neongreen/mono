package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMoveVarCommentsVariousScenarios tests multiple scenarios for comment preservation
// when moving var declarations, ensuring all types of comments are handled correctly.
func TestMoveVarCommentsVariousScenarios(t *testing.T) {
	tests := []struct {
		name           string
		sourceCode     string
		varName        string
		expectInTarget []string // Comments that should be in target
		expectInSource []string // Comments that should remain in source
	}{
		{
			name: "var with cobra.Command style comments",
			sourceCode: `package main

import "fmt"

// rootCmd simulates a cobra.Command structure
var rootCmd = struct {
	Use   string
	Short string
	Long  string
	Run   func()
}{
	Use:   "app",
	Short: "Application short description",
	// Long description comment
	Long: "Application long description",
	// Run function comment
	Run: func() {
		// Implementation comment
		fmt.Println("running")
	},
}

// otherVar is unrelated
var otherVar = 42

func main() {
	fmt.Println(rootCmd.Use)
}
`,
			varName: "rootCmd",
			expectInTarget: []string{
				"rootCmd simulates a cobra.Command structure",
				"Long description comment",
				"Run function comment",
				"Implementation comment",
			},
			expectInSource: []string{
				"otherVar is unrelated",
			},
		},
		{
			name: "var with struct literal and multiple internal comments",
			sourceCode: `package main

// Config holds configuration
var Config = struct {
	Host string
	Port int
	SSL  bool
}{
	// Default host
	Host: "localhost",
	// Default port
	Port: 8080,
	// SSL is disabled by default
	SSL: false,
}

// Another variable
var Another = "test"

func main() {
	println(Config.Host)
}
`,
			varName: "Config",
			expectInTarget: []string{
				"Config holds configuration",
				"Default host",
				"Default port",
				"SSL is disabled by default",
			},
			expectInSource: []string{
				"Another variable",
			},
		},
		{
			name: "var with map literal and comments",
			sourceCode: `package main

// Routes defines application routes
var Routes = map[string]string{
	// Home page
	"/":      "home",
	// About page
	"/about": "about",
	// Contact page with special handling
	"/contact": "contact",
}

// Other routes
var OtherRoutes = map[string]string{}

func main() {
	println(Routes["/"])
}
`,
			varName: "Routes",
			expectInTarget: []string{
				"Routes defines application routes",
				"Home page",
				"About page",
				"Contact page with special handling",
			},
			expectInSource: []string{
				"Other routes",
			},
		},
		{
			name: "var with slice literal and comments",
			sourceCode: `package main

// Handlers lists all handlers
var Handlers = []func(){
	// First handler
	func() { println("first") },
	// Second handler with logging
	func() { println("second") },
}

// Different handlers
var DifferentHandlers = []func(){}

func main() {
	Handlers[0]()
}
`,
			varName: "Handlers",
			expectInTarget: []string{
				"Handlers lists all handlers",
				"First handler",
				"Second handler with logging",
			},
			expectInSource: []string{
				"Different handlers",
			},
		},
		{
			name: "var with nested structs and comments",
			sourceCode: `package main

// Server configuration
var Server = struct {
	HTTP struct {
		Port int
		Host string
	}
	DB struct {
		Name string
		Host string
	}
}{
	HTTP: struct {
		Port int
		Host string
	}{
		// HTTP port
		Port: 8080,
		// HTTP host
		Host: "localhost",
	},
	DB: struct {
		Name string
		Host string
	}{
		// Database name
		Name: "mydb",
		// Database host
		Host: "localhost",
	},
}

// Cache configuration
var Cache = "redis"

func main() {
	println(Server.HTTP.Host)
}
`,
			varName: "Server",
			expectInTarget: []string{
				"Server configuration",
				"HTTP port",
				"HTTP host",
				"Database name",
				"Database host",
			},
			expectInSource: []string{
				"Cache configuration",
			},
		},
		{
			name: "var with only Doc comment (no internal comments)",
			sourceCode: `package main

// SimpleVar is a simple variable
var SimpleVar = "value"

// AnotherVar is another variable
var AnotherVar = "value2"

func main() {
	println(SimpleVar)
}
`,
			varName: "SimpleVar",
			expectInTarget: []string{
				"SimpleVar is a simple variable",
			},
			expectInSource: []string{
				"AnotherVar is another variable",
			},
		},
		{
			name: "var with comments on function literal",
			sourceCode: `package main

// Handler is a request handler
var Handler = func() {
	// Step 1: validate input
	validate()
	// Step 2: process data
	process()
	// Step 3: send response
	respond()
}

func validate() {}
func process() {}
func respond() {}

// OtherHandler is different
var OtherHandler = func() {}

func main() {
	Handler()
}
`,
			varName: "Handler",
			expectInTarget: []string{
				"Handler is a request handler",
				"Step 1: validate input",
				"Step 2: process data",
				"Step 3: send response",
			},
			expectInSource: []string{
				"OtherHandler is different",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the test
			tmpDir, err := os.MkdirTemp("", "dissect_var_comments_scenarios_")
			if err != nil {
				t.Fatalf("Failed to create temporary directory: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create a test Go module
			goMod := `module example.com/test

go 1.24
`
			if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644); err != nil {
				t.Fatalf("Failed to create go.mod: %v", err)
			}

			// Create source file
			sourceFile := filepath.Join(tmpDir, "source.go")
			if err := os.WriteFile(sourceFile, []byte(tt.sourceCode), 0o644); err != nil {
				t.Fatalf("Failed to create source file: %v", err)
			}

			// Build the dissect binary
			dissectBinary := filepath.Join(tmpDir, "dissect")
			buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./dissect")
			buildCmd.Dir = findRepoRoot(t)
			if output, err := buildCmd.CombinedOutput(); err != nil {
				t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
			}

			// Move the var to a new file
			targetFile := filepath.Join(tmpDir, "target.go")
			cmd := exec.Command(dissectBinary, "move", "source.go:"+tt.varName, "target.go")
			cmd.Dir = tmpDir
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("Failed to move var: %v\nOutput: %s", err, output)
			}

			// Read target file
			targetContent, err := os.ReadFile(targetFile)
			if err != nil {
				t.Fatalf("Failed to read target file: %v", err)
			}
			targetStr := string(targetContent)

			// Check that var is in target
			if !strings.Contains(targetStr, "var "+tt.varName) {
				t.Errorf("%s should be in target file", tt.varName)
			}

			// Check all expected comments are in target
			for _, expectedComment := range tt.expectInTarget {
				if !strings.Contains(targetStr, expectedComment) {
					t.Errorf("Expected comment in target not found: %q\nTarget file:\n%s", expectedComment, targetStr)
				}
			}

			// Read source file
			sourceContent, err := os.ReadFile(sourceFile)
			if err != nil {
				t.Fatalf("Failed to read source file: %v", err)
			}
			sourceStr := string(sourceContent)

			// Check that var is not in source
			if strings.Contains(sourceStr, "var "+tt.varName) {
				t.Errorf("%s should have been removed from source", tt.varName)
			}

			// Check that moved comments are not orphaned in source
			for _, movedComment := range tt.expectInTarget {
				if strings.Contains(sourceStr, movedComment) {
					t.Errorf("Moved comment should not be orphaned in source: %q\nSource file:\n%s", movedComment, sourceStr)
				}
			}

			// Check that other comments remain in source
			for _, remainingComment := range tt.expectInSource {
				if !strings.Contains(sourceStr, remainingComment) {
					t.Errorf("Expected comment in source not found: %q\nSource file:\n%s", remainingComment, sourceStr)
				}
			}

			// Verify the code still builds
			buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
			buildCmd.Dir = tmpDir
			if output, err := buildCmd.CombinedOutput(); err != nil {
				t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
			}
		})
	}
}

// TestMoveConstWithComments tests that const declarations also preserve internal comments
func TestMoveConstWithComments(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_const_comments_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test Go module
	goMod := `module example.com/consttest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a source file with a const that has internal comments
	sourceCode := `package main

// Status constants for application states
const (
	// StatusPending indicates a pending state
	StatusPending = "pending"
	// StatusActive indicates an active state
	StatusActive = "active"
	// StatusComplete indicates a completed state
	StatusComplete = "complete"
)

// Other constants
const Version = "1.0.0"

func main() {
	println(Version)
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0o644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Build the dissect binary
	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./dissect")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	// Move StatusPending to a new file (note: grouped const moves as a block)
	targetFile := filepath.Join(tmpDir, "target.go")
	cmd := exec.Command(dissectBinary, "move", "source.go:StatusPending", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move const: %v\nOutput: %s", err, output)
	}

	// Read target file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	targetStr := string(targetContent)

	// Check that all status constants and their comments are in target (grouped declaration moves as a block)
	expectedInTarget := []string{
		"Status constants for application states",
		"StatusPending indicates a pending state",
		"StatusActive indicates an active state",
		"StatusComplete indicates a completed state",
		"StatusPending",
		"StatusActive",
		"StatusComplete",
	}
	for _, expected := range expectedInTarget {
		if !strings.Contains(targetStr, expected) {
			t.Errorf("Expected in target not found: %q\nTarget file:\n%s", expected, targetStr)
		}
	}

	// Read source file
	sourceContent, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}
	sourceStr := string(sourceContent)

	// Check that moved const is not in source
	if strings.Contains(sourceStr, "StatusPending") {
		t.Errorf("StatusPending should have been removed from source")
	}

	// Check that Version constant remains in source
	if !strings.Contains(sourceStr, "const Version") {
		t.Errorf("Version constant should still be in source")
	}

	// Verify the code still builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestMoveTypeWithComments tests that type declarations preserve internal comments
func TestMoveTypeWithComments(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_type_comments_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test Go module
	goMod := `module example.com/typetest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a source file with a type that has internal comments
	sourceCode := `package main

// User represents a system user
type User struct {
	// ID is the unique identifier
	ID int
	// Name is the user's full name
	Name string
	// Email is the contact email
	Email string
}

// Product represents a product
type Product struct {
	Name  string
	Price float64
}

func main() {
	u := User{ID: 1, Name: "test", Email: "test@test.com"}
	println(u.Name)
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0o644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Build the dissect binary
	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./dissect")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	// Move User type to a new file
	targetFile := filepath.Join(tmpDir, "target.go")
	cmd := exec.Command(dissectBinary, "move", "source.go:User", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move type: %v\nOutput: %s", err, output)
	}

	// Read target file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	targetStr := string(targetContent)

	// Check that type and all field comments are in target
	expectedInTarget := []string{
		"User represents a system user",
		"ID is the unique identifier",
		"Name is the user's full name",
		"Email is the contact email",
		"type User struct",
	}
	for _, expected := range expectedInTarget {
		if !strings.Contains(targetStr, expected) {
			t.Errorf("Expected in target not found: %q\nTarget file:\n%s", expected, targetStr)
		}
	}

	// Read source file
	sourceContent, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}
	sourceStr := string(sourceContent)

	// Check that User type is not in source
	if strings.Contains(sourceStr, "type User") {
		t.Errorf("User type should have been removed from source")
	}

	// Check that User comments are not orphaned in source
	for _, movedComment := range expectedInTarget {
		if strings.Contains(sourceStr, movedComment) {
			t.Errorf("Moved comment should not be orphaned in source: %q\nSource file:\n%s", movedComment, sourceStr)
		}
	}

	// Check that Product type remains in source
	if !strings.Contains(sourceStr, "type Product") {
		t.Errorf("Product type should still be in source")
	}

	// Verify the code still builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}
