package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type Session struct {
	SessionID    string    `json:"session_id"`
	Summary      string    `json:"summary,omitempty"`
	FilePath     string    `json:"file_path"`
	ProjectPath  string    `json:"project_path,omitempty"`
	MessageCount int       `json:"message_count"`
	ModTime      time.Time `json:"mod_time"`
}

func runSessions(cmd *cobra.Command, args []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	// Hardcoded trace locations
	paths := []string{
		filepath.Join(homeDir, ".claude", "projects"),
		filepath.Join(homeDir, ".claude", "debug"),
	}

	for _, basePath := range paths {
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Only process .jsonl files
			if !strings.HasSuffix(path, ".jsonl") {
				return nil
			}

			session, err := extractSession(path, info)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to extract session from %s: %v\n", path, err)
				return nil
			}

			// Output as JSONL
			data, err := json.Marshal(session)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to marshal session: %v\n", err)
				return nil
			}
			fmt.Println(string(data))

			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking path %s: %v\n", basePath, err)
			os.Exit(1)
		}
	}
}

func extractSession(path string, info os.FileInfo) (*Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	session := &Session{
		FilePath:     path,
		MessageCount: 0,
		ModTime:      info.ModTime(),
	}

	// Parse project path from file path
	// e.g., ~/.claude/projects/-Users-artyom-code-neongreen-mono/...
	if strings.Contains(path, ".claude/projects/") {
		parts := strings.Split(path, ".claude/projects/")
		if len(parts) >= 2 {
			subPath := parts[1]
			// Extract directory name which is the project path
			dirName := filepath.Dir(subPath)
			if dirName != "." {
				// Convert back to path format
				session.ProjectPath = strings.ReplaceAll(dirName, "-", "/")
			}
		}
	}

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines (trace files can have very long lines)
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		session.MessageCount++

		// Parse the line to extract metadata
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		// Extract session ID if we don't have it yet
		if session.SessionID == "" {
			if sid, ok := raw["sessionId"].(string); ok {
				session.SessionID = sid
			}
		}

		// Extract summary if we don't have it yet and this is a summary type
		if session.Summary == "" {
			if typ, ok := raw["type"].(string); ok && typ == "summary" {
				if summary, ok := raw["summary"].(string); ok {
					session.Summary = summary
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return session, nil
}
