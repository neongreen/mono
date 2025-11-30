package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type Message struct {
	SessionID  string      `json:"session_id"`
	MessageID  string      `json:"message_id"`
	ParentID   string      `json:"parent_id,omitempty"`
	Timestamp  string      `json:"timestamp"`
	Type       string      `json:"type"`
	Content    interface{} `json:"content"`
	Model      string      `json:"model,omitempty"`
	CWD        string      `json:"cwd,omitempty"`
	ToolUseID  string      `json:"tool_use_id,omitempty"`
	IsError    bool        `json:"is_error,omitempty"`
	GitBranch  string      `json:"git_branch,omitempty"`
	Version    string      `json:"version,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
}

func runMessages(cmd *cobra.Command, args []string) {
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

			if err := extractMessages(path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to extract messages from %s: %v\n", path, err)
			}

			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking path %s: %v\n", basePath, err)
			os.Exit(1)
		}
	}
}

func extractMessages(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

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

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		message := extractMessage(raw)
		if message == nil {
			continue
		}

		// Output as JSONL
		data, err := json.Marshal(message)
		if err != nil {
			continue
		}
		fmt.Println(string(data))
	}

	return scanner.Err()
}

func extractMessage(raw map[string]interface{}) *Message {
	msg := &Message{}

	// Extract session ID
	if sid, ok := raw["sessionId"].(string); ok {
		msg.SessionID = sid
	}

	// Extract message UUID
	if uuid, ok := raw["uuid"].(string); ok {
		msg.MessageID = uuid
	}

	// Extract parent UUID
	if parentUUID, ok := raw["parentUuid"].(string); ok {
		msg.ParentID = parentUUID
	}

	// Extract timestamp
	if ts, ok := raw["timestamp"].(string); ok {
		msg.Timestamp = ts
	}

	// Extract type
	if typ, ok := raw["type"].(string); ok {
		msg.Type = typ
	}

	// Extract version
	if version, ok := raw["version"].(string); ok {
		msg.Version = version
	}

	// Extract git branch
	if branch, ok := raw["gitBranch"].(string); ok {
		msg.GitBranch = branch
	}

	// Extract CWD
	if cwd, ok := raw["cwd"].(string); ok {
		msg.CWD = cwd
	}

	// Extract request ID
	if reqID, ok := raw["requestId"].(string); ok {
		msg.RequestID = reqID
	}

	// Extract message content based on type
	if msgData, ok := raw["message"].(map[string]interface{}); ok {
		// Extract model for assistant messages
		if model, ok := msgData["model"].(string); ok {
			msg.Model = model
		}

		// Extract content
		if content, ok := msgData["content"]; ok {
			msg.Content = content

			// Check if this is a tool result
			if contentArr, ok := content.([]interface{}); ok && len(contentArr) > 0 {
				if firstBlock, ok := contentArr[0].(map[string]interface{}); ok {
					if toolUseID, ok := firstBlock["tool_use_id"].(string); ok {
						msg.ToolUseID = toolUseID
						msg.Type = "tool_result"
						if contentStr, ok := firstBlock["content"].(string); ok {
							msg.Content = contentStr
						}
						if isError, ok := firstBlock["is_error"].(bool); ok {
							msg.IsError = isError
						}
					}
				}
			}
		}
	}

	// Skip summary and other non-message types unless they're special
	if msg.Type != "user" && msg.Type != "assistant" && msg.Type != "tool_result" {
		return nil
	}

	return msg
}
