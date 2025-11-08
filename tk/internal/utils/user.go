package utils

import (
	"fmt"
	"os/user"
)

// GetCurrentUser returns the username of the current user
func GetCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return currentUser.Username, nil
}
