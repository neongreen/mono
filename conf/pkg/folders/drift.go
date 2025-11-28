package folders

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileStatus represents the status of a file in drift detection
type FileStatus string

const (
	StatusInSync   FileStatus = "IN_SYNC"  // File content matches
	StatusModified FileStatus = "MODIFIED" // File exists in both, content differs
	StatusAdded    FileStatus = "ADDED"    // File in source, not in conf copy
	StatusDeleted  FileStatus = "DELETED"  // File in conf copy, not in source
)

// FileDrift represents a difference between source and conf copy
type FileDrift struct {
	RelPath     string // Relative path from folder root
	Status      FileStatus
	SourceHash  string // SHA256 hash of source file (if exists)
	ConfHash    string // SHA256 hash of conf file (if exists)
	SourceMtime int64  // Modification time of source file
	ConfMtime   int64  // Modification time of conf file
	IsDir       bool   // Whether this is a directory
}

// DetectDrift compares source folder with conf copy and returns all differences.
// This is a convenience wrapper that calls DetectDriftWithExcludes with no excludes.
func DetectDrift(sourcePath, confPath string) ([]FileDrift, error) {
	return DetectDriftWithExcludes(sourcePath, confPath, nil)
}

// DetectDriftWithExcludes compares source folder with conf copy and returns all differences,
// excluding files that match any of the provided patterns.
// Patterns support shell-style wildcards (*, ?) via filepath.Match.
func DetectDriftWithExcludes(sourcePath, confPath string, excludePatterns []string) ([]FileDrift, error) {
	var drifts []FileDrift

	// Build maps of all files in source and conf
	sourceFiles := make(map[string]os.FileInfo)
	confFiles := make(map[string]os.FileInfo)

	// Walk source directory
	if err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if relPath == "." {
			return nil
		}
		// Skip excluded files
		if shouldExclude(relPath, info.Name(), excludePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		sourceFiles[relPath] = info
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk source directory: %w", err)
	}

	// Walk conf directory
	if err := filepath.Walk(confPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(confPath, path)
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if relPath == "." {
			return nil
		}
		// Skip excluded files
		if shouldExclude(relPath, info.Name(), excludePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		confFiles[relPath] = info
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk conf directory: %w", err)
	}

	// Collect all unique paths
	allPaths := make(map[string]bool)
	for path := range sourceFiles {
		allPaths[path] = true
	}
	for path := range confFiles {
		allPaths[path] = true
	}

	// Compare each path
	for relPath := range allPaths {
		sourceInfo, inSource := sourceFiles[relPath]
		confInfo, inConf := confFiles[relPath]

		drift := FileDrift{
			RelPath: relPath,
		}

		if inSource && !inConf {
			// File exists in source but not in conf
			drift.Status = StatusAdded
			drift.IsDir = sourceInfo.IsDir()
			if !sourceInfo.IsDir() {
				drift.SourceHash, _ = computeFileHash(filepath.Join(sourcePath, relPath))
				drift.SourceMtime = sourceInfo.ModTime().Unix()
			}
			drifts = append(drifts, drift)
			continue
		}

		if !inSource && inConf {
			// File exists in conf but not in source
			drift.Status = StatusDeleted
			drift.IsDir = confInfo.IsDir()
			if !confInfo.IsDir() {
				drift.ConfHash, _ = computeFileHash(filepath.Join(confPath, relPath))
				drift.ConfMtime = confInfo.ModTime().Unix()
			}
			drifts = append(drifts, drift)
			continue
		}

		// File exists in both
		drift.IsDir = sourceInfo.IsDir()

		// Skip directories - we only care about file content differences
		if sourceInfo.IsDir() {
			continue
		}

		// Compare file hashes
		sourceFullPath := filepath.Join(sourcePath, relPath)
		confFullPath := filepath.Join(confPath, relPath)

		sourceHash, err := computeFileHash(sourceFullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to hash source file %s: %w", relPath, err)
		}

		confHash, err := computeFileHash(confFullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to hash conf file %s: %w", relPath, err)
		}

		drift.SourceHash = sourceHash
		drift.ConfHash = confHash
		drift.SourceMtime = sourceInfo.ModTime().Unix()
		drift.ConfMtime = confInfo.ModTime().Unix()

		if sourceHash == confHash {
			drift.Status = StatusInSync
			// Don't include in-sync files in drift list by default
			continue
		} else {
			drift.Status = StatusModified
			drifts = append(drifts, drift)
		}
	}

	return drifts, nil
}

// computeFileHash computes SHA256 hash of a file
func computeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// FormatDriftSummary returns a human-readable summary of drift
func FormatDriftSummary(drifts []FileDrift) string {
	if len(drifts) == 0 {
		return "No drift detected"
	}

	counts := make(map[FileStatus]int)
	for _, drift := range drifts {
		counts[drift.Status]++
	}

	var parts []string
	if count := counts[StatusModified]; count > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", count))
	}
	if count := counts[StatusAdded]; count > 0 {
		parts = append(parts, fmt.Sprintf("%d added", count))
	}
	if count := counts[StatusDeleted]; count > 0 {
		parts = append(parts, fmt.Sprintf("%d deleted", count))
	}

	return fmt.Sprintf("%d files with drift (%s)", len(drifts), strings.Join(parts, ", "))
}

// shouldExclude returns true if a file path matches any of the exclude patterns.
// Patterns are matched against both the full relative path and the base filename.
// Supports shell-style wildcards (* and ?) via filepath.Match.
func shouldExclude(relPath, baseName string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		// Match against base filename (e.g., "*.tmp" matches "foo.tmp")
		if matched, _ := filepath.Match(pattern, baseName); matched {
			return true
		}
		// Match against full relative path (e.g., "subdir/*.log")
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
	}
	return false
}
