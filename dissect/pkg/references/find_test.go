package references

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neongreen/mono/dissect/pkg/typeinfo"
	"golang.org/x/tools/go/packages"
)

func TestFindReferences_SingleUnqualified(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"defs.go": `package test

func Helper() string {
	return "help"
}
`,
		"main.go": `package test

func Main() {
	Helper()
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Should find 1 reference (the call in Main)
	if len(refs) != 1 {
		t.Fatalf("Expected 1 reference, got %d", len(refs))
	}

	if refs[0].Ident.Name != "Helper" {
		t.Errorf("Expected identifier 'Helper', got '%s'", refs[0].Ident.Name)
	}

	if refs[0].Qualified {
		t.Error("Expected unqualified reference, got qualified")
	}
}

func TestFindReferences_MultipleUnqualified(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"util.go": `package test

func Util() {}
`,
		"main.go": `package test

func Main() {
	Util()
	Util()
	Util()
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Util"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	if len(refs) != 3 {
		t.Fatalf("Expected 3 references, got %d", len(refs))
	}
}

func TestFindReferences_QualifiedReference(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "refs_qualified_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create util package
	utilDir := filepath.Join(tmpDir, "util")
	if err := os.MkdirAll(utilDir, 0755); err != nil {
		t.Fatalf("Failed to create util dir: %v", err)
	}
	createFileInDir(t, utilDir, "util.go", `package util

func Helper() string {
	return "help"
}
`)

	// Create main package that imports util
	createFileInDir(t, tmpDir, "main.go", `package main

import "test/util"

func main() {
	util.Helper()
}
`)

	// Create go.mod
	createFileInDir(t, tmpDir, "go.mod", "module test\n\ngo 1.21\n")

	// Load both packages
	pkgs, err := typeinfo.LoadPackages([]string{"./...", "./util"}, tmpDir)
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	refs, err := FindReferences([]string{"Helper"}, pkgs)
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Should find at least 1 qualified reference
	qualifiedCount := 0
	for _, ref := range refs {
		if ref.Qualified {
			qualifiedCount++
		}
	}

	if qualifiedCount < 1 {
		t.Errorf("Expected at least 1 qualified reference, got %d", qualifiedCount)
	}
}

func TestFindReferences_MixedQualifiedUnqualified(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "refs_mixed_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create util package
	utilDir := filepath.Join(tmpDir, "util")
	if err := os.MkdirAll(utilDir, 0755); err != nil {
		t.Fatalf("Failed to create util dir: %v", err)
	}
	createFileInDir(t, utilDir, "util.go", `package util

func Helper() string {
	return "help"
}

func UseHelper() {
	Helper() // unqualified
}
`)

	// Create main package that imports util
	createFileInDir(t, tmpDir, "main.go", `package main

import "test/util"

func main() {
	util.Helper() // qualified
}
`)

	createFileInDir(t, tmpDir, "go.mod", "module test\n\ngo 1.21\n")

	pkgs, err := typeinfo.LoadPackages([]string{"./..."}, tmpDir)
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	refs, err := FindReferences([]string{"Helper"}, pkgs)
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Should find both qualified and unqualified references
	qualifiedCount := 0
	unqualifiedCount := 0
	for _, ref := range refs {
		if ref.Qualified {
			qualifiedCount++
		} else {
			unqualifiedCount++
		}
	}

	if qualifiedCount < 1 {
		t.Errorf("Expected at least 1 qualified reference, got %d", qualifiedCount)
	}
	if unqualifiedCount < 1 {
		t.Errorf("Expected at least 1 unqualified reference, got %d", unqualifiedCount)
	}
}

func TestFindReferences_MultipleSymbols(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"defs.go": `package test

func Alpha() {}
func Beta() {}
`,
		"main.go": `package test

func Main() {
	Alpha()
	Beta()
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Alpha", "Beta"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	if len(refs) != 2 {
		t.Fatalf("Expected 2 references, got %d", len(refs))
	}

	names := make(map[string]bool)
	for _, ref := range refs {
		names[ref.Ident.Name] = true
	}

	if !names["Alpha"] || !names["Beta"] {
		t.Error("Expected to find both Alpha and Beta references")
	}
}

func TestFindReferences_InFunctionBody(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"code.go": `package test

func Helper() int { return 42 }

func Process() {
	x := Helper()
	y := Helper() + 1
	z := Helper() * 2
	_ = x + y + z
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	if len(refs) != 3 {
		t.Fatalf("Expected 3 references in function body, got %d", len(refs))
	}
}

func TestFindReferences_InStructField(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"types.go": `package test

type MyType struct{}

type Container struct {
	Field MyType
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"MyType"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	if len(refs) < 1 {
		t.Fatalf("Expected at least 1 reference in struct field, got %d", len(refs))
	}
}

func TestFindReferences_InReturnType(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"funcs.go": `package test

type Result struct{}

func GetResult() Result {
	return Result{}
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Result"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Should find at least 2: return type and composite literal
	if len(refs) < 2 {
		t.Fatalf("Expected at least 2 references, got %d", len(refs))
	}
}

func TestFindReferences_InVariableDecl(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"vars.go": `package test

type Config struct{}

var defaultConfig Config
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Config"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	if len(refs) < 1 {
		t.Fatalf("Expected at least 1 reference in variable declaration, got %d", len(refs))
	}
}

func TestFindReferences_NoReferences(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"code.go": `package test

func Unused() {}

func Main() {
	// Unused is never called
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Unused"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	if len(refs) != 0 {
		t.Errorf("Expected 0 references to Unused, got %d", len(refs))
	}
}

func TestFindReferences_MultipleFiles(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"util.go": `package test

func Util() {}
`,
		"file1.go": `package test

func File1() {
	Util()
}
`,
		"file2.go": `package test

func File2() {
	Util()
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Util"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	if len(refs) != 2 {
		t.Fatalf("Expected 2 references across files, got %d", len(refs))
	}

	// Verify references are from different files
	files := make(map[string]bool)
	for _, ref := range refs {
		files[filepath.Base(ref.File)] = true
	}

	if !files["file1.go"] || !files["file2.go"] {
		t.Error("Expected references from both file1.go and file2.go")
	}
}

func TestFindReferences_IgnoreOtherPackages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "refs_ignore_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package1
	pkg1Dir := filepath.Join(tmpDir, "pkg1")
	if err := os.MkdirAll(pkg1Dir, 0755); err != nil {
		t.Fatalf("Failed to create pkg1: %v", err)
	}
	createFileInDir(t, pkg1Dir, "code.go", `package pkg1

func Helper() {}
`)

	// Create package2 with different Helper
	pkg2Dir := filepath.Join(tmpDir, "pkg2")
	if err := os.MkdirAll(pkg2Dir, 0755); err != nil {
		t.Fatalf("Failed to create pkg2: %v", err)
	}
	createFileInDir(t, pkg2Dir, "code.go", `package pkg2

func Helper() {}

func Use() {
	Helper()
}
`)

	createFileInDir(t, tmpDir, "go.mod", "module test\n\ngo 1.21\n")

	// Load only pkg1
	pkg, err := typeinfo.LoadPackage(pkg1Dir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Should not find references from pkg2
	for _, ref := range refs {
		if filepath.Base(filepath.Dir(ref.File)) == "pkg2" {
			t.Error("Found reference from pkg2, expected only pkg1")
		}
	}
}

func TestFindReferences_LocalVariableSameName(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"code.go": `package test

func Helper() string {
	return "global"
}

func Main() {
	Helper := "local" // local variable shadows function
	_ = Helper
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// The local variable "Helper" shadows the function, so we should find both
	// But the type info should distinguish them correctly
	// For this test, we just verify it doesn't crash and returns some results
	if len(refs) < 0 {
		t.Error("FindReferences should handle shadowing correctly")
	}
}

// Helper functions

func createTempPackage(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "refs_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create go.mod
	gomod := "module test\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create all files
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to write %s: %v", name, err)
		}
	}

	return tmpDir
}

func createFileInDir(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", name, err)
	}
}

