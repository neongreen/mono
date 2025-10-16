package git

import (
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitInfo struct {
	Hash           string
	Author         string
	AuthorEmail    string
	Committer      string
	CommitterEmail string
	Date           time.Time
	Message        string
	ParentHashes   []string
	Files          []FileInfo
}

type FileInfo struct {
	Path       string
	Size       int64
	Mode       string
	Content    []byte
	SHA256Hash string
}

type RefInfo struct {
	Type       string // "branch" or "tag"
	Name       string
	TargetHash string
}

type RemoteInfo struct {
	Name string
	URL  string
}

type RepoMetadata struct {
	Refs    []RefInfo
	Remotes []RemoteInfo
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

		// Get parent hashes
		var parentHashes []string
		for _, parent := range c.ParentHashes {
			parentHashes = append(parentHashes, parent.String())
		}

		commitInfo := CommitInfo{
			Hash:           c.Hash.String(),
			Author:         c.Author.Name,
			AuthorEmail:    c.Author.Email,
			Committer:      c.Committer.Name,
			CommitterEmail: c.Committer.Email,
			Date:           c.Author.When,
			Message:        c.Message,
			ParentHashes:   parentHashes,
			Files:          files,
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

// getFilesFromCommit retrieves all files in a commit with their contents
func getFilesFromCommit(commit *object.Commit) ([]FileInfo, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	err = tree.Files().ForEach(func(f *object.File) error {
		fileInfo := FileInfo{
			Path: f.Name,
			Size: f.Size,
			Mode: f.Mode.String(),
		}

		// Read file content
		reader, err := f.Reader()
		if err == nil {
			content, err := io.ReadAll(reader)
			reader.Close()
			if err == nil {
				fileInfo.Content = content
				// Calculate SHA256 hash
				hash := sha256.Sum256(content)
				fileInfo.SHA256Hash = fmt.Sprintf("%x", hash)
			}
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// GetRepoMetadata retrieves repository metadata (branches, tags, remotes)
func GetRepoMetadata(repoPath string) (*RepoMetadata, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	metadata := &RepoMetadata{}

	// Get all references (branches and tags)
	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.HashReference {
			return nil
		}

		refInfo := RefInfo{
			Name:       ref.Name().String(),
			TargetHash: ref.Hash().String(),
		}

		// Determine if it's a branch or tag
		if ref.Name().IsBranch() {
			refInfo.Type = "branch"
			refInfo.Name = ref.Name().Short()
		} else if ref.Name().IsTag() {
			refInfo.Type = "tag"
			refInfo.Name = ref.Name().Short()
		} else {
			// Other refs (like HEAD, remotes)
			refInfo.Type = "other"
		}

		metadata.Refs = append(metadata.Refs, refInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate references: %w", err)
	}

	// Get remotes
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}

	for _, remote := range remotes {
		config := remote.Config()
		for _, url := range config.URLs {
			metadata.Remotes = append(metadata.Remotes, RemoteInfo{
				Name: config.Name,
				URL:  url,
			})
		}
	}

	return metadata, nil
}
