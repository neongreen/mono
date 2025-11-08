package cmd

import (
	"os"
	"path/filepath"

	"github.com/neongreen/mono/tk/internal/utils"
)

// collectSegmentFiles walks a directory and collects all .zst segment files
func collectSegmentFiles(dir string) ([]string, error) {
	var segmentFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".zst" {
			segmentFiles = append(segmentFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return segmentFiles, nil
}

// getCurrentUser is a package-local wrapper that calls utils.GetCurrentUser
func getCurrentUser() (string, error) {
	return utils.GetCurrentUser()
}
