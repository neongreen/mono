package a

import (
	"fmt"
	"os/user"
	"strings"
)

// Useless wrapper - should be detected
func getCurrentUser() (string, error) { // want "useless wrapper function"
	return getUser()
}

// Useless wrapper with package qualifier - should be detected
func getUserInfo() (*user.User, error) { // want "useless wrapper function"
	return user.Current()
}

// NOT a useless wrapper - has additional logic
func getCurrentUserWithDefault() (string, error) {
	username, err := getUser()
	if err != nil {
		return "unknown", nil
	}
	return username, nil
}

// NOT a useless wrapper - transforms the result
func getCurrentUserUpper() string {
	user, _ := getUser()
	return strings.ToUpper(user)
}

// NOT a useless wrapper - has multiple statements
func getCurrentUserWithLogging() (string, error) {
	fmt.Println("Getting current user")
	return getUser()
}

// NOT a useless wrapper - has different parameters
func getUserByID(id int) (string, error) {
	return getUser()
}

// NOT a useless wrapper - reorders or transforms parameters
func swapParams(a, b int) (int, int) {
	return doSwap(b, a)
}

// Method - should be skipped (not a standalone function)
type Helper struct{}

func (h *Helper) getCurrentUser() (string, error) {
	return getUser()
}

// Helper functions (implementation doesn't matter for this test)
func getUser() (string, error) {
	return "testuser", nil
}

func doSwap(x, y int) (int, int) {
	return x, y
}
