package parse

import (
	"testing"
)

func TestParseClaims(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantLen    int
		checkClaim func(*testing.T, Claim)
	}{
		{
			name: "basic claim with bullets",
			content: `// @claim[test-1]: This is a test claim
// - first bullet
// - second bullet`,
			wantLen: 1,
			checkClaim: func(t *testing.T, c Claim) {
				if c.ID != "test-1" {
					t.Errorf("got ID %q, want %q", c.ID, "test-1")
				}
				if c.Statement != "This is a test claim" {
					t.Errorf("got statement %q", c.Statement)
				}
				if len(c.Bullets) != 2 {
					t.Errorf("got %d bullets, want 2", len(c.Bullets))
				}
			},
		},
		{
			name: "claim with tags",
			content: `// @claim[test-2] @pedantic: Claim with tag
// - bullet one`,
			wantLen: 1,
			checkClaim: func(t *testing.T, c Claim) {
				if len(c.Tags) != 1 {
					t.Fatalf("got %d tags, want 1", len(c.Tags))
				}
				if c.Tags[0] != "pedantic" {
					t.Errorf("got tag %q, want %q", c.Tags[0], "pedantic")
				}
			},
		},
		{
			name: "nested bullets",
			content: `// @claim[test-3]: Nested test
// - parent bullet
//   - child bullet
//   - another child
// - another parent`,
			wantLen: 1,
			checkClaim: func(t *testing.T, c Claim) {
				if len(c.Bullets) != 2 {
					t.Fatalf("got %d bullets, want 2", len(c.Bullets))
				}
				if len(c.Bullets[0].Children) != 2 {
					t.Errorf("parent bullet has %d children, want 2", len(c.Bullets[0].Children))
				}
				if c.Bullets[0].Path != "0" {
					t.Errorf("first bullet path is %q, want %q", c.Bullets[0].Path, "0")
				}
				if len(c.Bullets[0].Children) > 0 && c.Bullets[0].Children[0].Path != "0.0" {
					t.Errorf("first child path is %q, want %q", c.Bullets[0].Children[0].Path, "0.0")
				}
			},
		},
		{
			name: "sorry bullet",
			content: `// @claim[test-4]: Test sorry
// - normal bullet
// - @sorry`,
			wantLen: 1,
			checkClaim: func(t *testing.T, c Claim) {
				if len(c.Bullets) != 2 {
					t.Fatalf("got %d bullets, want 2", len(c.Bullets))
				}
				if !c.Bullets[1].IsSorry {
					t.Error("second bullet should be marked as sorry")
				}
				if c.Bullets[0].IsSorry {
					t.Error("first bullet should not be sorry")
				}
			},
		},
		{
			name: "bullet with claim reference",
			content: `// @claim[test-5]: Test refs
// - this depends on @claim[other-claim]
// - this is standalone`,
			wantLen: 1,
			checkClaim: func(t *testing.T, c Claim) {
				if len(c.Bullets) != 2 {
					t.Fatalf("got %d bullets, want 2", len(c.Bullets))
				}
				if len(c.Bullets[0].References) != 1 {
					t.Errorf("first bullet has %d refs, want 1", len(c.Bullets[0].References))
				}
				if len(c.Bullets[0].References) > 0 && c.Bullets[0].References[0] != "other-claim" {
					t.Errorf("got reference %q, want %q", c.Bullets[0].References[0], "other-claim")
				}
			},
		},
		{
			name: "TypeScript comment style",
			content: `// @claim[test-6]: TypeScript claim
// - first point
// - second point`,
			wantLen: 1,
			checkClaim: func(t *testing.T, c Claim) {
				if c.ID != "test-6" {
					t.Errorf("got ID %q, want %q", c.ID, "test-6")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseClaims(tt.content, "test.go")
			if err != nil {
				t.Fatalf("ParseClaims() error = %v", err)
			}
			if len(claims) != tt.wantLen {
				t.Fatalf("got %d claims, want %d", len(claims), tt.wantLen)
			}
			if tt.wantLen > 0 && tt.checkClaim != nil {
				tt.checkClaim(t, claims[0])
			}
		})
	}
}

func TestStripCommentLeader(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"// @claim[id]: test", " @claim[id]: test"},   // Space after // preserved
		{"  // @claim[id]: test", " @claim[id]: test"}, // Leading spaces removed, space after // preserved
		{"# @claim[id]: test", " @claim[id]: test"},    // Space after # preserved
		{"/* @claim[id]: test", " @claim[id]: test"},   // Space after /* preserved
		{"//   - bullet", "   - bullet"},               // Spaces after // preserved (for indentation)
		{"//     - nested", "     - nested"},           // Spaces after // preserved (for indentation)
		{"code @claim[id]", "code @claim[id]"},         // No comment prefix, return as-is
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripCommentLeader(tt.input)
			if got != tt.want {
				t.Errorf("stripCommentLeader(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
