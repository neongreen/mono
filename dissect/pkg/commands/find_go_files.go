package commands

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// GoListPackage represents a single package entry from `go list -json` output.
type GoListPackage struct {
	Dir     string   `json:"Dir"`
	GoFiles []string `json:"GoFiles"`
}

// Find all Go files in the project
func FindGoFiles(moduleRoot string) ([]string, error) {
	// Call `go list -json`
	stdout, stderr, err := RunGoCommand("list", []string{"-test", "-json", "./..."}, moduleRoot, nil)
	if err != nil {
		return nil, fmt.Errorf("error running 'go list': %w\nStderr: %s", err, stderr)
	}

	var allGoFiles []string
	decoder := json.NewDecoder(strings.NewReader(stdout))
	for {
		var pkg GoListPackage
		if err := decoder.Decode(&pkg); err != nil {
			if err.Error() == "EOF" {
				break // End of input
			}
			return nil, fmt.Errorf("error decoding JSON from 'go list': %w", err)
		}
		for _, goFile := range pkg.GoFiles {
			// Skip files with absolute paths, since they are not in the module root
			if !filepath.IsAbs(goFile) {
				allGoFiles = append(allGoFiles, filepath.Join(pkg.Dir, goFile))
			}
		}
	}

	if len(allGoFiles) == 0 {
		return nil, fmt.Errorf("no Go files found in module root %s", moduleRoot)
	}
	// Sort the list of Go files for reproducibility
	sort.Strings(allGoFiles)

	return allGoFiles, nil
}
