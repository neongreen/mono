package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/neongreen/mono/tk/internal/database"
)

func OpenExistingDB() (*database.DB, error) {
	path, err := database.GetDBPath()
	if err != nil {
		return nil, err
	}

	db, err := database.OpenDB(path)
	if err != nil {
		return nil, err
	}

	if err := db.InitDB(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return db, nil
}

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

func getCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return currentUser.Username, nil
}
