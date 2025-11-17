package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// MoveGroup represents a group of source files to be moved to a target
type MoveGroup struct {
	Sources []string // Source file paths (can include globs)
	Target  string   // Target path (directory or file)
}

// ParseBatchMoveArg parses a single arrow syntax argument like "a.go,b.go -> target/"
func ParseBatchMoveArg(arg string) (*MoveGroup, error) {
	// Split on " -> " (with spaces)
	parts := strings.Split(arg, " -> ")
	if len(parts) != 2 {
		if strings.Contains(arg, "->") {
			return nil, fmt.Errorf("invalid syntax: use ' -> ' (with spaces) to separate source and target")
		}
		return nil, fmt.Errorf("invalid syntax: expected 'source -> target' format")
	}

	sourcePart := strings.TrimSpace(parts[0])
	targetPart := strings.TrimSpace(parts[1])

	if sourcePart == "" {
		return nil, fmt.Errorf("invalid syntax: missing source files")
	}
	if targetPart == "" {
		return nil, fmt.Errorf("invalid syntax: missing target")
	}

	// Split sources on comma and trim each
	sourceStrs := strings.Split(sourcePart, ",")
	sources := make([]string, 0, len(sourceStrs))
	for _, s := range sourceStrs {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			sources = append(sources, trimmed)
		}
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("invalid syntax: no source files specified")
	}

	return &MoveGroup{
		Sources: sources,
		Target:  targetPart,
	}, nil
}

// ExpandGlobs expands glob patterns in source files
func ExpandGlobs(sources []string, baseDir string) ([]string, error) {
	var expanded []string
	seen := make(map[string]bool)

	for _, source := range sources {
		// If source contains glob characters, expand it
		if strings.ContainsAny(source, "*?[]") {
			// Convert to absolute path for globbing
			var pattern string
			if filepath.IsAbs(source) {
				pattern = source
			} else {
				pattern = filepath.Join(baseDir, source)
			}

			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid glob pattern %s: %w", source, err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("glob pattern %s matched no files", source)
			}

			for _, match := range matches {
				if !seen[match] {
					expanded = append(expanded, match)
					seen[match] = true
				}
			}
		} else {
			// Not a glob, use as-is
			var absPath string
			if filepath.IsAbs(source) {
				absPath = source
			} else {
				absPath = filepath.Join(baseDir, source)
			}

			if !seen[absPath] {
				expanded = append(expanded, absPath)
				seen[absPath] = true
			}
		}
	}

	return expanded, nil
}

// IsDirectory returns true if the target should be treated as a directory
func IsDirectory(target string) bool {
	// If it ends with a slash, it's explicitly a directory
	if strings.HasSuffix(target, "/") || strings.HasSuffix(target, "\\") {
		return true
	}
	// Otherwise, check filesystem (but this should be done by caller with os.Stat)
	return false
}
