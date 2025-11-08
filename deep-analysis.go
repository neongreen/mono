package main

import (
	"encoding/json"
	"fmt"
	"os"
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

	fmt.Println("# Deep Analysis - Additional Issues")
	fmt.Println()

	// 1. Files with very similar purposes (by directory structure)
	analyzeDirectoryStructure(files)

	// 2. Test helper functions
	analyzeTestHelpers(files)

	// 3. Mock/stub implementations
	analyzeMocksStubs(files)

	// 4. Init functions (potential legacy setup)
	analyzeInitFunctions(files)

	// 5. Type definitions that could be shared
	analyzeSharedTypes(files)

	// 6. Deprecated/legacy patterns
	analyzeLegacyPatterns(files)

	// 7. Over-fragmented packages
	analyzeFragmentation(files)
}

func analyzeDirectoryStructure(files []FileSymbols) {
	fmt.Println("## Directory Structure Analysis")
	fmt.Println()

	packageFiles := make(map[string][]string)
	packageSymbolCount := make(map[string]int)

	for _, file := range files {
		pkg := file.Package
		packageFiles[pkg] = append(packageFiles[pkg], file.FilePath)
		packageSymbolCount[pkg] += len(file.Symbols)
	}

	fmt.Println("### Packages with few symbols (potential for consolidation):")
	for pkg, count := range packageSymbolCount {
		if count < 5 && len(packageFiles[pkg]) > 1 && pkg != "main" {
			fmt.Printf("- **%s**: %d symbols across %d files\n", pkg, count, len(packageFiles[pkg]))
			for _, f := range packageFiles[pkg] {
				fmt.Printf("  - %s\n", f)
			}
			fmt.Println()
		}
	}
}

func analyzeTestHelpers(files []FileSymbols) {
	fmt.Println("## Test Helper Functions (Scattered)")
	fmt.Println()

	var helpers []string

	for _, file := range files {
		if !strings.HasSuffix(file.FilePath, "_test.go") {
			continue
		}

		for _, symbol := range file.Symbols {
			name := strings.ToLower(symbol.Name)
			if (strings.Contains(name, "setup") ||
				strings.Contains(name, "helper") ||
				strings.Contains(name, "mock") ||
				strings.Contains(name, "stub") ||
				strings.Contains(name, "fixture") ||
				strings.Contains(name, "temp") ||
				strings.Contains(name, "cleanup")) &&
				!strings.HasPrefix(symbol.Name, "Test") {
				helpers = append(helpers, fmt.Sprintf("%s in %s:%d", symbol.Name, file.FilePath, symbol.Line))
			}
		}
	}

	if len(helpers) > 0 {
		fmt.Println("Test helper functions that could be centralized:")
		for _, h := range helpers {
			fmt.Printf("- %s\n", h)
		}
		fmt.Println()
	}
}

func analyzeMocksStubs(files []FileSymbols) {
	fmt.Println("## Mock/Stub Implementations")
	fmt.Println()

	var mocks []string

	for _, file := range files {
		for _, symbol := range file.Symbols {
			name := strings.ToLower(symbol.Name)
			if strings.Contains(name, "mock") || strings.Contains(name, "stub") || strings.Contains(name, "fake") {
				mocks = append(mocks, fmt.Sprintf("%s (%s) in %s:%d", symbol.Name, symbol.Kind, file.FilePath, symbol.Line))
			}
		}
	}

	if len(mocks) > 0 {
		fmt.Println("Mock/stub implementations found:")
		for _, m := range mocks {
			fmt.Printf("- %s\n", m)
		}
		fmt.Println()
		fmt.Println("**Recommendation**: Consider consolidating these into a shared testing package or using a mocking framework.")
		fmt.Println()
	}
}

func analyzeInitFunctions(files []FileSymbols) {
	fmt.Println("## Init Functions (Legacy Setup Code)")
	fmt.Println()

	var inits []string

	for _, file := range files {
		for _, symbol := range file.Symbols {
			if symbol.Name == "init" {
				inits = append(inits, fmt.Sprintf("%s:%d (package: %s)", file.FilePath, symbol.Line, file.Package))
			}
		}
	}

	fmt.Printf("Found %d init functions:\n", len(inits))
	for _, i := range inits {
		fmt.Printf("- %s\n", i)
	}
	fmt.Println()
	fmt.Println("**Note**: Init functions can make code harder to test and reason about. Consider explicit initialization where possible.")
	fmt.Println()
}

func analyzeSharedTypes(files []FileSymbols) {
	fmt.Println("## Types That Appear in Multiple Packages")
	fmt.Println()

	typeNames := make(map[string][]string)

	for _, file := range files {
		for _, symbol := range file.Symbols {
			if symbol.Kind == "type" && len(symbol.Name) > 0 && symbol.Name[0] >= 'A' && symbol.Name[0] <= 'Z' {
				// Look for common type patterns
				if strings.Contains(symbol.Name, "Config") ||
					strings.Contains(symbol.Name, "Options") ||
					strings.Contains(symbol.Name, "Request") ||
					strings.Contains(symbol.Name, "Response") ||
					strings.Contains(symbol.Name, "Result") ||
					strings.Contains(symbol.Name, "Error") ||
					strings.Contains(symbol.Name, "Info") {
					pkg := file.Package
					location := fmt.Sprintf("%s (%s:%d)", pkg, file.FilePath, symbol.Line)
					typeNames[symbol.Name] = append(typeNames[symbol.Name], location)
				}
			}
		}
	}

	fmt.Println("Common type patterns across packages (may indicate need for shared types):")
	for name, locs := range typeNames {
		if len(locs) > 1 {
			fmt.Printf("\n**%s** (%d occurrences):\n", name, len(locs))
			for _, loc := range locs {
				fmt.Printf("- %s\n", loc)
			}
		}
	}
	fmt.Println()
}

func analyzeLegacyPatterns(files []FileSymbols) {
	fmt.Println("## Potential Legacy/Deprecated Code Patterns")
	fmt.Println()

	legacyIndicators := []struct {
		pattern string
		reason  string
		matches []string
	}{
		{"New", "Constructor functions", []string{}},
		{"Helper", "Helper functions (often indicates scattered utility code)", []string{}},
		{"Util", "Utility functions (generic name, should be more specific)", []string{}},
		{"Tmp", "Temporary code", []string{}},
		{"Temp", "Temporary code", []string{}},
		{"Old", "Old/deprecated code", []string{}},
		{"Legacy", "Legacy code marker", []string{}},
		{"Deprecated", "Deprecated code", []string{}},
	}

	for _, file := range files {
		for _, symbol := range file.Symbols {
			for i := range legacyIndicators {
				if strings.Contains(symbol.Name, legacyIndicators[i].pattern) {
					location := fmt.Sprintf("%s in %s:%d", symbol.Name, file.FilePath, symbol.Line)
					legacyIndicators[i].matches = append(legacyIndicators[i].matches, location)
				}
			}
		}
	}

	for _, indicator := range legacyIndicators {
		if len(indicator.matches) > 3 && indicator.pattern != "New" { // Skip "New" pattern as it's too common
			fmt.Printf("### %s\n", indicator.reason)
			fmt.Printf("Pattern: '%s' (%d matches)\n\n", indicator.pattern, len(indicator.matches))
			for i, match := range indicator.matches {
				if i < 10 {
					fmt.Printf("- %s\n", match)
				}
			}
			if len(indicator.matches) > 10 {
				fmt.Printf("- ... and %d more\n", len(indicator.matches)-10)
			}
			fmt.Println()
		}
	}
}

func analyzeFragmentation(files []FileSymbols) {
	fmt.Println("## Over-Fragmented Code")
	fmt.Println()

	packageDirs := make(map[string]int)
	for _, file := range files {
		dir := file.FilePath
		// Get the directory path
		lastSlash := strings.LastIndex(dir, "/")
		if lastSlash != -1 {
			dir = dir[:lastSlash]
		}
		packageDirs[dir]++
	}

	fmt.Println("### Directories with many single-purpose files (>10 files):")
	for dir, count := range packageDirs {
		if count > 10 && !strings.Contains(dir, "_test") {
			fmt.Printf("- **%s**: %d files (consider consolidating related functionality)\n", dir, count)
		}
	}
	fmt.Println()

	// Count files with same prefix
	fmt.Println("### Files with common prefixes (candidates for consolidation):")
	prefixGroups := make(map[string][]string)

	for _, file := range files {
		fileName := file.FilePath
		lastSlash := strings.LastIndex(fileName, "/")
		if lastSlash != -1 {
			fileName = fileName[lastSlash+1:]
		}
		fileName = strings.TrimSuffix(fileName, ".go")
		fileName = strings.TrimSuffix(fileName, "_test")

		// Extract prefix (first word before underscore)
		parts := strings.Split(fileName, "_")
		if len(parts) > 1 {
			prefix := parts[0]
			dir := file.FilePath[:strings.LastIndex(file.FilePath, "/")]
			key := dir + "/" + prefix
			prefixGroups[key] = append(prefixGroups[key], file.FilePath)
		}
	}

	for prefix, group := range prefixGroups {
		if len(group) > 4 {
			fmt.Printf("\n**%s*** pattern (%d files):\n", prefix, len(group))
			for i, f := range group {
				if i < 8 {
					fmt.Printf("- %s\n", f)
				}
			}
			if len(group) > 8 {
				fmt.Printf("- ... and %d more\n", len(group)-8)
			}
		}
	}
	fmt.Println()
}
