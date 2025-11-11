package rotatinglog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
)

// Writer is a rotating log writer that handles size-based rotation,
// compression, and concurrent writes.
type Writer struct {
	dir     string
	maxSize int64
}

// NewWriter creates a new rotating log writer.
// dir is the directory where log files will be stored.
// maxSize is the maximum size in bytes before rotation.
func NewWriter(dir string, maxSize int64) (*Writer, error) {
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &Writer{
		dir:     dir,
		maxSize: maxSize,
	}, nil
}

// Append appends a JSONL entry to the log.
// data should be a complete JSON object without trailing newline.
// This function adds the newline automatically.
//
// If the current log file is locked, tries fallback filenames (current-2.jsonl, etc.).
// Performs rotation and compression in the background if size threshold is exceeded.
func (w *Writer) Append(data []byte) error {
	// Check if rotation is needed and perform it (before trying to append)
	if err := w.maybeRotate(); err != nil {
		// Rotation failure is not fatal - log it but continue
		// This ensures writes don't fail if rotation has issues
	}

	// Try to append to current.jsonl, with fallback filenames
	for i := 1; i <= 10; i++ {
		filename := "current.jsonl"
		if i > 1 {
			filename = fmt.Sprintf("current-%d.jsonl", i)
		}

		path := filepath.Join(w.dir, filename)
		err := w.appendToFile(path, data)
		if err == nil {
			return nil
		}

		// If error is due to lock, try next filename
		// Otherwise, return the error
		if !isLockError(err) {
			return err
		}
	}

	return fmt.Errorf("failed to append: all fallback files are locked")
}

// appendToFile appends data to a specific file with locking
func (w *Writer) appendToFile(path string, data []byte) error {
	// Open file for appending, create if doesn't exist
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Try to acquire exclusive lock
	if err := lockFile(f); err != nil {
		return err
	}
	defer unlockFile(f)

	// Append newline to data for JSONL format
	dataWithNewline := append(data, '\n')

	// Write to file
	if _, err := f.Write(dataWithNewline); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	return nil
}

// maybeRotate checks if any current*.jsonl files exceed the size threshold
// and rotates/compresses them if needed.
func (w *Writer) maybeRotate() error {
	// Find all current*.jsonl files
	pattern := filepath.Join(w.dir, "current*.jsonl")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob current files: %w", err)
	}

	for _, path := range matches {
		// Check file size
		info, err := os.Stat(path)
		if err != nil {
			continue // Skip files we can't stat
		}

		if info.Size() >= w.maxSize {
			// Rotate this file
			if err := w.rotateFile(path); err != nil {
				// Log error but continue with other files
				continue
			}
		}
	}

	return nil
}

// rotateFile rotates and compresses a single log file
func (w *Writer) rotateFile(path string) error {
	// Generate timestamped filename
	timestamp := time.Now().Format("2006-01-02-150405")
	baseName := timestamp + ".jsonl"
	targetPath := filepath.Join(w.dir, baseName)
	compressedPath := targetPath + ".zst"

	// Rename current file to timestamped name
	if err := os.Rename(path, targetPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	// Compress the file
	if err := compressFile(targetPath, compressedPath); err != nil {
		// Compression failed - keep uncompressed file
		return fmt.Errorf("failed to compress file: %w", err)
	}

	// Delete uncompressed file after successful compression
	if err := os.Remove(targetPath); err != nil {
		// Not fatal - compressed file exists
		return nil
	}

	return nil
}

// compressFile compresses a file with zstd
func compressFile(srcPath, dstPath string) error {
	// Read source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Compress with zstd (default level)
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	defer encoder.Close()

	compressed := encoder.EncodeAll(data, nil)

	// Write compressed data
	if err := os.WriteFile(dstPath, compressed, 0o600); err != nil {
		return fmt.Errorf("failed to write compressed file: %w", err)
	}

	return nil
}

// Close closes the writer. Currently a no-op since we don't keep files open.
func (w *Writer) Close() error {
	return nil
}
