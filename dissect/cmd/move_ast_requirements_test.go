package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestASTRequirements contains tests that demonstrate why AST-based manipulation
// is required and string manipulation would fail.
//
// These tests verify edge cases that would break with string-based approaches:
// - Different formatting styles (tabs vs spaces, line breaks)
// - Build tags and compiler directives
// - Unicode in comments and identifiers
// - Complex comment positioning

func TestMoveWithDifferentFormatting(t *testing.T) {
	// This test demonstrates that string-based extraction would fail when
	// the same code is formatted differently (e.g., different indentation,
	// line breaks, spacing). AST handles this correctly because it preserves
	// semantic structure, not textual representation.

	tmpDir, err := os.MkdirTemp("", "dissect_formatting_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := `module example.com/formattingtest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Source with unusual but valid formatting
	sourceCode := `package main

import "fmt"

// oddlyFormattedFunc has unusual formatting
func oddlyFormattedFunc(  a   int,
b int,
	c    int   )    (   int  ,  error   ){
		if   a   >   0   {
			return   a+b+c  ,  nil
		}
return 0,fmt.Errorf("error")
}

func main() {
	oddlyFormattedFunc(1, 2, 3)
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./dissect/cmd")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	targetFile := filepath.Join(tmpDir, "target.go")
	cmd := exec.Command(dissectBinary, "move", "source.go:oddlyFormattedFunc", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move function: %v\nOutput: %s", err, output)
	}

	// Verify the function was moved and code builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}

	// Verify target file contains the function
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	if !strings.Contains(string(targetContent), "func oddlyFormattedFunc") {
		t.Errorf("Function not found in target file")
	}
}

func TestMoveWithBuildTags(t *testing.T) {
	// This test demonstrates that string manipulation would fail with build tags
	// because it would need to understand Go's build tag syntax. AST parsing
	// handles this correctly.

	tmpDir, err := os.MkdirTemp("", "dissect_buildtags_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := `module example.com/buildtagstest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Source with build tags
	sourceCode := `//go:build linux
// +build linux

package main

import "fmt"

// linuxSpecificFunc is only compiled on Linux
//
// This function has multiple comment blocks
func linuxSpecificFunc() {
	fmt.Println("Linux specific")
}

func main() {
	linuxSpecificFunc()
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./dissect/cmd")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	targetFile := filepath.Join(tmpDir, "target.go")
	cmd := exec.Command(dissectBinary, "move", "source.go:linuxSpecificFunc", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move function: %v\nOutput: %s", err, output)
	}

	// Verify the function was moved with its doc comments
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	targetStr := string(targetContent)
	if !strings.Contains(targetStr, "func linuxSpecificFunc") {
		t.Errorf("Function not found in target file")
	}

	// Verify doc comments were preserved
	if !strings.Contains(targetStr, "linuxSpecificFunc is only compiled on Linux") {
		t.Errorf("Doc comment not preserved in target file")
	}

	// Note: Build tags at file level are NOT moved with the function
	// They apply to the entire file, not individual functions
}

func TestMoveWithUnicodeComments(t *testing.T) {
	// This test demonstrates that string manipulation might fail with Unicode
	// characters, especially when using byte offsets vs rune offsets.
	// AST handles Unicode correctly.

	tmpDir, err := os.MkdirTemp("", "dissect_unicode_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := `module example.com/unicodetest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Source with Unicode in comments
	sourceCode := `package main

import "fmt"

// unicodeFunc handles Unicode: 你好世界 🚀 Привет мир
// Multiple languages: مرحبا العالم, שלום עולם
// Emoji and symbols: ✓ ✗ → ⇒ ∞
func unicodeFunc() {
	fmt.Println("Unicode: 你好")
}

func main() {
	unicodeFunc()
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./dissect/cmd")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	targetFile := filepath.Join(tmpDir, "target.go")
	cmd := exec.Command(dissectBinary, "move", "source.go:unicodeFunc", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move function: %v\nOutput: %s", err, output)
	}

	// Verify the function was moved with Unicode comments intact
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	targetStr := string(targetContent)
	if !strings.Contains(targetStr, "func unicodeFunc") {
		t.Errorf("Function not found in target file")
	}

	// Verify all Unicode characters were preserved
	unicodeTests := []string{
		"你好世界",
		"🚀",
		"Привет мир",
		"مرحبا العالم",
		"שלום עולם",
		"✓ ✗ → ⇒ ∞",
	}

	for _, unicode := range unicodeTests {
		if !strings.Contains(targetStr, unicode) {
			t.Errorf("Unicode string %q not preserved in target file", unicode)
		}
	}

	// Verify code builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

func TestMoveWithComplexCommentPatterns(t *testing.T) {
	// This test demonstrates that string manipulation would struggle with
	// complex comment patterns (multiple comment blocks, inline comments, etc.)
	// AST correctly associates comments with declarations.

	tmpDir, err := os.MkdirTemp("", "dissect_complex_comments_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := `module example.com/complexcommentstest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Source with complex comment patterns
	sourceCode := `package main

import "fmt"

// complexFunc demonstrates various comment styles
//
// This is a longer description that spans multiple lines.
// It includes blank comment lines and detailed documentation.
//
// Parameters:
//   - x: the first parameter
//   - y: the second parameter
//
// Returns:
//   - The sum of x and y
func complexFunc(x, y int) int {
	return x + y
}

// anotherFunc is separate
func anotherFunc() {
	fmt.Println("another")
}

func main() {
	complexFunc(1, 2)
	anotherFunc()
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./dissect/cmd")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	targetFile := filepath.Join(tmpDir, "target.go")
	cmd := exec.Command(dissectBinary, "move", "source.go:complexFunc", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move function: %v\nOutput: %s", err, output)
	}

	// Verify the function was moved with all comment blocks
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	targetStr := string(targetContent)
	if !strings.Contains(targetStr, "func complexFunc") {
		t.Errorf("Function not found in target file")
	}

	// Verify all parts of the doc comment are preserved
	commentParts := []string{
		"complexFunc demonstrates various comment styles",
		"This is a longer description",
		"Parameters:",
		"- x: the first parameter",
		"- y: the second parameter",
		"Returns:",
		"- The sum of x and y",
	}

	for _, part := range commentParts {
		if !strings.Contains(targetStr, part) {
			t.Errorf("Comment part %q not preserved in target file", part)
		}
	}

	// Verify that anotherFunc's comment is NOT in the target file
	if strings.Contains(targetStr, "anotherFunc is separate") {
		t.Errorf("Other function's comments should not be in target file")
	}

	// Verify code builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}
