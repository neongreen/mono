package index

import (
	"fmt"

	"github.com/neongreen/mono/claim/internal/parse"
	"github.com/neongreen/mono/claim/internal/scan"
)

// Index contains all claims and lenses found in the scanned files
type Index struct {
	Claims    map[string]parse.Claim // Claim ID -> Claim
	Lenses    map[string]string      // Lens name -> Lens content
	Locations map[string][]Location  // For tracking duplicates
}

// Location represents where a claim was found
type Location struct {
	File string
	Line int
}

// Build creates an index from scanned files
func Build(files []scan.ScannedFile) (*Index, error) {
	idx := &Index{
		Claims:    make(map[string]parse.Claim),
		Lenses:    make(map[string]string),
		Locations: make(map[string][]Location),
	}

	for _, file := range files {
		content := string(file.Content)

		// Parse claims
		claims, err := parse.ParseClaims(content, file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse claims in %s: %w", file.Path, err)
		}

		for _, claim := range claims {
			// Track location for duplicate detection
			idx.Locations[claim.ID] = append(idx.Locations[claim.ID], Location{
				File: claim.File,
				Line: claim.Line,
			})

			// Store claim (last one wins if duplicate)
			idx.Claims[claim.ID] = claim
		}

		// Parse lenses
		lenses, err := parse.ParseLenses(content, file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse lenses in %s: %w", file.Path, err)
		}

		for name, content := range lenses {
			idx.Lenses[name] = content
		}
	}

	return idx, nil
}

// FindDuplicates returns claim IDs that appear in multiple locations
func (idx *Index) FindDuplicates() map[string][]Location {
	duplicates := make(map[string][]Location)

	for id, locations := range idx.Locations {
		if len(locations) > 1 {
			duplicates[id] = locations
		}
	}

	return duplicates
}

// GetClaim retrieves a claim by ID
func (idx *Index) GetClaim(id string) (parse.Claim, bool) {
	claim, ok := idx.Claims[id]
	return claim, ok
}

// ExpandReferences recursively collects all referenced claims up to maxDepth
func (idx *Index) ExpandReferences(claimID string, maxDepth int) (map[string]parse.Claim, error) {
	result := make(map[string]parse.Claim)
	visited := make(map[string]bool)

	var expand func(id string, depth int) error
	expand = func(id string, depth int) error {
		if depth > maxDepth {
			return nil
		}
		if visited[id] {
			return nil
		}
		visited[id] = true

		claim, ok := idx.GetClaim(id)
		if !ok {
			return fmt.Errorf("referenced claim %q not found", id)
		}

		result[id] = claim

		// Recursively expand references from bullets
		refs := collectReferences(claim.Bullets)
		for _, ref := range refs {
			if err := expand(ref, depth+1); err != nil {
				return err
			}
		}

		return nil
	}

	if err := expand(claimID, 0); err != nil {
		return nil, err
	}

	// Remove the original claim from results (caller already has it)
	delete(result, claimID)

	return result, nil
}

// collectReferences gathers all claim references from bullets
func collectReferences(bullets []parse.Bullet) []string {
	var refs []string
	for _, bullet := range bullets {
		refs = append(refs, bullet.References...)
		refs = append(refs, collectReferences(bullet.Children)...)
	}
	return refs
}
