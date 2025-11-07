package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Symbol struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Doc  string `json:"doc"`
	Line int    `json:"line"`
}

type FileSymbols struct {
	FilePath string   `json:"file_path"`
	Package  string   `json:"package"`
	Symbols  []Symbol `json:"symbols"`
}

func main() {
	data, err := os.ReadFile("dissect-output.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var files []FileSymbols
	if err := json.Unmarshal(data, &files); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("# Codebase Audit Report")
	fmt.Println()

	// 1. Find duplicate symbol names
	analyzeSymbolNames(files)

	// 2. Find undocumented code
	analyzeDocumentation(files)

	// 3. Find test utilities that might be consolidated
	analyzeTestUtils(files)

	// 4. Find files with very few symbols (potential for consolidation)
	analyzeSmallFiles(files)

	// 5. Find similar/duplicate function patterns
	analyzeSimilarNames(files)

	// 6. Find potential library candidates
	analyzeLibraryCandidates(files)
}

func analyzeSymbolNames(files []FileSymbols) {
	fmt.Println("## 1. Duplicate Symbol Names (Potential Code Duplication)")
	fmt.Println()

	// Track symbol names across packages
	symbolLocations := make(map[string][]string)

	for _, file := range files {
		for _, symbol := range file.Symbols {
			// Skip common names and unexported symbols
			if symbol.Kind == "const" || symbol.Kind == "var" {
				continue
			}
			if len(symbol.Name) > 0 && symbol.Name[0] >= 'A' && symbol.Name[0] <= 'Z' {
				key := fmt.Sprintf("%s (%s)", symbol.Name, symbol.Kind)
				location := fmt.Sprintf("%s:%d", file.FilePath, symbol.Line)
				symbolLocations[key] = append(symbolLocations[key], location)
			}
		}
	}

	// Report duplicates
	var duplicates []string
	for name, locations := range symbolLocations {
		if len(locations) > 1 {
			duplicates = append(duplicates, name)
		}
	}
	sort.Strings(duplicates)

	for _, name := range duplicates {
		fmt.Printf("- **%s** appears in %d locations:\n", name, len(symbolLocations[name]))
		for _, loc := range symbolLocations[name] {
			fmt.Printf("  - %s\n", loc)
		}
		fmt.Println()
	}

	if len(duplicates) == 0 {
		fmt.Println("No duplicate exported symbols found.\n")
	}
}

func analyzeDocumentation(files []FileSymbols) {
	fmt.Println("## 2. Undocumented Code")
	fmt.Println()

	undocCount := 0
	var undocFiles []string

	for _, file := range files {
		fileUndoc := 0
		for _, symbol := range file.Symbols {
			// Skip private symbols and common patterns
			if len(symbol.Name) == 0 || symbol.Name[0] < 'A' || symbol.Name[0] > 'Z' {
				continue
			}
			if symbol.Name == "String" || symbol.Name == "Error" {
				continue
			}
			if strings.HasSuffix(file.FilePath, "_test.go") && strings.HasPrefix(symbol.Name, "Test") {
				continue // Test functions don't need docs
			}

			if strings.TrimSpace(symbol.Doc) == "" {
				fileUndoc++
				undocCount++
			}
		}

		if fileUndoc > 3 {
			undocFiles = append(undocFiles, fmt.Sprintf("%s (%d undocumented)", file.FilePath, fileUndoc))
		}
	}

	fmt.Printf("Total undocumented exported symbols: %d\n\n", undocCount)
	fmt.Println("Files with >3 undocumented symbols:")
	for _, f := range undocFiles {
		fmt.Printf("- %s\n", f)
	}
	fmt.Println()
}

func analyzeTestUtils(files []FileSymbols) {
	fmt.Println("## 3. Test Utilities (Potential for Consolidation)")
	fmt.Println()

	testUtilFiles := make(map[string][]string)

	for _, file := range files {
		if strings.Contains(file.FilePath, "test") && (strings.Contains(file.FilePath, "util") || strings.Contains(file.FilePath, "helper")) {
			dir := filepath.Dir(file.FilePath)
			for _, symbol := range file.Symbols {
				if symbol.Kind == "func" && len(symbol.Name) > 0 && symbol.Name[0] >= 'A' {
					testUtilFiles[dir] = append(testUtilFiles[dir], fmt.Sprintf("%s (%s)", symbol.Name, file.FilePath))
				}
			}
		}
	}

	if len(testUtilFiles) > 0 {
		fmt.Println("Test utility functions found in multiple locations:")
		for dir, funcs := range testUtilFiles {
			fmt.Printf("\n**%s:**\n", dir)
			for _, f := range funcs {
				fmt.Printf("- %s\n", f)
			}
		}
	} else {
		fmt.Println("No test utility files found.")
	}
	fmt.Println()
}

func analyzeSmallFiles(files []FileSymbols) {
	fmt.Println("## 4. Small Files (Potential for Consolidation)")
	fmt.Println()

	var smallFiles []string

	for _, file := range files {
		if strings.HasSuffix(file.FilePath, "_test.go") {
			continue
		}

		exportedCount := 0
		for _, symbol := range file.Symbols {
			if len(symbol.Name) > 0 && symbol.Name[0] >= 'A' && symbol.Name[0] <= 'Z' {
				exportedCount++
			}
		}

		if exportedCount == 1 {
			symbolNames := []string{}
			for _, symbol := range file.Symbols {
				if len(symbol.Name) > 0 && symbol.Name[0] >= 'A' && symbol.Name[0] <= 'Z' {
					symbolNames = append(symbolNames, fmt.Sprintf("%s (%s)", symbol.Name, symbol.Kind))
				}
			}
			smallFiles = append(smallFiles, fmt.Sprintf("%s: %s", file.FilePath, strings.Join(symbolNames, ", ")))
		}
	}

	fmt.Printf("Files with exactly 1 exported symbol (%d total):\n\n", len(smallFiles))
	for _, f := range smallFiles {
		fmt.Printf("- %s\n", f)
	}
	fmt.Println()
}

func analyzeSimilarNames(files []FileSymbols) {
	fmt.Println("## 5. Similar Function Names (Potential Duplication)")
	fmt.Println()

	// Group by similar prefixes
	nameGroups := make(map[string][]string)

	for _, file := range files {
		for _, symbol := range file.Symbols {
			if symbol.Kind != "func" && symbol.Kind != "method" {
				continue
			}

			// Extract prefix (e.g., "Run" from "RunCommand", "RunGoBuild", etc.)
			if len(symbol.Name) > 3 {
				for _, prefix := range []string{"Run", "Find", "Get", "Create", "Build", "Process", "Extract", "Move", "Test"} {
					if strings.HasPrefix(symbol.Name, prefix) && len(symbol.Name) > len(prefix) {
						location := fmt.Sprintf("%s (%s:%d)", symbol.Name, file.FilePath, symbol.Line)
						nameGroups[prefix] = append(nameGroups[prefix], location)
						break
					}
				}
			}
		}
	}

	for prefix, funcs := range nameGroups {
		if len(funcs) > 5 {
			fmt.Printf("**%s*** prefix (%d functions):\n", prefix, len(funcs))
			sort.Strings(funcs)
			for i, f := range funcs {
				if i < 10 { // Only show first 10
					fmt.Printf("- %s\n", f)
				}
			}
			if len(funcs) > 10 {
				fmt.Printf("- ... and %d more\n", len(funcs)-10)
			}
			fmt.Println()
		}
	}
}

func analyzeLibraryCandidates(files []FileSymbols) {
	fmt.Println("## 6. Potential Library Candidates")
	fmt.Println()

	// Look for common utility patterns
	utilPatterns := map[string][]string{
		"Command Execution": {},
		"File Operations":   {},
		"String Utilities":  {},
		"Go Tooling":        {},
	}

	for _, file := range files {
		for _, symbol := range file.Symbols {
			if len(symbol.Name) == 0 || symbol.Name[0] < 'A' || symbol.Name[0] > 'Z' {
				continue
			}

			location := fmt.Sprintf("%s (%s:%d)", symbol.Name, file.FilePath, symbol.Line)

			if strings.Contains(symbol.Name, "Command") || strings.Contains(symbol.Name, "Run") {
				utilPatterns["Command Execution"] = append(utilPatterns["Command Execution"], location)
			}
			if strings.Contains(symbol.Name, "File") || strings.Contains(symbol.Name, "Path") {
				utilPatterns["File Operations"] = append(utilPatterns["File Operations"], location)
			}
			if strings.Contains(symbol.Name, "String") || strings.Contains(symbol.Name, "Format") {
				utilPatterns["String Utilities"] = append(utilPatterns["String Utilities"], location)
			}
			if strings.Contains(symbol.Name, "Go") && (strings.Contains(symbol.Name, "Build") || strings.Contains(symbol.Name, "Mod") || strings.Contains(symbol.Name, "List")) {
				utilPatterns["Go Tooling"] = append(utilPatterns["Go Tooling"], location)
			}
		}
	}

	for category, items := range utilPatterns {
		if len(items) > 3 {
			fmt.Printf("**%s** (%d functions):\n", category, len(items))
			sort.Strings(items)
			for i, item := range items {
				if i < 15 {
					fmt.Printf("- %s\n", item)
				}
			}
			if len(items) > 15 {
				fmt.Printf("- ... and %d more\n", len(items)-15)
			}
			fmt.Println()
		}
	}
}
