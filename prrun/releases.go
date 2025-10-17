package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

var prReleaseTagPattern = regexp.MustCompile(`^(.+?)(?:--|/)pr-(\d+)\.(\d+)$`)

func parsePRReleaseTag(tag string) (project string, prNumber int, sequence int, ok bool) {
	matches := prReleaseTagPattern.FindStringSubmatch(tag)
	if matches == nil {
		return "", 0, 0, false
	}

	var err error
	prNumber, err = strconv.Atoi(matches[2])
	if err != nil {
		return "", 0, 0, false
	}

	sequence, err = strconv.Atoi(matches[3])
	if err != nil {
		return "", 0, 0, false
	}

	return matches[1], prNumber, sequence, true
}

func findPreviousReleaseTag(cacheDir, project string, prNumber, currentSequence int) (string, error) {
	dirEntries, err := os.ReadDir(cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read cache directory %s: %w", cacheDir, err)
	}

	latestSequence := -1
	latestTag := ""

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			continue
		}

		tag := entry.Name()
		entryProject, entryPRNumber, entrySequence, ok := parsePRReleaseTag(tag)
		if !ok {
			continue
		}
		if entryProject != project {
			continue
		}
		if entryPRNumber != prNumber {
			continue
		}
		if entrySequence >= currentSequence {
			continue
		}

		if entrySequence > latestSequence {
			latestSequence = entrySequence
			latestTag = tag
		}
	}

	return latestTag, nil
}
