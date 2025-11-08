// Package parser provides Go bindings to Postlight Parser using WebAssembly via wazero.
//
// This package wraps the Postlight Parser JavaScript library (https://github.com/postlight/parser)
// and executes it through WebAssembly using the wazero runtime. This approach avoids CGO dependencies
// while providing access to the full Postlight Parser functionality.
//
// IMPORTANT: This implementation requires a pre-compiled WASM bundle of Postlight Parser.
// See BUILD.md for instructions on creating the WASM bundle. Currently, parser.wasm is a stub.
package parser

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed parser.wasm
var parserWasmBundle []byte

// Article represents the parsed content from a web page as returned by Postlight Parser.
// See https://github.com/postlight/parser#usage for complete field descriptions.
type Article struct {
	Title         string  `json:"title"`           // Article title
	Content       string  `json:"content"`         // Main article content (HTML by default)
	Author        *string `json:"author"`          // Article author
	DatePublished *string `json:"date_published"`  // Publication date (ISO 8601)
	LeadImageURL  *string `json:"lead_image_url"`  // URL of the lead image
	Dek           *string `json:"dek"`             // Article summary/description
	URL           string  `json:"url"`             // Original article URL
	Domain        string  `json:"domain"`          // Domain of the article
	Excerpt       *string `json:"excerpt"`         // Short excerpt
	WordCount     int     `json:"word_count"`      // Approximate word count
	Direction     string  `json:"direction"`       // Text direction (ltr/rtl)
	NextPageURL   *string `json:"next_page_url"`   // URL of next page (for paginated articles)
	TotalPages    int     `json:"total_pages"`     // Total pages in article
	RenderedPages int     `json:"rendered_pages"`  // Number of pages rendered
}

// Parser executes the Postlight Parser JavaScript library via WebAssembly.
type Parser struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
}

// New creates a new Parser instance that executes Postlight Parser via WASM.
//
// The parser must be closed with Close() when done to release resources.
//
// Returns an error if the WASM module cannot be compiled (e.g., if parser.wasm is a stub).
func New(ctx context.Context) (*Parser, error) {
	// Create wazero runtime
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI for WASM module support
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	// Compile the Postlight Parser WASM module
	compiledModule, err := r.CompileModule(ctx, parserWasmBundle)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to compile WASM module (parser.wasm may be a stub - see BUILD.md): %w", err)
	}

	return &Parser{
		runtime: r,
		module:  compiledModule,
	}, nil
}

// Close releases resources associated with the parser.
func (p *Parser) Close(ctx context.Context) error {
	var errs []error
	if p.module != nil {
		if err := p.module.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to close module: %w", err))
		}
	}
	if p.runtime != nil {
		if err := p.runtime.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to close runtime: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing parser: %v", errs)
	}
	return nil
}

// Extract parses HTML content using Postlight Parser and extracts article information.
//
// This method calls the Postlight Parser JavaScript library compiled to WASM. The actual
// Postlight Parser provides sophisticated article extraction including:
// - Smart content detection and cleaning
// - Author and date extraction
// - Lead image identification
// - Reading time calculation
// - Support for custom extractors
//
// Parameters:
//   - url: The URL of the article (required by Postlight Parser for context and relative link resolution)
//   - html: The HTML content to parse
//
// Returns the extracted article information or an error.
//
// Note: This implementation requires a proper parser.wasm bundle. See BUILD.md.
func (p *Parser) Extract(ctx context.Context, url string, html string) (*Article, error) {
	if url == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}
	if html == "" {
		return nil, fmt.Errorf("html cannot be empty")
	}

	// Instantiate the WASM module
	mod, err := p.runtime.InstantiateModule(ctx, p.module, wazero.NewModuleConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}
	defer mod.Close(ctx)

	// Create input JSON for Postlight Parser
	input := map[string]string{
		"url":  url,
		"html": html,
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Call the parse function in WASM
	// The WASM module should export a "parse" function that:
	// 1. Takes a pointer to input JSON and its length
	// 2. Calls Postlight Parser with the URL and HTML
	// 3. Returns a pointer to result JSON and its length
	parseFunc := mod.ExportedFunction("parse")
	if parseFunc == nil {
		return nil, fmt.Errorf("WASM module does not export 'parse' function - parser.wasm may be a stub (see BUILD.md)")
	}

	// TODO: Complete WASM function call implementation
	// This requires:
	// 1. Allocating memory in WASM for input
	// 2. Copying inputJSON to WASM memory
	// 3. Calling parseFunc with memory pointers
	// 4. Reading result from WASM memory
	// 5. Freeing allocated memory
	_ = inputJSON

	// Parse the result JSON into Article struct
	var article Article
	// TODO: Get resultJSON from WASM
	// err = json.Unmarshal(resultJSON, &article)

	return &article, fmt.Errorf("WASM execution not fully implemented - requires proper parser.wasm bundle with Postlight Parser (see BUILD.md)")
}
