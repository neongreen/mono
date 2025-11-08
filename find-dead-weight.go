package main

import (
	"encoding/json"
	"fmt"
	"os"
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

	fmt.Println("# Dead Weight Code Analysis")
	fmt.Println()
	fmt.Println("Focus: Useless code, duplicate logic, and maintenance burden")
	fmt.Println()

	// 1. Find near-duplicate functions (similar names, likely duplicate logic)
	findNearDuplicates(files)

	// 2. Find test utilities that are likely duplicates
	findDuplicateTestLogic(files)

	// 3. Find deprecated/old code patterns
	findDeprecatedCode(files)

	// 4. Find potential dead code (very specific names that suggest single use)
	findPotentialDeadCode(files)

	// 5. Find duplicate GitHub/HTTP logic
	findDuplicateGitHubLogic(files)
}

func findNearDuplicates(files []FileSymbols) {
	fmt.Println("## 1. Near-Duplicate Functions (Likely Duplicate Logic)")
	fmt.Println()
	fmt.Println("These functions have very similar names and likely contain duplicate or near-duplicate logic:")
	fmt.Println()

	// Group by normalized name (remove prefixes like Test, Run, etc.)
	nameGroups := make(map[string][]string)

	for _, file := range files {
		for _, symbol := range file.Symbols {
			if symbol.Kind != "func" && symbol.Kind != "method" {
				continue
			}

			// Normalize the name
			normalized := symbol.Name
			// Remove common prefixes for grouping
			for _, prefix := range []string{"Test", "test", "Run", "run", "Get", "get", "Find", "find", "Create", "create", "Build", "build"} {
				if strings.HasPrefix(normalized, prefix) {
					normalized = strings.TrimPrefix(normalized, prefix)
					break
				}
			}

			if len(normalized) < 3 {
				continue
			}

			location := fmt.Sprintf("%s in %s:%d", symbol.Name, file.FilePath, symbol.Line)
			nameGroups[normalized] = append(nameGroups[normalized], location)
		}
	}

	// Find groups with multiple functions
	type nameGroup struct {
		normalized string
		locations  []string
	}
	var groups []nameGroup
	for name, locs := range nameGroups {
		if len(locs) > 1 {
			groups = append(groups, nameGroup{name, locs})
		}
	}
	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i].locations) > len(groups[j].locations)
	})

	for i, group := range groups {
		if i >= 20 { // Only show top 20
			break
		}
		fmt.Printf("### '%s' variations (%d functions)\n", group.normalized, len(group.locations))
		for _, loc := range group.locations {
			fmt.Printf("- %s\n", loc)
		}
		fmt.Println()
	}
}

func findDuplicateTestLogic(files []FileSymbols) {
	fmt.Println("## 2. Duplicate Test Logic")
	fmt.Println()

	// Look for test helper functions with similar purposes
	setupFunctions := []string{}
	teardownFunctions := []string{}
	createFunctions := []string{}
	mockFunctions := []string{}

	for _, file := range files {
		if !strings.HasSuffix(file.FilePath, "_test.go") {
			continue
		}

		for _, symbol := range file.Symbols {
			nameLower := strings.ToLower(symbol.Name)
			location := fmt.Sprintf("%s in %s:%d", symbol.Name, file.FilePath, symbol.Line)

			if strings.Contains(nameLower, "setup") {
				setupFunctions = append(setupFunctions, location)
			}
			if strings.Contains(nameLower, "teardown") || strings.Contains(nameLower, "cleanup") {
				teardownFunctions = append(teardownFunctions, location)
			}
			if strings.HasPrefix(nameLower, "create") && !strings.HasPrefix(symbol.Name, "Test") {
				createFunctions = append(createFunctions, location)
			}
			if strings.Contains(nameLower, "mock") || strings.Contains(nameLower, "stub") || strings.Contains(nameLower, "fake") {
				mockFunctions = append(mockFunctions, location)
			}
		}
	}

	if len(setupFunctions) > 0 {
		fmt.Printf("### Setup Functions (%d total)\n", len(setupFunctions))
		fmt.Println("**LIKELY DUPLICATE LOGIC**: These should probably be consolidated into a shared test package")
		fmt.Println()
		for _, f := range setupFunctions {
			fmt.Printf("- %s\n", f)
		}
		fmt.Println()
	}

	if len(createFunctions) > 0 {
		fmt.Printf("### Create/Factory Functions in Tests (%d total)\n", len(createFunctions))
		fmt.Println("**LIKELY DUPLICATE LOGIC**: These create test fixtures and are probably duplicated")
		fmt.Println()
		for i, f := range createFunctions {
			if i < 30 {
				fmt.Printf("- %s\n", f)
			}
		}
		if len(createFunctions) > 30 {
			fmt.Printf("- ... and %d more\n", len(createFunctions)-30)
		}
		fmt.Println()
	}

	if len(mockFunctions) > 0 {
		fmt.Printf("### Mock/Stub Functions (%d total)\n", len(mockFunctions))
		fmt.Println("**MAINTENANCE BURDEN**: Consider using a mocking framework instead")
		fmt.Println()
		for _, f := range mockFunctions {
			fmt.Printf("- %s\n", f)
		}
		fmt.Println()
	}
}

func findDeprecatedCode(files []FileSymbols) {
	fmt.Println("## 3. Deprecated/Legacy Code Patterns")
	fmt.Println()

	deprecated := []string{}
	old := []string{}
	legacy := []string{}
	todo := []string{}
	hack := []string{}

	for _, file := range files {
		for _, symbol := range file.Symbols {
			nameLower := strings.ToLower(symbol.Name)
			docLower := strings.ToLower(symbol.Doc)
			location := fmt.Sprintf("%s (%s) in %s:%d", symbol.Name, symbol.Kind, file.FilePath, symbol.Line)

			if strings.Contains(nameLower, "deprecated") || strings.Contains(docLower, "deprecated") {
				deprecated = append(deprecated, location)
			}
			if strings.Contains(nameLower, "old") || strings.Contains(docLower, "old implementation") {
				old = append(old, location)
			}
			if strings.Contains(nameLower, "legacy") || strings.Contains(docLower, "legacy") {
				legacy = append(legacy, location)
			}
			if strings.Contains(docLower, "todo") || strings.Contains(docLower, "fixme") || strings.Contains(docLower, "hack") {
				todo = append(todo, location+" - "+strings.Split(symbol.Doc, "\n")[0])
			}
			if strings.Contains(nameLower, "hack") || strings.Contains(nameLower, "workaround") {
				hack = append(hack, location)
			}
		}
	}

	if len(deprecated) > 0 {
		fmt.Printf("### Explicitly Deprecated Code (%d items)\n", len(deprecated))
		fmt.Println("**ACTION**: Remove these entirely")
		fmt.Println()
		for _, d := range deprecated {
			fmt.Printf("- %s\n", d)
		}
		fmt.Println()
	}

	if len(old) > 0 {
		fmt.Printf("### 'Old' Code (%d items)\n", len(old))
		fmt.Println("**ACTION**: Verify these are still needed, remove if not")
		fmt.Println()
		for _, d := range old {
			fmt.Printf("- %s\n", d)
		}
		fmt.Println()
	}

	if len(legacy) > 0 {
		fmt.Printf("### Legacy Code (%d items)\n", len(legacy))
		fmt.Println("**ACTION**: Migrate or remove")
		fmt.Println()
		for _, d := range legacy {
			fmt.Printf("- %s\n", d)
		}
		fmt.Println()
	}

	if len(todo) > 0 {
		fmt.Printf("### TODOs/FIXMEs in Documentation (%d items)\n", len(todo))
		fmt.Println("**ACTION**: Address these technical debt items")
		fmt.Println()
		for _, d := range todo {
			fmt.Printf("- %s\n", d)
		}
		fmt.Println()
	}

	if len(hack) > 0 {
		fmt.Printf("### Hacks/Workarounds (%d items)\n", len(hack))
		fmt.Println("**ACTION**: Find proper solutions")
		fmt.Println()
		for _, d := range hack {
			fmt.Printf("- %s\n", d)
		}
		fmt.Println()
	}
}

func findPotentialDeadCode(files []FileSymbols) {
	fmt.Println("## 4. Potential Dead Code (Suspiciously Specific Names)")
	fmt.Println()
	fmt.Println("These symbols have very specific names suggesting they might only be used once or for a specific edge case:")
	fmt.Println()

	suspects := []string{}

	for _, file := range files {
		// Skip test files for this analysis
		if strings.HasSuffix(file.FilePath, "_test.go") {
			continue
		}

		for _, symbol := range file.Symbols {
			// Skip unexported symbols
			if len(symbol.Name) == 0 || symbol.Name[0] < 'A' || symbol.Name[0] > 'Z' {
				continue
			}

			nameLower := strings.ToLower(symbol.Name)

			// Look for overly specific names
			specificPatterns := []string{
				"bug", "fix", "temp", "tmp", "test", "debug", "experimental",
				"issue", "workaround", "patch", "special", "specific",
			}

			for _, pattern := range specificPatterns {
				if strings.Contains(nameLower, pattern) {
					location := fmt.Sprintf("%s (%s) in %s:%d", symbol.Name, symbol.Kind, file.FilePath, symbol.Line)
					if symbol.Doc != "" {
						location += " - " + strings.Split(strings.TrimSpace(symbol.Doc), "\n")[0]
					}
					suspects = append(suspects, location)
					break
				}
			}
		}
	}

	if len(suspects) > 0 {
		fmt.Printf("Found %d suspicious symbols:\n\n", len(suspects))
		for _, s := range suspects {
			fmt.Printf("- %s\n", s)
		}
		fmt.Println()
		fmt.Println("**ACTION**: Review each of these. If they're for specific bugs/workarounds, ensure they're still needed.")
	} else {
		fmt.Println("No obviously suspicious symbols found.")
	}
	fmt.Println()
}

func findDuplicateGitHubLogic(files []FileSymbols) {
	fmt.Println("## 5. Duplicate GitHub/HTTP Logic")
	fmt.Println()

	githubFunctions := []string{}
	httpFunctions := []string{}
	tokenFunctions := []string{}
	authFunctions := []string{}

	for _, file := range files {
		for _, symbol := range file.Symbols {
			if symbol.Kind != "func" && symbol.Kind != "method" {
				continue
			}

			nameLower := strings.ToLower(symbol.Name)
			location := fmt.Sprintf("%s in %s:%d", symbol.Name, file.FilePath, symbol.Line)

			if strings.Contains(nameLower, "github") {
				githubFunctions = append(githubFunctions, location)
			}
			if strings.Contains(nameLower, "http") || strings.Contains(nameLower, "request") || strings.Contains(nameLower, "client") {
				httpFunctions = append(httpFunctions, location)
			}
			if strings.Contains(nameLower, "token") {
				tokenFunctions = append(tokenFunctions, location)
			}
			if strings.Contains(nameLower, "auth") {
				authFunctions = append(authFunctions, location)
			}
		}
	}

	if len(githubFunctions) > 3 {
		fmt.Printf("### GitHub-related Functions (%d total)\n", len(githubFunctions))
		fmt.Println("**LIKELY DUPLICATE LOGIC**: Multiple implementations of GitHub API access")
		fmt.Println()
		for _, f := range githubFunctions {
			fmt.Printf("- %s\n", f)
		}
		fmt.Println()
	}

	if len(tokenFunctions) > 2 {
		fmt.Printf("### Token-related Functions (%d total)\n", len(tokenFunctions))
		fmt.Println("**LIKELY DUPLICATE LOGIC**: Token handling should be centralized")
		fmt.Println()
		for _, f := range tokenFunctions {
			fmt.Printf("- %s\n", f)
		}
		fmt.Println()
	}

	if len(authFunctions) > 2 {
		fmt.Printf("### Auth-related Functions (%d total)\n", len(authFunctions))
		fmt.Println("**LIKELY DUPLICATE LOGIC**: Authentication logic should be in one place")
		fmt.Println()
		for _, f := range authFunctions {
			fmt.Printf("- %s\n", f)
		}
		fmt.Println()
	}
}
