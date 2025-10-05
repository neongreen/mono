package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runGoimportsOnDirectory runs goimports on all Go files in a given directory.
func RunGoimportsOnDirectory(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %w", dirPath, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
			err = RunGoimportsOnFile(filepath.Join(dirPath, file.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
