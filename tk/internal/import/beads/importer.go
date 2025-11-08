package beads

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
)

// Import imports beads issues from a file into tk
func Import(db *database.DB, opts ImportOptions) (*ImportResult, error) {
	// Read and parse beads file
	issues, err := ReadBeadsFile(opts.BeadsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read beads file: %w", err)
	}

	if len(issues) == 0 {
		return &ImportResult{}, nil
	}

	// Group by prefix
	prefixGroups := ExtractPrefixesFromBeads(issues)

	// Get current node
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return nil, fmt.Errorf("failed to get node ID: %w", err)
	}

	// Check for alias clashes before starting import
	for beadsPrefix := range prefixGroups {
		alias := opts.AliasPrefix + beadsPrefix
		clash, err := CheckAliasClash(db, alias, nodeID)
		if err != nil {
			return nil, err
		}
		if clash {
			return nil, fmt.Errorf("alias '%s' already exists for this node - cannot import (try a different --prefix)", alias)
		}
	}

	// If dry run, just return preview info
	if opts.DryRun {
		return &ImportResult{
			TotalImported: len(issues),
			ProjectsCreated: func() map[string]string {
				m := make(map[string]string)
				for prefix := range prefixGroups {
					m[prefix] = opts.AliasPrefix + prefix
				}
				return m
			}(),
		}, nil
	}

	// Import each prefix as a separate project
	totalImported := 0
	totalSkipped := 0
	issueMap := make(map[string]string) // beads ID -> tk task UID
	var renumberedIssues []string
	var failedNotes []string
	var failedRelationships []string
	projectsCreated := make(map[string]string) // prefix -> project UID

	// Get actor (will be used for all operations)
	// In the future we could make this configurable
	actor := "importer"

	for prefix, prefixIssues := range prefixGroups {
		// Create project for this prefix
		projectAlias := opts.AliasPrefix + prefix
		projectUID, err := CreateProjectForImport(db, prefix, projectAlias, actor)
		if err != nil {
			return nil, fmt.Errorf("failed to create project for prefix %s: %w", prefix, err)
		}
		projectsCreated[prefix] = projectUID

		// Pre-scan to find highest numeric ID (for fallback numbering)
		highestNumber := int64(0)
		for _, issue := range prefixIssues {
			if num, err := ParseBeadsNumber(issue.ID); err == nil {
				if num > highestNumber {
					highestNumber = num
				}
			}
		}

		// Import issues for this prefix
		for _, issue := range prefixIssues {
			// Try to parse number from beads ID
			number, err := ParseBeadsNumber(issue.ID)
			var renumbered bool

			if err != nil {
				// Fallback: assign next available number after highest
				highestNumber++
				number = highestNumber
				renumbered = true
			}

			taskUID, err := ImportBeadsIssue(db, issue, projectUID, number)
			if err != nil {
				// Skip this issue but continue with others
				totalSkipped++
				continue
			}

			// Map by original beads ID for relationships
			issueMap[issue.ID] = taskUID
			totalImported++

			// Add note explaining renumbering
			if renumbered {
				if err := AddRenumberNote(db, taskUID, issue.ID, number); err != nil {
					// Non-fatal, track and continue
					failedNotes = append(failedNotes, fmt.Sprintf("task %s: %v", issue.ID, err))
				}
				renumberedIssues = append(renumberedIssues,
					fmt.Sprintf("%s → %d (non-numeric ID)", issue.ID, number))
			}
		}
	}

	// Second pass: import relationships (across all prefixes)
	relImported := 0
	if totalImported > 0 {
		for _, issue := range issues {
			if taskUID, ok := issueMap[issue.ID]; ok {
				count, err := ImportBeadsRelationships(db, issue, taskUID, issueMap)
				if err != nil {
					// Non-fatal, track and continue
					failedRelationships = append(failedRelationships, fmt.Sprintf("task %s: %v", issue.ID, err))
					continue
				}
				relImported += count
			}
		}
	}

	return &ImportResult{
		TotalImported:       totalImported,
		TotalSkipped:        totalSkipped,
		RelationsImported:   relImported,
		RenumberedIssues:    renumberedIssues,
		ProjectsCreated:     projectsCreated,
		FailedNotes:         failedNotes,
		FailedRelationships: failedRelationships,
	}, nil
}
