package invlog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
)

// MigrateLegacyLogs migrates old JSONL log files to the new invlogs directory.
// This should be called once on upgrade to consolidate historical logs.
//
// It looks for:
//   - ~/.tk/log.jsonl
//   - ~/.tk/log_*.jsonl
//
// And moves them to ~/.tk/invlogs/ with compression.
func MigrateLegacyLogs() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	tkDir := filepath.Join(home, ".tk")
	logDir, err := GetLogDir()
	if err != nil {
		return err
	}

	// Find all legacy log files
	patterns := []string{
		filepath.Join(tkDir, "log.jsonl"),
		filepath.Join(tkDir, "log_*.jsonl"),
	}

	var legacyFiles []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		legacyFiles = append(legacyFiles, matches...)
	}

	if len(legacyFiles) == 0 {
		// No legacy files to migrate
		return nil
	}

	// Migrate each file
	for _, srcPath := range legacyFiles {
		if err := migrateLegacyFile(srcPath, logDir); err != nil {
			// Log error but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: failed to migrate %s: %v\n", srcPath, err)
			continue
		}
	}

	return nil
}

// migrateLegacyFile migrates a single legacy JSONL file
func migrateLegacyFile(srcPath, logDir string) error {
	// Generate archive filename
	timestamp := time.Now().Format("2006-01-02-150405")
	baseName := filepath.Base(srcPath)
	// Use original filename in archive name for clarity
	archiveName := fmt.Sprintf("migrated-%s-%s.jsonl.zst", baseName, timestamp)
	dstPath := filepath.Join(logDir, archiveName)

	// Read source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Compress with zstd
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	defer encoder.Close()

	compressed := encoder.EncodeAll(data, nil)

	// Write compressed archive
	if err := os.WriteFile(dstPath, compressed, 0o600); err != nil {
		return fmt.Errorf("failed to write compressed archive: %w", err)
	}

	// Delete source file after successful migration
	if err := os.Remove(srcPath); err != nil {
		// Not fatal - compressed file exists
		return nil
	}

	return nil
}

// CountLegacyEntries counts entries in a legacy JSONL file for reporting
func CountLegacyEntries(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			var entry map[string]any
			if json.Unmarshal(line, &entry) == nil {
				count++
			}
		}
	}

	return count, scanner.Err()
}
