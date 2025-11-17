package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/neongreen/mono/tk/internal/clock"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

const maxAttachmentSize = 50 * 1024 * 1024 // 50MB

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var attachCmd = &cobra.Command{
	Use:   "attach [task-id] [file]",
	Short: "Attach a file to a task",
	Long: `Attach a file to a task. Files are stored in the database and synced with events.

Examples:
  tk attach tk-123 screenshot.png --description "Login bug screenshot"
  tk attach tk-123 --list
  tk attach tk-123 --get att-1
  tk attach tk-123 --open att-1`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]

		listFlag, _ := cmd.Flags().GetBool("list")
		getFlag, _ := cmd.Flags().GetString("get")
		openFlag, _ := cmd.Flags().GetString("open")
		descriptionFlag, _ := cmd.Flags().GetString("description")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
		if err != nil {
			return err
		}

		// Handle --list flag
		if listFlag {
			return listAttachments(db, taskUUID, taskRef)
		}

		// Handle --get flag
		if getFlag != "" {
			return getAttachment(db, taskUUID, taskRef, getFlag)
		}

		// Handle --open flag
		if openFlag != "" {
			return openAttachment(db, taskUUID, taskRef, openFlag)
		}

		// Add attachment
		if len(args) < 2 {
			return fmt.Errorf("file path required (use --list to list attachments)")
		}

		filePath := args[1]
		return addAttachment(db, taskUUID, taskRef, filePath, descriptionFlag)
	},
}

func addAttachment(db *database.DB, taskUUID, taskRef, filePath, description string) error {
	// Check file exists and get info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot attach directory: %s", filePath)
	}

	if fileInfo.Size() > maxAttachmentSize {
		return fmt.Errorf("file too large: %d bytes (max %d bytes)", fileInfo.Size(), maxAttachmentSize)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Compute hash
	hash := sha256.Sum256(content)
	hashStr := "sha256:" + hex.EncodeToString(hash[:])

	// Detect MIME type
	mimeType := http.DetectContentType(content)

	// Get current user
	currentUser, err := utils.GetCurrentUser()
	if err != nil {
		return err
	}

	// Get current task state to generate attachment ID
	config, err := config_pkg.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reducer, err := db.GetCachedReducerWithConfig(config)
	if err != nil {
		return err
	}

	task, ok := reducer.Tasks()[taskUUID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskRef)
	}

	// Generate attachment ID based on count
	attachmentID := fmt.Sprintf("att-%d", len(task.Attachments)+1)

	// Insert attachment blob if it doesn't exist
	exists, err := db.AttachmentExists(hashStr)
	if err != nil {
		return err
	}
	if !exists {
		if err := db.InsertAttachment(hashStr, content, mimeType, currentUser); err != nil {
			return err
		}
	}

	// Add attachment to task via event
	filename := filepath.Base(filePath)
	if err := tasks.AddAttachment(db, taskUUID, attachmentID, hashStr, filename, description, mimeType, int64(len(content)), currentUser, &clock.RealClock{}); err != nil {
		return err
	}

	displayID, _ := database.RenderTaskDisplayID(db, taskUUID)
	if displayID == "" {
		displayID = taskRef
	}

	fmt.Printf("Attached %s to task %s as %s\n", filename, displayID, attachmentID)
	if description != "" {
		fmt.Printf("Description: %s\n", description)
	}
	return nil
}

func listAttachments(db *database.DB, taskUUID, taskRef string) error {
	// Get reducer with cached state
	config, err := config_pkg.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reducer, err := db.GetCachedReducerWithConfig(config)
	if err != nil {
		return err
	}

	task, ok := reducer.Tasks()[taskUUID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskRef)
	}

	if len(task.Attachments) == 0 {
		fmt.Printf("No attachments for task %s\n", taskRef)
		return nil
	}

	displayID, _ := database.RenderTaskDisplayID(db, taskUUID)
	if displayID == "" {
		displayID = taskRef
	}

	fmt.Printf("Attachments for task %s:\n", displayID)
	for _, att := range task.Attachments {
		sizeStr := formatSize(att.Size)
		fmt.Printf("  %s: %s (%s)\n", att.ID, att.Filename, sizeStr)
		if att.Description != "" {
			fmt.Printf("      %s\n", att.Description)
		}
	}
	return nil
}

func getAttachment(db *database.DB, taskUUID, taskRef, attachmentID string) error {
	// Get reducer with cached state
	config, err := config_pkg.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reducer, err := db.GetCachedReducerWithConfig(config)
	if err != nil {
		return err
	}

	task, ok := reducer.Tasks()[taskUUID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskRef)
	}

	// Find attachment by ID
	var attachment *types.Attachment
	for i := range task.Attachments {
		if task.Attachments[i].ID == attachmentID {
			attachment = &task.Attachments[i]
			break
		}
	}

	if attachment == nil {
		return fmt.Errorf("attachment not found: %s", attachmentID)
	}

	// Get attachment content
	content, _, err := db.GetAttachment(attachment.Hash)
	if err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(attachment.Filename, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Wrote %s (%d bytes)\n", attachment.Filename, len(content))
	return nil
}

func openAttachment(db *database.DB, taskUUID, taskRef, attachmentID string) error {
	// Get reducer with cached state
	config, err := config_pkg.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reducer, err := db.GetCachedReducerWithConfig(config)
	if err != nil {
		return err
	}

	task, ok := reducer.Tasks()[taskUUID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskRef)
	}

	// Find attachment by ID
	var attachment *types.Attachment
	for i := range task.Attachments {
		if task.Attachments[i].ID == attachmentID {
			attachment = &task.Attachments[i]
			break
		}
	}

	if attachment == nil {
		return fmt.Errorf("attachment not found: %s", attachmentID)
	}

	// Get attachment content
	content, _, err := db.GetAttachment(attachment.Hash)
	if err != nil {
		return err
	}

	// Write to temporary file
	tmpFile := filepath.Join(os.TempDir(), attachment.Filename)
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Open using platform-specific command
	var openCmd string
	switch {
	case fileExists("/usr/bin/open"): // macOS
		openCmd = "open"
	case fileExists("/usr/bin/xdg-open"): // Linux
		openCmd = "xdg-open"
	default:
		return fmt.Errorf("no suitable open command found (tried: open, xdg-open)")
	}

	// Execute open command
	if err := execCommand(openCmd, tmpFile); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	fmt.Printf("Opened %s (%s)\n", attachment.Filename, attachmentID)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func execCommand(command string, arg string) error {
	cmd := exec.Command(command, arg)
	return cmd.Start()
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)

	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
