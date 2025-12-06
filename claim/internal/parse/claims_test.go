package parse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClaims(t *testing.T) {
	t.Run("basic claim", func(t *testing.T) {
		content := `// @claim[test-1]: This is a test claim`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 1)

		c := claims[0]
		assert.Equal(t, "test-1", c.ID)
		assert.Equal(t, "This is a test claim", c.Statement)
	})

	t.Run("claim with tags", func(t *testing.T) {
		content := `// @claim[test-2] @pedantic: Claim with tag`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 1)

		assert.Equal(t, []string{"pedantic"}, claims[0].Tags)
	})

	t.Run("claim with proof", func(t *testing.T) {
		content := `// @claim[test-3]: Something is true
// @proof[test-3]:
// This is the proof body.
// It can span multiple lines.`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 1)

		c := claims[0]
		assert.Equal(t, "test-3", c.ID)
		assert.Equal(t, "Something is true", c.Statement)
		assert.Contains(t, c.Proof, "This is the proof body")
		assert.Contains(t, c.Proof, "It can span multiple lines")
	})

	t.Run("proof with @see references", func(t *testing.T) {
		content := `// @claim[test-4]: Depends on others
// @proof[test-4]:
// Given @see[other-claim] and @see[another-claim], this follows.`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 1)

		c := claims[0]
		assert.Equal(t, []string{"other-claim", "another-claim"}, c.SeeRefs)
	})

	t.Run("multiple @see same id deduped", func(t *testing.T) {
		content := `// @claim[test-5]: Uses same ref twice
// @proof[test-5]:
// First @see[foo], then again @see[foo].`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 1)

		// Should only have one reference to foo
		assert.Equal(t, []string{"foo"}, claims[0].SeeRefs)
	})

	t.Run("@see with multiple comma-separated refs", func(t *testing.T) {
		content := `// @claim[test-multi]: Uses multiple refs
// @proof[test-multi]:
// By @see[foo, bar, baz], all three are true.`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 1)

		assert.Equal(t, []string{"foo", "bar", "baz"}, claims[0].SeeRefs)
	})

	t.Run("@see multi-ref mixed with single refs", func(t *testing.T) {
		content := `// @claim[test-mixed]: Mixed refs
// @proof[test-mixed]:
// By @see[a, b] and @see[c], all are true.`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 1)

		assert.Equal(t, []string{"a", "b", "c"}, claims[0].SeeRefs)
	})

	t.Run("proof ends at next claim", func(t *testing.T) {
		content := `// @claim[a]: First claim
// @proof[a]:
// Proof for a.

// @claim[b]: Second claim`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 2)

		// Find claim a
		var claimA, claimB *Claim
		for i := range claims {
			if claims[i].ID == "a" {
				claimA = &claims[i]
			}
			if claims[i].ID == "b" {
				claimB = &claims[i]
			}
		}
		require.NotNil(t, claimA)
		require.NotNil(t, claimB)

		assert.Contains(t, claimA.Proof, "Proof for a")
		assert.NotContains(t, claimA.Proof, "Second claim")
	})

	t.Run("claim with context", func(t *testing.T) {
		content := `// @claim[test-ctx]: Something about a function
// @context[test-ctx]:
// function processEvents in handler.go
// lines 75-85 of worker.go
// @proof[test-ctx]:
// Given the context, we can see that X.`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 1)

		c := claims[0]
		assert.Equal(t, "test-ctx", c.ID)
		assert.Contains(t, c.Context, "function processEvents")
		assert.Contains(t, c.Context, "lines 75-85")
		assert.Contains(t, c.Proof, "Given the context")
	})

	t.Run("context ends at proof", func(t *testing.T) {
		content := `// @claim[test-end]: Test
// @context[test-end]:
// Some context here.
// @proof[test-end]:
// Proof here.`
		claims, err := ParseClaims(content, "test.go")
		require.NoError(t, err)
		require.Len(t, claims, 1)

		c := claims[0]
		assert.Contains(t, c.Context, "Some context here")
		assert.NotContains(t, c.Context, "Proof here")
		assert.Contains(t, c.Proof, "Proof here")
		assert.NotContains(t, c.Proof, "Some context")
	})
}

func TestStripCommentLeader(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"// @claim[id]: test", " @claim[id]: test"},
		{"  // @claim[id]: test", " @claim[id]: test"},
		{"# @claim[id]: test", " @claim[id]: test"},
		{"/* @claim[id]: test", " @claim[id]: test"},
		{"code @claim[id]", "code @claim[id]"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := StripCommentLeader(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
