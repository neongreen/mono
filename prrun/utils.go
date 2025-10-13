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
	fmt.Println("prrun - Run binaries from GitHub PR releases")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  prrun <github-pr-url> [args...]")
	fmt.Println("  prrun <github-pr-url> --project <name> [args...]")
	fmt.Println("  prrun <github-pr-url> -p <name> [args...]")
	fmt.Println("  prrun --version")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <github-pr-url>  GitHub PR URL (e.g., github.com/owner/repo/pull/123)")
	fmt.Println("  --project, -p    Specify project name (required if multiple projects in PR)")
	fmt.Println("  --version, -v    Show version information and check for updates")
	fmt.Println("  [args...]        Arguments to pass to the binary (no -- separator needed)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Run dissect from PR #123")
	fmt.Println("  prrun https://github.com/neongreen/mono/pull/123 --project dissect")
	fmt.Println()
	fmt.Println("  # Run with arguments (no -- needed)")
	fmt.Println("  prrun github.com/neongreen/mono/pull/123 -p dissect --help")
	fmt.Println()
	fmt.Println("  # Run markdown-format on a file")
	fmt.Println("  prrun github.com/neongreen/mono/pull/123 -p markdown-format file.md")
	fmt.Println()
	fmt.Println("  # Auto-detect project (if only one project in PR)")
	fmt.Println("  prrun github.com/neongreen/mono/pull/123 --help")
	fmt.Println()
	fmt.Println("  # Old syntax still works (with -- separator)")
	fmt.Println("  prrun github.com/neongreen/mono/pull/123 -p dissect -- --help")
	fmt.Println()
	fmt.Println("The tool will:")
	fmt.Println("  1. Detect the project from PR releases (or use --project flag)")
	fmt.Println("  2. Download the binary to ~/.cache/prrun/")
	fmt.Println("  3. Run the binary with your arguments")
	fmt.Println("  4. Warn if the release workflow is pending approval")
}
