package testutils

// ContainsString checks if a string contains a substring
func ContainsString(s string, substr string) bool {
	return len(s) >= len(substr) && FindSubstring(s, substr)
}
