package project

import (
	"fmt"
	"os"
)

func getCurrentUser() (string, error) {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		return "", fmt.Errorf("could not determine current user")
	}
	return user, nil
}
