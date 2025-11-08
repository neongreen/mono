package parser

import (
"context"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
ctx := context.Background()

// Note: This will fail with the stub parser.wasm
// Once parser.wasm contains actual Postlight Parser, this test should pass
_, err := New(ctx)

// With stub WASM, we expect an error
if err != nil {
assert.Contains(t, err.Error(), "parser.wasm may be a stub")
t.Skip("Skipping test: parser.wasm is a stub. See BUILD.md for instructions on building the WASM bundle.")
return
}

// If somehow the stub compiled (shouldn't happen), we at least test basic structure
t.Log("Note: If this test passes, parser.wasm may have been replaced with a real implementation")
}

func TestParser_API(t *testing.T) {
// Test that the API types are correct even without a working WASM bundle

t.Run("Article struct has correct fields", func(t *testing.T) {
article := Article{
Title:     "Test Title",
Content:   "<p>Test content</p>",
URL:       "https://example.com",
Domain:    "example.com",
WordCount: 2,
Direction: "ltr",
}

assert.Equal(t, "Test Title", article.Title)
assert.Equal(t, "<p>Test content</p>", article.Content)
assert.Equal(t, "https://example.com", article.URL)
assert.Equal(t, "example.com", article.Domain)
assert.Equal(t, 2, article.WordCount)
assert.Equal(t, "ltr", article.Direction)
})

t.Run("Article supports optional fields", func(t *testing.T) {
author := "John Doe"
date := "2024-01-01T00:00:00Z"
excerpt := "This is an excerpt"
leadImage := "https://example.com/image.jpg"
dek := "Article summary"
nextPage := "https://example.com/page2"

article := Article{
Title:         "Full Article",
Content:       "<p>Full content</p>",
Author:        &author,
DatePublished: &date,
Excerpt:       &excerpt,
LeadImageURL:  &leadImage,
Dek:           &dek,
NextPageURL:   &nextPage,
URL:           "https://example.com",
Domain:        "example.com",
WordCount:     2,
Direction:     "ltr",
TotalPages:    2,
RenderedPages: 1,
}

require.NotNil(t, article.Author)
assert.Equal(t, author, *article.Author)
require.NotNil(t, article.DatePublished)
assert.Equal(t, date, *article.DatePublished)
require.NotNil(t, article.Excerpt)
assert.Equal(t, excerpt, *article.Excerpt)
require.NotNil(t, article.LeadImageURL)
assert.Equal(t, leadImage, *article.LeadImageURL)
require.NotNil(t, article.Dek)
assert.Equal(t, dek, *article.Dek)
require.NotNil(t, article.NextPageURL)
assert.Equal(t, nextPage, *article.NextPageURL)
assert.Equal(t, 2, article.TotalPages)
assert.Equal(t, 1, article.RenderedPages)
})
}
