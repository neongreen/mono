package parser

import (
"context"
"testing"
"time"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
ctx := context.Background()

parser, err := New(ctx)
require.NoError(t, err, "Failed to create parser")
require.NotNil(t, parser)
defer parser.Close(ctx)
}

func TestExtract_SimpleHTML(t *testing.T) {
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

parser, err := New(ctx)
require.NoError(t, err)
defer parser.Close(ctx)

html := "<html><head><title>Test Article</title></head><body><article><h1>Test Article Heading</h1><p>This is a test article with some content.</p></article></body></html>"

article, err := parser.Extract(ctx, "https://example.com/article", html)
require.NoError(t, err)
require.NotNil(t, article)

// Verify basic fields
assert.NotEmpty(t, article.Title)
assert.Equal(t, "example.com", article.Domain)
assert.NotEmpty(t, article.Content)
assert.Greater(t, article.WordCount, 0)

t.Logf("Parsed article: Title=%s, Domain=%s, WordCount=%d", article.Title, article.Domain, article.WordCount)
}

func TestExtract_EmptyURL(t *testing.T) {
ctx := context.Background()

parser, err := New(ctx)
require.NoError(t, err)
defer parser.Close(ctx)

_, err = parser.Extract(ctx, "", "<html><body>test</body></html>")
assert.Error(t, err)
assert.Contains(t, err.Error(), "url cannot be empty")
}

func TestExtract_EmptyHTML(t *testing.T) {
ctx := context.Background()

parser, err := New(ctx)
require.NoError(t, err)
defer parser.Close(ctx)

_, err = parser.Extract(ctx, "https://example.com", "")
assert.Error(t, err)
assert.Contains(t, err.Error(), "html cannot be empty")
}
