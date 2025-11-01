package symbols

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neongreen/mono/dissect/pkg/typeinfo"
)

func TestFindExportedSymbols_SingleFunction(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"main.go": `package main

func Hello() string {
	return "hello"
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "main.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(symbols))
	}

	sym := symbols[0]
	if sym.Name != "Hello" {
		t.Errorf("Expected name 'Hello', got '%s'", sym.Name)
	}
	if sym.Kind != "func" {
		t.Errorf("Expected kind 'func', got '%s'", sym.Kind)
	}
}

func TestFindExportedSymbols_MultipleFunctions(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"funcs.go": `package funcs

func Alpha() {}
func Beta() {}
func Gamma() {}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "funcs.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	if len(symbols) != 3 {
		t.Fatalf("Expected 3 symbols, got %d", len(symbols))
	}

	names := []string{symbols[0].Name, symbols[1].Name, symbols[2].Name}
	expected := []string{"Alpha", "Beta", "Gamma"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("Expected %s at position %d, got %s", expected[i], i, name)
		}
	}
}

func TestFindExportedSymbols_ExportedType(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"types.go": `package types

type Person struct {
	Name string
	Age  int
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "types.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(symbols))
	}

	sym := symbols[0]
	if sym.Name != "Person" {
		t.Errorf("Expected name 'Person', got '%s'", sym.Name)
	}
	if sym.Kind != "type" {
		t.Errorf("Expected kind 'type', got '%s'", sym.Kind)
	}
}

func TestFindExportedSymbols_ExportedInterface(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"interfaces.go": `package interfaces

type Reader interface {
	Read(p []byte) (n int, err error)
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "interfaces.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(symbols))
	}

	sym := symbols[0]
	if sym.Name != "Reader" {
		t.Errorf("Expected name 'Reader', got '%s'", sym.Name)
	}
	if sym.Kind != "type" {
		t.Errorf("Expected kind 'type', got '%s'", sym.Kind)
	}
}

func TestFindExportedSymbols_ExportedVar(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"vars.go": `package vars

var DefaultTimeout = 30
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "vars.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(symbols))
	}

	sym := symbols[0]
	if sym.Name != "DefaultTimeout" {
		t.Errorf("Expected name 'DefaultTimeout', got '%s'", sym.Name)
	}
	if sym.Kind != "var" {
		t.Errorf("Expected kind 'var', got '%s'", sym.Kind)
	}
}

func TestFindExportedSymbols_ExportedConst(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"consts.go": `package consts

const MaxRetries = 3
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "consts.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(symbols))
	}

	sym := symbols[0]
	if sym.Name != "MaxRetries" {
		t.Errorf("Expected name 'MaxRetries', got '%s'", sym.Name)
	}
	if sym.Kind != "const" {
		t.Errorf("Expected kind 'const', got '%s'", sym.Kind)
	}
}

func TestFindExportedSymbols_Mixed(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"mixed.go": `package mixed

const Version = "1.0.0"

var DefaultName = "unknown"

type Config struct {
	Host string
}

func NewConfig() *Config {
	return &Config{}
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "mixed.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	if len(symbols) != 4 {
		t.Fatalf("Expected 4 symbols, got %d", len(symbols))
	}

	// Check we got all kinds
	kinds := make(map[string]bool)
	names := make(map[string]bool)
	for _, sym := range symbols {
		kinds[sym.Kind] = true
		names[sym.Name] = true
	}

	if !kinds["const"] || !kinds["var"] || !kinds["type"] || !kinds["func"] {
		t.Errorf("Expected all kinds (const, var, type, func), got: %v", kinds)
	}

	expectedNames := []string{"Version", "DefaultName", "Config", "NewConfig"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("Expected to find symbol '%s'", name)
		}
	}
}

func TestFindExportedSymbols_UnexportedIgnored(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"unexported.go": `package unexported

func Exported() {}
func unexported() {}

type ExportedType struct{}
type unexportedType struct{}

var ExportedVar = 1
var unexportedVar = 2

const ExportedConst = "public"
const unexportedConst = "private"
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "unexported.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	// Should only find 4 exported symbols
	if len(symbols) != 4 {
		t.Fatalf("Expected 4 exported symbols, got %d", len(symbols))
	}

	// Verify none of the unexported names are present
	for _, sym := range symbols {
		if sym.Name == "unexported" || sym.Name == "unexportedType" ||
			sym.Name == "unexportedVar" || sym.Name == "unexportedConst" {
			t.Errorf("Found unexported symbol: %s", sym.Name)
		}
	}
}

func TestFindExportedSymbols_Methods(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"methods.go": `package methods

type Person struct {
	Name string
}

func (p *Person) GetName() string {
	return p.Name
}

func (p *Person) SetName(name string) {
	p.Name = name
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "methods.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	// Should find: Person (type), GetName (method), SetName (method)
	if len(symbols) != 3 {
		t.Fatalf("Expected 3 symbols, got %d", len(symbols))
	}

	// Count methods
	methodCount := 0
	for _, sym := range symbols {
		if sym.Kind == "method" {
			methodCount++
		}
	}

	if methodCount != 2 {
		t.Errorf("Expected 2 methods, got %d", methodCount)
	}
}

func TestFindExportedSymbols_EmptyFile(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"empty.go": `package empty
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	symbols, err := FindExportedSymbols(filepath.Join(tmpDir, "empty.go"), pkg)
	if err != nil {
		t.Fatalf("FindExportedSymbols failed: %v", err)
	}

	if len(symbols) != 0 {
		t.Errorf("Expected 0 symbols, got %d", len(symbols))
	}
}

func TestFindExportedSymbols_NonExistentFile(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"exists.go": `package test

func Exists() {}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	_, err = FindExportedSymbols(filepath.Join(tmpDir, "nonexistent.go"), pkg)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// Helper function

func createTempPackage(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "symbols_test_*")
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

