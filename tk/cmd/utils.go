package cmd

import (
	"fmt"
	"os/user"

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

func getCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return currentUser.Username, nil
}
