package readability

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// ExtractReadableContent extracts the main content from HTML
// This implements key parts of Mozilla's Readability algorithm
func ExtractReadableContent(htmlContent []byte) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Remove obviously non-content elements
	removeUnlikelyCandidates(doc)

	// Find the main content using readability scoring
	content := findBestContent(doc)
	if content == nil {
		// Fallback to simpler heuristics
		content = findMainContent(doc)
	}
	if content == nil {
		// Final fallback to body
		content = findBody(doc)
	}

	if content == nil {
		return htmlContent, nil // Return as-is if we can't find content
	}

	// Clean the content
	cleanContent(content)

	// Extract and clean the content
	var buf bytes.Buffer
	if err := html.Render(&buf, content); err != nil {
		return nil, fmt.Errorf("failed to render content: %w", err)
	}

	// Wrap in a clean HTML document
	cleanHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    max-width: 800px;
    margin: 40px auto;
    padding: 0 20px;
    line-height: 1.6;
    color: #24292e;
}
h1, h2, h3, h4, h5, h6 {
    margin-top: 24px;
    margin-bottom: 16px;
    font-weight: 600;
    line-height: 1.25;
}
p {
    margin-top: 0;
    margin-bottom: 16px;
}
img {
    max-width: 100%%;
    height: auto;
}
a {
    color: #0366d6;
    text-decoration: none;
}
a:hover {
    text-decoration: underline;
}
pre, code {
    background-color: #f6f8fa;
    border-radius: 3px;
    padding: 2px 4px;
    font-family: 'Courier New', monospace;
}
pre {
    padding: 16px;
    overflow: auto;
    line-height: 1.45;
}
blockquote {
    border-left: 4px solid #dfe2e5;
    padding-left: 16px;
    color: #6a737d;
    margin-left: 0;
}
ul, ol {
    padding-left: 2em;
}
li {
    margin-bottom: 0.5em;
}
</style>
</head>
<body>
%s
</body>
</html>`, buf.String())

	return []byte(cleanHTML), nil
}

// Regular expressions for removing unlikely candidates
var (
	unlikelyCandidatesRe   = regexp.MustCompile(`(?i)combx|comment|community|disqus|extra|foot|header|menu|modal|nav|remark|rss|share|shoutbox|sidebar|skyscraper|sponsor|ad-break|agegate|pagination|pager|popup`)
	okMaybeItsACandidateRe = regexp.MustCompile(`(?i)and|article|body|column|main|shadow`)
	positiveRe             = regexp.MustCompile(`(?i)article|body|content|entry|hentry|h-entry|main|page|pagination|post|text|blog|story`)
	negativeRe             = regexp.MustCompile(`(?i)hidden|^hid$|hid$|hid |^hid-|banner|combx|comment|com-|contact|foot|footer|footnote|masthead|media|meta|modal|outbrain|promo|related|scroll|share|shoutbox|sidebar|skyscraper|sponsor|shopping|tags|tool|widget`)
)

// removeUnlikelyCandidates removes elements that are unlikely to be content
func removeUnlikelyCandidates(n *html.Node) {
	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; {
			next := c.NextSibling
			removeUnlikelyCandidates(c)
			c = next
		}
		return
	}

	// Skip divs that might contain content
	if n.Data == "div" || n.Data == "section" || n.Data == "article" {
		for c := n.FirstChild; c != nil; {
			next := c.NextSibling
			removeUnlikelyCandidates(c)
			c = next
		}
		return
	}

	// Check ID and class for unlikely patterns
	matchString := getAttributeString(n, "class") + " " + getAttributeString(n, "id")
	if unlikelyCandidatesRe.MatchString(matchString) && !okMaybeItsACandidateRe.MatchString(matchString) {
		// Remove this node
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return
	}

	// Remove specific elements
	switch n.Data {
	case "script", "style", "noscript", "iframe", "object", "embed":
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return
	}

	// Recurse on children
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		removeUnlikelyCandidates(c)
		c = next
	}
}

// findBestContent finds the best content container using readability scoring
func findBestContent(n *html.Node) *html.Node {
	candidates := make(map[*html.Node]float64)
	findCandidates(n, candidates)

	// Find the best candidate
	var bestCandidate *html.Node
	bestScore := -1.0
	for node, score := range candidates {
		if score > bestScore {
			bestScore = score
			bestCandidate = node
		}
	}

	return bestCandidate
}

// findCandidates identifies and scores potential content containers
func findCandidates(n *html.Node, candidates map[*html.Node]float64) {
	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findCandidates(c, candidates)
		}
		return
	}

	// Only consider divs, sections, and articles as candidates
	if n.Data != "div" && n.Data != "section" && n.Data != "article" && n.Data != "main" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findCandidates(c, candidates)
		}
		return
	}

	// Score this element
	score := scoreElement(n)
	if score > 0 {
		candidates[n] = score
	}

	// Recurse on children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findCandidates(c, candidates)
	}
}

// scoreElement scores an element based on its content and structure
func scoreElement(n *html.Node) float64 {
	score := 0.0

	// Base score based on tag name
	switch n.Data {
	case "article":
		score += 10.0
	case "main":
		score += 10.0
	case "section":
		score += 5.0
	case "div":
		score += 5.0
	}

	// Score based on class and id
	matchString := getAttributeString(n, "class") + " " + getAttributeString(n, "id")
	if positiveRe.MatchString(matchString) {
		score += 25.0
	}
	if negativeRe.MatchString(matchString) {
		score -= 25.0
	}

	// Score based on content
	textLength := getTextLength(n)
	paragraphCount := countParagraphs(n)
	linkDensity := getLinkDensity(n)

	// More text is better
	score += math.Min(float64(textLength)/100.0, 20.0)

	// More paragraphs is better
	score += float64(paragraphCount) * 3.0

	// Lower link density is better
	score -= linkDensity * 25.0

	// Penalize if too short
	if textLength < 25 {
		score -= 20.0
	}

	return score
}

// getTextLength returns the total text length in an element
func getTextLength(n *html.Node) int {
	if n.Type == html.TextNode {
		return len(strings.TrimSpace(n.Data))
	}

	length := 0
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		length += getTextLength(c)
	}
	return length
}

// countParagraphs counts paragraph elements
func countParagraphs(n *html.Node) int {
	if n.Type == html.ElementNode && n.Data == "p" {
		return 1
	}

	count := 0
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		count += countParagraphs(c)
	}
	return count
}

// getLinkDensity calculates the ratio of link text to total text
func getLinkDensity(n *html.Node) float64 {
	textLength := getTextLength(n)
	if textLength == 0 {
		return 0
	}

	linkLength := getLinkTextLength(n)
	return float64(linkLength) / float64(textLength)
}

// getLinkTextLength returns the total length of text within links
func getLinkTextLength(n *html.Node) int {
	if n.Type == html.ElementNode && n.Data == "a" {
		return getTextLength(n)
	}

	length := 0
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		length += getLinkTextLength(c)
	}
	return length
}

// cleanContent removes unwanted elements from the content
func cleanContent(n *html.Node) {
	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; {
			next := c.NextSibling
			cleanContent(c)
			c = next
		}
		return
	}

	// Remove specific elements
	switch n.Data {
	case "script", "style", "noscript", "iframe", "object", "embed", "form":
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return
	}

	// Remove hidden elements
	if hasAttribute(n, "hidden") || getAttributeString(n, "style") != "" && strings.Contains(getAttributeString(n, "style"), "display:none") {
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return
	}

	// Remove navigation elements
	matchString := getAttributeString(n, "class") + " " + getAttributeString(n, "id") + " " + getAttributeString(n, "role")
	if n.Data == "nav" || strings.Contains(matchString, "nav") || strings.Contains(matchString, "menu") {
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return
	}

	// Recurse on children
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		cleanContent(c)
		c = next
	}
}

// getAttributeString gets an attribute value as a string
func getAttributeString(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// hasAttribute checks if an element has an attribute
func hasAttribute(n *html.Node, key string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return true
		}
	}
	return false
}

// findMainContent tries to find the main content element using simple heuristics
func findMainContent(n *html.Node) *html.Node {
	// Look for common content indicators
	if n.Type == html.ElementNode {
		// Check for semantic elements
		if n.Data == "main" || n.Data == "article" {
			return n
		}

		// Check for common content IDs and classes
		for _, attr := range n.Attr {
			if attr.Key == "id" || attr.Key == "class" {
				val := strings.ToLower(attr.Val)
				if strings.Contains(val, "main") ||
					strings.Contains(val, "content") ||
					strings.Contains(val, "article") ||
					strings.Contains(val, "post") {
					return n
				}
			}
		}
	}

	// Recursively search children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findMainContent(c); result != nil {
			return result
		}
	}

	return nil
}

// findBody finds the body element
func findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findBody(c); result != nil {
			return result
		}
	}

	return nil
}
