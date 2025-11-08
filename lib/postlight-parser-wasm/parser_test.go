package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	require.NotNil(t, parser)
	
	err = parser.Close()
	assert.NoError(t, err)
}

func TestExtract_EmptyURL(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	ctx := context.Background()
	_, err = parser.Extract(ctx, "", "<html><body>test</body></html>")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url cannot be empty")
}

func TestExtract_EmptyHTML(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	ctx := context.Background()
	_, err = parser.Extract(ctx, "https://example.com", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "html cannot be empty")
}

func TestExtract_SimpleHTML(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Test Article</title>
</head>
<body>
	<h1>Hello World</h1>
	<p>This is a test article with some content.</p>
</body>
</html>`
	
	ctx := context.Background()
	article, err := parser.Extract(ctx, "https://example.com/article", html)
	require.NoError(t, err)
	require.NotNil(t, article)
	
	assert.Equal(t, "Test Article", article.Title)
	assert.Equal(t, "https://example.com/article", article.URL)
	assert.Equal(t, "example.com", article.Domain)
	assert.Equal(t, "ltr", article.Direction)
	assert.Greater(t, article.WordCount, 0)
	assert.Contains(t, article.Content, "Hello World")
}

func TestExtract_ComplexHTML(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Advanced Article Title</title>
	<meta name="author" content="John Doe">
</head>
<body>
	<article>
		<h1>Main Heading</h1>
		<p>First paragraph with <strong>bold text</strong> and <em>italic text</em>.</p>
		<p>Second paragraph with <a href="https://example.com">a link</a>.</p>
		<img src="https://example.com/image.jpg" alt="Test image">
		<p>Third paragraph with more content to parse.</p>
	</article>
	<aside>
		<p>This is sidebar content that should be filtered out.</p>
	</aside>
</body>
</html>`
	
	ctx := context.Background()
	article, err := parser.Extract(ctx, "https://example.com/advanced", html)
	require.NoError(t, err)
	require.NotNil(t, article)
	
	assert.Equal(t, "Advanced Article Title", article.Title)
	assert.Equal(t, "https://example.com/advanced", article.URL)
	assert.Equal(t, "example.com", article.Domain)
	assert.Contains(t, article.Content, "Main Heading")
	assert.Greater(t, article.WordCount, 5)
}

func TestExtract_NoTitle(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	html := `<!DOCTYPE html>
<html>
<body>
	<p>Article without a title</p>
</body>
</html>`
	
	ctx := context.Background()
	article, err := parser.Extract(ctx, "https://example.com/notitle", html)
	require.NoError(t, err)
	require.NotNil(t, article)
	
	assert.Equal(t, "", article.Title)
	assert.Equal(t, "https://example.com/notitle", article.URL)
}

func TestExtract_WithSubdomain(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	html := `<html><head><title>Blog Post</title></head><body><p>Content</p></body></html>`
	
	ctx := context.Background()
	article, err := parser.Extract(ctx, "https://blog.example.com/post", html)
	require.NoError(t, err)
	
	assert.Equal(t, "blog.example.com", article.Domain)
}

func TestExtract_WithPort(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	html := `<html><head><title>Local Article</title></head><body><p>Test</p></body></html>`
	
	ctx := context.Background()
	article, err := parser.Extract(ctx, "http://localhost:8080/article", html)
	require.NoError(t, err)
	
	assert.Equal(t, "localhost:8080", article.Domain)
}

func TestExtractFromJSON_Valid(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	jsonData := []byte(`{
		"title": "Test Title",
		"content": "<p>Test content</p>",
		"url": "https://example.com",
		"domain": "example.com",
		"direction": "ltr",
		"word_count": 10
	}`)
	
	article, err := parser.ExtractFromJSON(jsonData)
	require.NoError(t, err)
	require.NotNil(t, article)
	
	assert.Equal(t, "Test Title", article.Title)
	assert.Equal(t, "<p>Test content</p>", article.Content)
	assert.Equal(t, "https://example.com", article.URL)
	assert.Equal(t, "example.com", article.Domain)
	assert.Equal(t, 10, article.WordCount)
}

func TestExtractFromJSON_WithOptionalFields(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	author := "Jane Doe"
	date := "2024-01-01T00:00:00Z"
	excerpt := "This is an excerpt"
	
	jsonData := []byte(`{
		"title": "Full Article",
		"content": "<p>Full content</p>",
		"author": "Jane Doe",
		"date_published": "2024-01-01T00:00:00Z",
		"excerpt": "This is an excerpt",
		"url": "https://example.com",
		"domain": "example.com",
		"direction": "ltr",
		"word_count": 20
	}`)
	
	article, err := parser.ExtractFromJSON(jsonData)
	require.NoError(t, err)
	require.NotNil(t, article)
	
	assert.Equal(t, "Full Article", article.Title)
	require.NotNil(t, article.Author)
	assert.Equal(t, author, *article.Author)
	require.NotNil(t, article.DatePublished)
	assert.Equal(t, date, *article.DatePublished)
	require.NotNil(t, article.Excerpt)
	assert.Equal(t, excerpt, *article.Excerpt)
}

func TestExtractFromJSON_Invalid(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	jsonData := []byte(`{invalid json}`)
	
	_, err = parser.ExtractFromJSON(jsonData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal article")
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"simple domain", "https://example.com", "example.com"},
		{"with path", "https://example.com/path", "example.com"},
		{"with subdomain", "https://blog.example.com", "blog.example.com"},
		{"with port", "http://localhost:8080", "localhost:8080"},
		{"with query", "https://example.com?query=test", "example.com"},
		{"invalid url", "not-a-url", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomain(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			"basic title",
			"<html><head><title>My Title</title></head></html>",
			"My Title",
		},
		{
			"title with whitespace",
			"<html><head><title>  Spaced Title  </title></head></html>",
			"  Spaced Title  ",
		},
		{
			"no title",
			"<html><head></head></html>",
			"",
		},
		{
			"unclosed title",
			"<html><head><title>Unclosed",
			"",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTitle(tt.html)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		minWords int
	}{
		{
			"simple text",
			"<p>Hello world</p>",
			2,
		},
		{
			"multiple paragraphs",
			"<p>First paragraph</p><p>Second paragraph with more words</p>",
			7,
		},
		{
			"with tags",
			"<p>Text with <strong>bold</strong> and <em>italic</em> words</p>",
			6,
		},
		{
			"empty",
			"",
			0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countWords(tt.html)
			assert.GreaterOrEqual(t, result, tt.minWords)
		})
	}
}

func TestExtract_RealWorldExample(t *testing.T) {
	parser, err := New()
	require.NoError(t, err)
	defer parser.Close()
	
	// Simulate a real-world blog post HTML
	html := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Understanding WebAssembly: A Practical Guide</title>
	<meta name="author" content="Tech Writer">
	<meta name="description" content="Learn about WebAssembly and its applications">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/blog">Blog</a>
		</nav>
	</header>
	<main>
		<article>
			<h1>Understanding WebAssembly: A Practical Guide</h1>
			<p class="author">By Tech Writer</p>
			<p class="date">Published on November 8, 2024</p>
			<img src="/images/wasm.jpg" alt="WebAssembly logo">
			<p>WebAssembly (Wasm) is a binary instruction format designed as a portable compilation target for programming languages. It enables deployment on the web for client and server applications.</p>
			<h2>Why WebAssembly?</h2>
			<p>WebAssembly provides near-native performance, allowing languages like C, C++, and Rust to run in the browser. This opens up new possibilities for web applications.</p>
			<h2>Use Cases</h2>
			<ul>
				<li>Game engines running in browsers</li>
				<li>Video and audio processing</li>
				<li>Scientific computing</li>
				<li>Cryptographic operations</li>
			</ul>
			<p>As we've seen, WebAssembly is a powerful technology that bridges the gap between native performance and web accessibility.</p>
		</article>
	</main>
	<footer>
		<p>&copy; 2024 Tech Blog</p>
	</footer>
</body>
</html>`
	
	ctx := context.Background()
	article, err := parser.Extract(ctx, "https://techblog.example.com/wasm-guide", html)
	require.NoError(t, err)
	require.NotNil(t, article)
	
	assert.Equal(t, "Understanding WebAssembly: A Practical Guide", article.Title)
	assert.Equal(t, "techblog.example.com", article.Domain)
	assert.Contains(t, article.Content, "WebAssembly")
	assert.Greater(t, article.WordCount, 50)
}

func TestNewWithWasm(t *testing.T) {
	ctx := context.Background()
	parser, err := NewWithWasm(ctx)
	require.NoError(t, err)
	require.NotNil(t, parser)
	
	// Verify parser is functional even with WASM runtime
	html := `<html><head><title>WASM Test</title></head><body><p>Content</p></body></html>`
	article, err := parser.Extract(ctx, "https://example.com", html)
	require.NoError(t, err)
	assert.Equal(t, "WASM Test", article.Title)
	
	err = parser.Close()
	assert.NoError(t, err)
}
