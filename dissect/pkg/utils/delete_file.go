package utils

import (
	"fmt"
	"os"
)

// DeleteFile deletes a file.
func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("error deleting file %s: %w", filePath, err)
	}
	return nil
}
