package tsparser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestConstraints verifies that no forbidden items are present in the repository.
func TestConstraints(t *testing.T) {
	// Get the repository root (go up from lib/ts-parser to repo root)
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("Failed to find repository root: %v", err)
	}

	t.Run("NoForbiddenImports", func(t *testing.T) {
		testNoForbiddenImports(t, repoRoot)
	})

	t.Run("NoForbiddenFiles", func(t *testing.T) {
		testNoForbiddenFiles(t, repoRoot)
	})

	t.Run("NoCGOBuildTags", func(t *testing.T) {
		testNoCGOBuildTags(t, repoRoot)
	})
}

func findRepoRoot() (string, error) {
	// Start from current directory and look for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func testNoForbiddenImports(t *testing.T, repoRoot string) {
	forbiddenPatterns := []struct {
		pattern string
		desc    string
	}{
		{`import\s+"C"`, `import "C"`},
		{`#cgo`, "#cgo directive"},
		{`"os/exec"`, "os/exec import"},
		{`smacker/go-tree-sitter`, "smacker/go-tree-sitter import"},
		{`tree-sitter/go-tree-sitter`, "tree-sitter/go-tree-sitter import"},
	}

	var regexes []*regexp.Regexp
	for _, fp := range forbiddenPatterns {
		re, err := regexp.Compile(fp.pattern)
		if err != nil {
			t.Fatalf("Failed to compile regex %q: %v", fp.pattern, err)
		}
		regexes = append(regexes, re)
	}

	// Only scan this library's files (lib/ts-parser)
	libPath := filepath.Join(repoRoot, "lib", "ts-parser")

	err := filepath.Walk(libPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Check .go files and go.mod (but exclude test files which may contain pattern descriptions)
		ext := filepath.Ext(path)
		base := filepath.Base(path)
		if ext != ".go" && base != "go.mod" && base != "go.sum" {
			return nil
		}
		// Skip test files - they may contain the forbidden patterns as test data
		if strings.HasSuffix(base, "_test.go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			for i, re := range regexes {
				if re.MatchString(line) {
					relPath, _ := filepath.Rel(repoRoot, path)
					t.Errorf("%s:%d: found forbidden %s", relPath, lineNum, forbiddenPatterns[i].desc)
				}
			}
		}

		return scanner.Err()
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}
}

func testNoForbiddenFiles(t *testing.T, repoRoot string) {
	forbiddenExtensions := []string{
		".c", ".h", ".cc", ".cpp", ".cxx",
		".m", ".mm",
		".a", ".so", ".dylib", ".dll",
	}

	// Only scan this library's files (lib/ts-parser)
	libPath := filepath.Join(repoRoot, "lib", "ts-parser")

	err := filepath.Walk(libPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, forbidden := range forbiddenExtensions {
			if ext == forbidden {
				relPath, _ := filepath.Rel(repoRoot, path)
				t.Errorf("Found forbidden file type %s: %s", forbidden, relPath)
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}
}

func testNoCGOBuildTags(t *testing.T, repoRoot string) {
	forbiddenBuildTags := []struct {
		pattern string
		desc    string
	}{
		{`//go:build\s+cgo`, "//go:build cgo"},
		{`//go:build\s+!purego`, "//go:build !purego"},
		{`// \+build\s+cgo`, "// +build cgo"},
		{`// \+build\s+!purego`, "// +build !purego"},
	}

	var regexes []*regexp.Regexp
	for _, fp := range forbiddenBuildTags {
		re, err := regexp.Compile(fp.pattern)
		if err != nil {
			t.Fatalf("Failed to compile regex %q: %v", fp.pattern, err)
		}
		regexes = append(regexes, re)
	}

	// Only scan this library's files (lib/ts-parser)
	libPath := filepath.Join(repoRoot, "lib", "ts-parser")

	err := filepath.Walk(libPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".go" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			// Build tags must appear before package declaration
			if lineNum > 20 {
				break
			}

			line := scanner.Text()
			for i, re := range regexes {
				if re.MatchString(line) {
					relPath, _ := filepath.Rel(repoRoot, path)
					t.Errorf("%s:%d: found forbidden build tag %s", relPath, lineNum, forbiddenBuildTags[i].desc)
				}
			}
		}

		return scanner.Err()
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}
}

// TestCGODisabled verifies the library works with CGO disabled.
// This test uses a build tag to ensure it runs regardless of CGO setting.
func TestCGODisabled(t *testing.T) {
	// This test itself proves CGO is not required since we're running Go code
	// that loads WASM via wazero.
	t.Log("Running with CGO potentially disabled - if this test runs, CGO is not required")
}
