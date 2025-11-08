// Package postlight provides Go bindings to the Postlight Parser library via WASM.
// It extracts clean, structured article content from web pages.
package postlight

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed parser.wasm
var parserWASM []byte

// Article represents the parsed article content returned by the parser.
type Article struct {
	// Title of the article
	Title string `json:"title"`

	// Author of the article
	Author string `json:"author,omitempty"`

	// DatePublished is the publication date
	DatePublished string `json:"date_published,omitempty"`

	// Dek is the article's excerpt/description
	Dek string `json:"dek,omitempty"`

	// LeadImageURL is the URL of the lead/hero image
	LeadImageURL string `json:"lead_image_url,omitempty"`

	// Content is the article's main content (HTML)
	Content string `json:"content"`

	// Excerpt is a short excerpt of the article
	Excerpt string `json:"excerpt,omitempty"`

	// WordCount is the number of words in the article
	WordCount int `json:"word_count,omitempty"`

	// Direction is the text direction (ltr/rtl)
	Direction string `json:"direction,omitempty"`

	// TotalPages for multi-page articles
	TotalPages int `json:"total_pages,omitempty"`

	// RenderedPages for multi-page articles
	RenderedPages int `json:"rendered_pages,omitempty"`

	// NextPageURL for multi-page articles
	NextPageURL string `json:"next_page_url,omitempty"`

	// URL is the original URL
	URL string `json:"url"`

	// Domain of the article
	Domain string `json:"domain"`
}

// ParseInput represents the input to the WASM parser.
type ParseInput struct {
	URL  string `json:"url"`
	HTML string `json:"html"`
}

// ParseResult represents the result from the WASM parser.
type ParseResult struct {
	Success bool     `json:"success"`
	Data    *Article `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Parser wraps the WASM module for parsing articles.
type Parser struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
}

// NewParser creates a new parser instance.
func NewParser(ctx context.Context) (*Parser, error) {
	// Create a new WebAssembly runtime
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI to support the WASM module
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	// Compile the WASM module
	compiledModule, err := r.CompileModule(ctx, parserWASM)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	return &Parser{
		runtime: r,
		module:  compiledModule,
	}, nil
}

// Parse extracts article content from HTML.
// The url parameter is used for context (extracting relative links, etc).
// The html parameter contains the HTML content to parse.
func (p *Parser) Parse(ctx context.Context, url, html string) (*Article, error) {
	// Prepare input
	input := ParseInput{
		URL:  url,
		HTML: html,
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Create buffers for stdin and stdout
	stdin := bytes.NewReader(inputJSON)
	var stdout bytes.Buffer

	// Configure the module
	config := wazero.NewModuleConfig().
		WithStdin(stdin).
		WithStdout(&stdout).
		WithStderr(io.Discard)

	// Instantiate and run the module
	instance, err := p.runtime.InstantiateModule(ctx, p.module, config)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}
	defer instance.Close(ctx)

	// Parse the result
	var result ParseResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w (output: %s)", err, stdout.String())
	}

	if !result.Success {
		return nil, fmt.Errorf("parser error: %s", result.Error)
	}

	if result.Data == nil {
		return nil, fmt.Errorf("parser returned no data")
	}

	return result.Data, nil
}

// ParseURL fetches and parses an article from a URL.
// This is a convenience method that handles the HTTP request.
func (p *Parser) ParseURL(ctx context.Context, url string) (*Article, error) {
	// Fetch the URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set a realistic User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PostlightParser/1.0)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Read the HTML
	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the HTML
	return p.Parse(ctx, url, string(htmlBytes))
}

// Close releases resources used by the parser.
func (p *Parser) Close(ctx context.Context) error {
	return p.runtime.Close(ctx)
}
