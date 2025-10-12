package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(homeDir, ".cache", "prrun")
	return cacheDir, nil
}

func getGitHubToken() string {

	// getCacheDir returns the cache directory path
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}
	if token := os.Getenv("MISE_GITHUB_TOKEN"); token != "" {
		return token
	}
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}
	return ""
}

func createAuthenticatedRequest(method,
	url string) (*http.Request, error) {
	req, err := http.NewRequest(method,
		url, nil)
	if err !=

		nil {
		return nil, err
	}

	if token := getGitHubToken(); token != "" {
		req.Header.Set("Authorization",

			"Bearer "+
				token)
	}
	return req, nil
}

func runBinary(binaryPath string, args []string) error {
	cmd := exec.
		Command(binaryPath,

			args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
func printUsage() {
	fmt.Println("Usage: prrun <github-pr-url> [project-name] [-- args...]")
	fmt.Println()
	fmt.Println(
		"Examples:")
	fmt.
		Println("  prrun https://github.com/neongreen/mono/pull/123 dissect")
	fmt.
		Println("  prrun https://github.com/neongreen/mono/pull/123 dissect -- --help")
	fmt.Println("  prrun github.com/neongreen/mono/pull/123 markdown-format -- file.md")
	fmt.Println()
	fmt.Println("The tool will:")
	fmt.Println("  1. Find the GitHub release for the PR")
	fmt.Println("  2. Download the binary to ~/.cache/prrun/")
	fmt.Println("  3. Run the binary with any arguments after --")
}
