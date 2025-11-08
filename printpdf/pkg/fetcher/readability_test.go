package fetcher

import (
	"strings"
	"testing"
)

func TestExtractReadableContent(t *testing.T) {
	// Test with a typical blog post structure with navigation and sidebar
	html := `<!DOCTYPE html>
<html>
<head><title>Test Article</title></head>
<body>
	<nav class="navigation">
		<ul>
			<li><a href="/">Home</a></li>
			<li><a href="/about">About</a></li>
		</ul>
	</nav>
	<div class="sidebar">
		<h3>Recent Posts</h3>
		<ul>
			<li><a href="/post1">Post 1</a></li>
			<li><a href="/post2">Post 2</a></li>
		</ul>
	</div>
	<article class="post-content">
		<h1>Main Article Title</h1>
		<p>This is the first paragraph of the article with actual content.</p>
		<p>This is the second paragraph with more meaningful text.</p>
		<p>And a third paragraph to ensure we have enough content.</p>
	</article>
	<footer>
		<p>&copy; 2024 Example Site</p>
	</footer>
</body>
</html>`

	result, err := extractReadableContent([]byte(html))
	if err != nil {
		t.Fatalf("extractReadableContent failed: %v", err)
	}

	resultStr := string(result)

	// Check that main content is preserved
	if !strings.Contains(resultStr, "Main Article Title") {
		t.Error("Main article title should be preserved")
	}
	if !strings.Contains(resultStr, "first paragraph of the article") {
		t.Error("Article content should be preserved")
	}

	// Check that navigation is removed
	if strings.Contains(resultStr, "navigation") {
		t.Error("Navigation element should be removed")
	}

	// The result should be valid HTML
	if !strings.Contains(resultStr, "<!DOCTYPE html>") {
		t.Error("Result should be a complete HTML document")
	}
	if !strings.Contains(resultStr, "<html>") {
		t.Error("Result should have html tag")
	}
}

func TestExtractReadableContentWithScoringDivsVsSections(t *testing.T) {
	// Test that the scoring algorithm correctly identifies content
	html := `<!DOCTYPE html>
<html>
<body>
	<div class="ads">
		<p>Advertisement text with some filler</p>
		<a href="/ad">Click here</a>
	</div>
	<section class="article">
		<h2>Real Article Content</h2>
		<p>This is a real article with substantial content that should be selected.</p>
		<p>It has multiple paragraphs with meaningful information.</p>
		<p>And more content to ensure proper scoring.</p>
	</section>
	<div class="comments">
		<a href="/comment1">Comment 1</a>
		<a href="/comment2">Comment 2</a>
		<a href="/comment3">Comment 3</a>
	</div>
</body>
</html>`

	result, err := extractReadableContent([]byte(html))
	if err != nil {
		t.Fatalf("extractReadableContent failed: %v", err)
	}

	resultStr := string(result)

	// The article section should be selected due to higher score
	if !strings.Contains(resultStr, "Real Article Content") {
		t.Error("Article content should be selected")
	}
	if !strings.Contains(resultStr, "substantial content") {
		t.Error("Article paragraphs should be preserved")
	}
}

func TestExtractReadableContentWithUnlikelyCandidates(t *testing.T) {
	// Test that unlikely candidates are removed
	html := `<!DOCTYPE html>
<html>
<body>
	<div class="main-content">
		<div class="sponsor">
			<p>This is a sponsor message that should be removed</p>
		</div>
		<article>
			<h1>Article Title</h1>
			<p>This is the main article content that should be kept.</p>
			<p>More article content with details.</p>
		</article>
		<div class="modal">
			<p>Modal content to be removed</p>
		</div>
	</div>
</body>
</html>`

	result, err := extractReadableContent([]byte(html))
	if err != nil {
		t.Fatalf("extractReadableContent failed: %v", err)
	}

	resultStr := string(result)

	// Article content should be kept
	if !strings.Contains(resultStr, "Article Title") {
		t.Error("Article title should be preserved")
	}
	if !strings.Contains(resultStr, "main article content") {
		t.Error("Article content should be preserved")
	}
}

func TestExtractReadableContentWithHighLinkDensity(t *testing.T) {
	// Test that areas with high link density (like navigation) are penalized
	html := `<!DOCTYPE html>
<html>
<body>
	<div class="links">
		<a href="/1">Link 1</a>
		<a href="/2">Link 2</a>
		<a href="/3">Link 3</a>
		<a href="/4">Link 4</a>
	</div>
	<div class="content">
		<h1>Real Content</h1>
		<p>This is a paragraph with actual content and information.</p>
		<p>Another paragraph with more details about the topic.</p>
		<p>And a third paragraph to provide depth.</p>
		<p>You can <a href="/related">read more here</a> if interested.</p>
	</div>
</body>
</html>`

	result, err := extractReadableContent([]byte(html))
	if err != nil {
		t.Fatalf("extractReadableContent failed: %v", err)
	}

	resultStr := string(result)

	// Content with low link density should win
	if !strings.Contains(resultStr, "Real Content") {
		t.Error("Real content should be selected")
	}
	if !strings.Contains(resultStr, "actual content and information") {
		t.Error("Content paragraphs should be preserved")
	}
}

func TestExtractReadableContentFallback(t *testing.T) {
	// Test fallback to simple heuristics when scoring doesn't find good candidates
	html := `<!DOCTYPE html>
<html>
<body>
	<main>
		<h1>Simple Content</h1>
		<p>A simple paragraph.</p>
	</main>
</body>
</html>`

	result, err := extractReadableContent([]byte(html))
	if err != nil {
		t.Fatalf("extractReadableContent failed: %v", err)
	}

	resultStr := string(result)

	// Main tag should be recognized
	if !strings.Contains(resultStr, "Simple Content") {
		t.Error("Content in main tag should be preserved")
	}
}
