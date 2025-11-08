package utils

import (
	"fmt"
	"os"
	"os/user"
)

// GetCurrentUser returns the username of the current user
// It tries multiple methods to determine the username:
// 1. os/user.Current() (works on most systems)
// 2. USER environment variable (common on Unix)
// 3. USERNAME environment variable (common on Windows)
func GetCurrentUser() (string, error) {
	// Try os/user.Current() first (most reliable)
	currentUser, err := user.Current()
	if err == nil && currentUser.Username != "" {
		return currentUser.Username, nil
	}

	// Fallback to environment variables
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if username == "" {
		return "", fmt.Errorf("could not determine current user")
	}
	return username, nil
}
