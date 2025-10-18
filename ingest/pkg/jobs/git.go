package jobs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"ingest/pkg/database"
	"ingest/pkg/git"
)

// RunGit ingests a git repository and stores metadata in the database.
func RunGit(ctx context.Context, out io.Writer, opts GitOptions) (Result, error) {
	if out == nil {
		out = os.Stdout
	}
	_ = ctx

	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return Result{}, fmt.Errorf("failed to resolve path %s: %w", opts.Path, err)
	}

	if _, err := os.Stat(filepath.Join(absPath, ".git")); os.IsNotExist(err) {
		return Result{}, fmt.Errorf("not a git repository: %s", absPath)
	}

	fmt.Fprintf(out, "Ingesting repository: %s\n", absPath)

	db, err := database.Open()
	if err != nil {
		return Result{}, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	runID, err := db.CreateRun(absPath, "git")
	if err != nil {
		return Result{}, fmt.Errorf("failed to create run: %w", err)
	}

	result := Result{RunID: runID}
	runStatus := "failed"
	defer func() {
		_ = db.FinishRun(runID, runStatus)
	}()

	fmt.Fprintf(out, "Started ingestion run #%d\n", runID)

	fmt.Fprintln(out, "Collecting repository metadata...")
	metadata, err := git.GetRepoMetadata(absPath)
	if err != nil {
		return Result{}, fmt.Errorf("failed to get repository metadata: %w", err)
	}

	for _, remote := range metadata.Remotes {
		if err := db.CreateGitRemote(runID, remote.Name, remote.URL); err != nil {
			return Result{}, fmt.Errorf("failed to store remote %s: %w", remote.Name, err)
		}
	}
	for _, ref := range metadata.Refs {
		if err := db.CreateGitRef(runID, ref.Type, ref.Name, ref.TargetHash); err != nil {
			return Result{}, fmt.Errorf("failed to store ref %s: %w", ref.Name, err)
		}
	}
	fmt.Fprintf(out, "Found %d remotes and %d refs\n", len(metadata.Remotes), len(metadata.Refs))

	fmt.Fprintln(out, "Looking for commits...")
	commits, err := git.WalkRepository(absPath, func(count int) {
		fmt.Fprintf(out, "Found %d commits so far...\r", count)
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to walk repository: %w", err)
	}
	fmt.Fprintf(out, "\nFound %d commits total\n", len(commits))

	totalFiles := 0
	totalBlobs := 0
	for i, commit := range commits {
		if (i+1)%100 == 0 || (i+1) == len(commits) {
			fmt.Fprintf(out, "Processing commit %d/%d...\r", i+1, len(commits))
		}

		commitID, err := db.CreateCommit(
			runID,
			commit.Hash,
			commit.Author,
			commit.AuthorEmail,
			commit.Committer,
			commit.CommitterEmail,
			commit.Date,
			commit.Message,
			commit.ParentHashes,
		)
		if err != nil {
			return Result{}, fmt.Errorf("failed to store commit %s: %w", commit.Hash, err)
		}

		for _, file := range commit.Files {
			var blobID *int64
			if len(file.Content) > 0 {
				id, err := db.GetOrCreateBlob(file.Content, file.SHA256Hash)
				if err != nil {
					return Result{}, fmt.Errorf("failed to create blob for %s: %w", file.Path, err)
				}
				blobID = &id
				totalBlobs++
			}
			if err := db.CreateFile(commitID, file.Path, file.Size, file.Mode, blobID); err != nil {
				return Result{}, fmt.Errorf("failed to store file %s: %w", file.Path, err)
			}
			totalFiles++
		}
	}

	fmt.Fprintf(out, "\nProcessed %d commits with %d files and %d blobs\n", len(commits), totalFiles, totalBlobs)

	if err := db.UpdateRunItemCount(runID); err != nil {
		return Result{}, fmt.Errorf("failed to update run item count: %w", err)
	}

	runStatus = "completed"
	result.ItemCount = len(commits)
	result.Details = map[string]int{
		"commits": len(commits),
		"files":   totalFiles,
		"blobs":   totalBlobs,
	}

	fmt.Fprintln(out, "Ingestion completed successfully!")
	return result, nil
}
