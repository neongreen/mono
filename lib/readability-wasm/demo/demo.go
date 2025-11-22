package main

import (
	"context"
	"fmt"
	"log"

	"github.com/neongreen/mono/lib/readability-wasm"
)

func main() {
	fmt.Println("=== Readability WASM Parser Demo ===")

	ctx := context.Background()

	// Create a parser
	parser, err := readability.NewParser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer parser.Close(ctx)

	// Sample HTML to parse
	html := `<!DOCTYPE html>
<html>
<head>
	<title>The Amazing World of WebAssembly</title>
	<meta name="author" content="Jane Smith">
	<meta name="description" content="Exploring the power of WASM in Go">
</head>
<body>
	<article>
		<h1>The Amazing World of WebAssembly</h1>
		<p class="byline">By Jane Smith</p>
		<p>WebAssembly (WASM) is revolutionizing how we think about running code in browsers and beyond. With the advent of tools like wazero, we can now run WASM modules directly in Go applications without needing CGO.</p>
		<h2>Why WASM Matters</h2>
		<p>WASM provides a sandboxed, portable execution environment that works across platforms. This makes it perfect for embedding complex logic from other languages into Go applications.</p>
		<h2>Real-World Applications</h2>
		<p>From image processing to text parsing, WASM opens up new possibilities for Go developers. This very demo shows how we can leverage JavaScript libraries through WASM!</p>
		<p>The future of cross-language interoperability looks bright with technologies like WASM leading the way.</p>
	</article>
</body>
</html>`

	// Parse the HTML
	article, err := parser.Parse(ctx, "https://example.com/wasm-article", html)
	if err != nil {
		log.Fatalf("Parse failed: %v", err)
	}

	// Display results
	fmt.Println("📄 Article Information:")
	fmt.Println("─────────────────────────")
	fmt.Printf("Title: %s\n", article.Title)
	fmt.Printf("URL: %s\n", article.URL)
	fmt.Printf("Domain: %s\n", article.Domain)
	fmt.Printf("Word Count: %d\n", article.WordCount)
	fmt.Printf("Direction: %s\n", article.Direction)
	fmt.Println()

	if article.Excerpt != "" {
		fmt.Println("📝 Excerpt:")
		fmt.Println("─────────────────────────")
		if len(article.Excerpt) > 200 {
			fmt.Println(article.Excerpt[:200] + "...")
		} else {
			fmt.Println(article.Excerpt)
		}
		fmt.Println()
	}

	fmt.Println("📖 Content (first 300 chars):")
	fmt.Println("─────────────────────────")
	if len(article.Content) > 300 {
		fmt.Println(article.Content[:300] + "...")
	} else {
		fmt.Println(article.Content)
	}
	fmt.Println()

	fmt.Println("✅ Demo completed successfully!")
	fmt.Println()
	fmt.Println("This demonstrates:")
	fmt.Println("  • Go calling WASM via wazero")
	fmt.Println("  • JavaScript execution in a sandboxed environment")
	fmt.Println("  • HTML parsing without CGO")
	fmt.Println("  • Cross-language interoperability")
}
