package jobs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/ingest/pkg/database"
	ingfs "github.com/neongreen/mono/ingest/pkg/fs"
)

// RunFS ingests filesystem entries recursively.
func RunFS(ctx context.Context, out io.Writer, opts FSOptions) (Result, error) {
	if out == nil {
		out = os.Stdout
	}
	_ = ctx

	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return Result{}, fmt.Errorf("failed to resolve path %s: %w", opts.Path, err)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return Result{}, fmt.Errorf("path does not exist: %s", absPath)
	}

	fmt.Fprintf(out, "Ingesting filesystem: %s\n", absPath)

	db, err := database.Open()
	if err != nil {
		return Result{}, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	runID, err := db.CreateRun(absPath, "fs")
	if err != nil {
		return Result{}, fmt.Errorf("failed to create run: %w", err)
	}

	result := Result{RunID: runID}
	runStatus := "failed"
	defer func() {
		_ = db.FinishRun(runID, runStatus)
	}()

	fmt.Fprintf(out, "Started ingestion run #%d\n", runID)

	fmt.Fprintln(out, "Walking filesystem...")
	entries, err := ingfs.WalkFilesystemWithOptions(absPath, func(count int) {
		fmt.Fprintf(out, "Found %d entries so far...\r", count)
	}, ingfs.WalkOptions{RespectGitignore: opts.RespectGitignore})
	if err != nil {
		return Result{}, fmt.Errorf("failed to walk filesystem: %w", err)
	}

	fmt.Fprintf(out, "\nFound %d entries total\n", len(entries))

	totalBlobs := 0
	for i, entry := range entries {
		if (i+1)%100 == 0 || (i+1) == len(entries) {
			fmt.Fprintf(out, "Processing entry %d/%d...\r", i+1, len(entries))
		}

		var blobID *int64
		if len(entry.Content) > 0 {
			id, err := db.GetOrCreateBlob(entry.Content, entry.SHA256Hash)
			if err != nil {
				return Result{}, fmt.Errorf("failed to create blob for %s: %w", entry.Path, err)
			}
			blobID = &id
			totalBlobs++
		}

		if err := db.CreateFSEntry(runID, entry.Path, entry.IsDir, entry.Size, entry.Mode, entry.ModTime, blobID); err != nil {
			return Result{}, fmt.Errorf("failed to store entry %s: %w", entry.Path, err)
		}
	}

	if err := db.UpdateRunItemCount(runID); err != nil {
		return Result{}, fmt.Errorf("failed to update run item count: %w", err)
	}

	runStatus = "completed"
	result.ItemCount = len(entries)
	result.Details = map[string]int{
		"entries": len(entries),
		"blobs":   totalBlobs,
	}

	fmt.Fprintln(out, "Ingestion completed successfully!")
	return result, nil
}
