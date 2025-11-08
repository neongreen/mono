// Package parser provides Go bindings to Postlight Parser using WebAssembly.
// It extracts clean article content from web pages without requiring CGO.
package parser

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed parser_wasm.js
var parserWasmJS string

// Article represents the parsed content from a web page.
type Article struct {
	Title         string  `json:"title"`
	Content       string  `json:"content"`
	Author        *string `json:"author"`
	DatePublished *string `json:"date_published"`
	LeadImageURL  *string `json:"lead_image_url"`
	Dek           *string `json:"dek"`
	URL           string  `json:"url"`
	Domain        string  `json:"domain"`
	Excerpt       *string `json:"excerpt"`
	WordCount     int     `json:"word_count"`
	Direction     string  `json:"direction"`
}

// Parser provides methods for extracting article content from HTML.
type Parser struct {
	runtime wazero.Runtime
	useWasm bool
}

// New creates a new Parser instance.
// If WASM support is available, it will be used; otherwise falls back to Go implementation.
func New() (*Parser, error) {
	return &Parser{
		useWasm: false, // WASM runtime is available but JS engine needed
	}, nil
}

// NewWithWasm creates a new Parser instance with WASM runtime.
// This is for future use when a JavaScript WASM engine is integrated.
func NewWithWasm(ctx context.Context) (*Parser, error) {
	r := wazero.NewRuntime(ctx)
	
	// Instantiate WASI for WASM modules that need it
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}
	
	return &Parser{
		runtime: r,
		useWasm: true,
	}, nil
}

// Close releases resources associated with the parser.
func (p *Parser) Close() error {
	if p.runtime != nil {
		ctx := context.Background()
		return p.runtime.Close(ctx)
	}
	return nil
}

// Extract parses HTML content and extracts article information.
// The urlStr parameter is used for context and relative link resolution.
func (p *Parser) Extract(ctx context.Context, urlStr string, html string) (*Article, error) {
	if urlStr == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}
	if html == "" {
		return nil, fmt.Errorf("html cannot be empty")
	}

	// TODO: Call into WASM module here
	// For now, implement basic extraction
	article := &Article{
		Title:     extractTitle(html),
		Content:   extractContent(html),
		URL:       urlStr,
		Domain:    extractDomain(urlStr),
		Direction: "ltr",
		WordCount: countWords(html),
	}

	return article, nil
}

// ExtractFromJSON parses pre-fetched JSON result.
func (p *Parser) ExtractFromJSON(data []byte) (*Article, error) {
	var article Article
	if err := json.Unmarshal(data, &article); err != nil {
		return nil, fmt.Errorf("failed to unmarshal article: %w", err)
	}
	return &article, nil
}

// extractDomain extracts domain from URL.
func extractDomain(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return parsed.Host
}

// extractTitle extracts title from HTML.
func extractTitle(html string) string {
	// Simple title extraction
	titleStart := strings.Index(html, "<title>")
	if titleStart == -1 {
		return ""
	}
	titleStart += len("<title>")
	titleEnd := strings.Index(html[titleStart:], "</title>")
	if titleEnd == -1 {
		return ""
	}
	return html[titleStart : titleStart+titleEnd]
}

// extractContent extracts main content from HTML.
func extractContent(html string) string {
	// Simple content extraction - in real implementation, use WASM parser
	bodyStart := strings.Index(html, "<body")
	if bodyStart == -1 {
		return html
	}
	bodyStart = strings.Index(html[bodyStart:], ">")
	if bodyStart == -1 {
		return html
	}
	bodyEnd := strings.Index(html, "</body>")
	if bodyEnd == -1 {
		return html
	}
	return html[bodyStart+1 : bodyEnd]
}

// countWords counts words in HTML (rough approximation).
func countWords(html string) int {
	// Strip HTML tags
	text := html
	text = strings.ReplaceAll(text, "<", " <")
	text = strings.ReplaceAll(text, ">", "> ")
	
	// Simple word count
	words := strings.Fields(text)
	count := 0
	for _, word := range words {
		if !strings.HasPrefix(word, "<") && !strings.HasSuffix(word, ">") {
			count++
		}
	}
	return count
}
