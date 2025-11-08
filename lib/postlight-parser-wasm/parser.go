// Package parser provides Go bindings to Postlight Parser using WebAssembly via wazero.
//
// This package wraps the Postlight Parser JavaScript library (https://github.com/postlight/parser)
// and executes it through WebAssembly using the wazero runtime. This approach avoids CGO dependencies
// while providing access to the full Postlight Parser functionality.
package parser

import (
	"bytes"
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
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
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
// This method calls the Postlight Parser JavaScript library compiled to WASM via Javy.
// The Postlight Parser provides sophisticated article extraction including:
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
func (p *Parser) Extract(ctx context.Context, url string, html string) (*Article, error) {
	if url == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}
	if html == "" {
		return nil, fmt.Errorf("html cannot be empty")
	}

	// Create input JSON for Postlight Parser
	input := map[string]string{
		"url":  url,
		"html": html,
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Javy uses stdin/stdout for I/O
	stdin := bytes.NewReader(inputJSON)
	var stdout bytes.Buffer

	// Configure module with stdin/stdout
	config := wazero.NewModuleConfig().
		WithStdin(stdin).
		WithStdout(&stdout)

	// Instantiate and run the WASM module
	mod, err := p.runtime.InstantiateModule(ctx, p.module, config)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}
	defer mod.Close(ctx)

	// The module runs on instantiation with Javy
	// Read the output from stdout
	outputJSON := stdout.Bytes()
	if len(outputJSON) == 0 {
		return nil, fmt.Errorf("WASM module produced no output")
	}

	// Check if the output is an error
	var errResult struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(outputJSON, &errResult); err == nil && errResult.Error != "" {
		return nil, fmt.Errorf("parser error: %s", errResult.Error)
	}

	// Parse the result JSON into Article struct
	var article Article
	if err := json.Unmarshal(outputJSON, &article); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w (output: %s)", err, string(outputJSON))
	}

	return &article, nil
}
