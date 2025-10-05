package commands

import (
	"fmt"
	"log/slog"
)

// RunGoimportsOnFile runs goimports on a single file
func RunGoimportsOnFile(filePath string) error {
	_, _, err := RunCommand("goimports", []string{"-w", filePath}, "", nil)
	if err != nil {
		slog.Error("Error running goimports", "file", filePath, "error", err)
		return fmt.Errorf("error running goimports on %s: %w", filePath, err)
	}
	return nil
}
