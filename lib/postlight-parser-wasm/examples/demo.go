package main

import (
	"context"
	"fmt"
	"log"

	"github.com/neongreen/mono/lib/postlight-parser-wasm"
)

func main() {
	fmt.Println("=== lib/parser Demonstration ===\n")

	// Create parser instance
	p, err := parser.New()
	if err != nil {
		log.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Simulate parsing a Wikipedia article
	wikipediaHTML := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>WebAssembly - Wikipedia</title>
</head>
<body>
	<article>
		<h1>WebAssembly</h1>
		<p>WebAssembly (abbreviated Wasm) is a binary instruction format for a stack-based virtual machine. 
		Wasm is designed as a portable compilation target for programming languages, enabling deployment on 
		the web for client and server applications.</p>
		
		<h2>Design</h2>
		<p>WebAssembly describes a memory-safe, sandboxed execution environment that may even be implemented 
		inside existing JavaScript virtual machines. When embedded in the web, WebAssembly will enforce the 
		same-origin and permissions security policies of the browser.</p>
		
		<h2>Features</h2>
		<ul>
			<li>Fast, efficient, and portable</li>
			<li>Readable and debuggable</li>
			<li>Part of the open web platform</li>
		</ul>
		
		<p>WebAssembly is developed by a W3C Community Group that includes representatives from all major 
		browsers. It is a web standard.</p>
	</article>
</body>
</html>`

	ctx := context.Background()
	article, err := p.Extract(ctx, "https://en.wikipedia.org/wiki/WebAssembly", wikipediaHTML)
	if err != nil {
		log.Fatalf("Failed to extract article: %v", err)
	}

	// Display results
	fmt.Println("📄 Extracted Article Information:")
	fmt.Println("─────────────────────────────────")
	fmt.Printf("Title:      %s\n", article.Title)
	fmt.Printf("URL:        %s\n", article.URL)
	fmt.Printf("Domain:     %s\n", article.Domain)
	fmt.Printf("Word Count: %d\n", article.WordCount)
	fmt.Printf("Direction:  %s\n", article.Direction)
	fmt.Println()

	// Show content preview
	fmt.Println("📝 Content Preview (first 200 characters):")
	fmt.Println("─────────────────────────────────")
	contentPreview := article.Content
	if len(contentPreview) > 200 {
		contentPreview = contentPreview[:200] + "..."
	}
	fmt.Println(contentPreview)
	fmt.Println()

	// Test with WASM-enabled parser
	fmt.Println("🚀 Testing WASM-Enabled Parser:")
	fmt.Println("─────────────────────────────────")
	wasmParser, err := parser.NewWithWasm(ctx)
	if err != nil {
		log.Fatalf("Failed to create WASM parser: %v", err)
	}
	defer wasmParser.Close()

	simpleHTML := `<html><head><title>WASM Test</title></head><body><p>Testing wazero runtime integration.</p></body></html>`
	wasmArticle, err := wasmParser.Extract(ctx, "https://example.com/test", simpleHTML)
	if err != nil {
		log.Fatalf("Failed to extract with WASM parser: %v", err)
	}

	fmt.Printf("✅ WASM parser initialized successfully\n")
	fmt.Printf("✅ Parsed article: %s\n", wasmArticle.Title)
	fmt.Println()

	fmt.Println("✅ All demonstrations completed successfully!")
	fmt.Println("\nThe lib/parser library is ready for use!")
}
