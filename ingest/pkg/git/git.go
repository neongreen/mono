package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitInfo struct {
	Hash        string
	Author      string
	AuthorEmail string
	Date        time.Time
	Message     string
	Files       []FileInfo
}

type FileInfo struct {
	Path string
	Size int64
	Mode string
}

// WalkRepository walks through all commits in a git repository
// The progressCallback is called periodically with the number of commits found so far
func WalkRepository(repoPath string, progressCallback func(int)) ([]CommitInfo, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get all commit objects in the repository
	commitIter, err := repo.CommitObjects()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit objects: %w", err)
	}

	var commits []CommitInfo
	count := 0
	err = commitIter.ForEach(func(c *object.Commit) error {
		files, err := getFilesFromCommit(c)
		if err != nil {
			// If we can't get files for this commit, just continue with empty files
			files = []FileInfo{}
		}

		commitInfo := CommitInfo{
			Hash:        c.Hash.String(),
			Author:      c.Author.Name,
			AuthorEmail: c.Author.Email,
			Date:        c.Author.When,
			Message:     c.Message,
			Files:       files,
		}

		commits = append(commits, commitInfo)
		count++
		
		// Call progress callback every 100 commits or on first commit
		if progressCallback != nil && (count%100 == 0 || count == 1) {
			progressCallback(count)
		}
		
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// getFilesFromCommit retrieves all files in a commit
func getFilesFromCommit(commit *object.Commit) ([]FileInfo, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	err = tree.Files().ForEach(func(f *object.File) error {
		files = append(files, FileInfo{
			Path: f.Name,
			Size: f.Size,
			Mode: f.Mode.String(),
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
