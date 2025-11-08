package main

import (
	"context"
	"fmt"
	"log"

	"github.com/neongreen/mono/lib/parser"
)

func main() {
	// Create a parser instance
	p, err := parser.New()
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	// Example 1: Parse a simple blog post
	fmt.Println("=== Example 1: Simple Blog Post ===")
	simpleBlog := `<!DOCTYPE html>
<html>
<head>
	<title>Getting Started with Go</title>
</head>
<body>
	<article>
		<h1>Getting Started with Go</h1>
		<p>Go is a statically typed, compiled programming language designed at Google. It is syntactically similar to C, but with memory safety, garbage collection, structural typing, and CSP-style concurrency.</p>
		<h2>Why Go?</h2>
		<p>Go combines the best of both worlds: the efficiency of compiled languages and the ease of use of interpreted languages.</p>
	</article>
</body>
</html>`

	article1, err := p.Extract(context.Background(), "https://blog.example.com/go-intro", simpleBlog)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Title: %s\n", article1.Title)
	fmt.Printf("URL: %s\n", article1.URL)
	fmt.Printf("Domain: %s\n", article1.Domain)
	fmt.Printf("Word Count: %d\n", article1.WordCount)
	fmt.Printf("Content Preview: %.100s...\n\n", article1.Content)

	// Example 2: Parse a news article
	fmt.Println("=== Example 2: News Article ===")
	newsArticle := `<!DOCTYPE html>
<html lang="en">
<head>
	<title>Breaking: New Discovery in Science</title>
	<meta name="author" content="Science Reporter">
</head>
<body>
	<main>
		<article>
			<h1>Breaking: New Discovery in Science</h1>
			<p class="byline">By Science Reporter</p>
			<p class="date">November 8, 2024</p>
			<img src="/images/discovery.jpg" alt="Scientific discovery">
			<p>Scientists have made a groundbreaking discovery that could change our understanding of the universe. The research team, working at a leading university, has published their findings in a prestigious journal.</p>
			<p>This discovery has implications for multiple fields of study, including physics, chemistry, and biology. Experts around the world are calling it one of the most significant findings of the decade.</p>
			<h2>What This Means</h2>
			<p>The implications of this discovery are far-reaching. It opens up new avenues for research and could lead to practical applications in the coming years.</p>
		</article>
	</main>
</body>
</html>`

	article2, err := p.Extract(context.Background(), "https://news.example.com/science-discovery", newsArticle)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Title: %s\n", article2.Title)
	fmt.Printf("URL: %s\n", article2.URL)
	fmt.Printf("Domain: %s\n", article2.Domain)
	fmt.Printf("Word Count: %d\n\n", article2.WordCount)

	// Example 3: Using WASM-enabled parser
	fmt.Println("=== Example 3: WASM-Enabled Parser ===")
	ctx := context.Background()
	wasmParser, err := parser.NewWithWasm(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer wasmParser.Close()

	techArticle := `<html>
<head><title>WebAssembly in Production</title></head>
<body>
	<h1>WebAssembly in Production</h1>
	<p>WebAssembly is transforming how we build web applications. Companies are using it to bring desktop-class performance to the browser.</p>
	<p>From video editing to 3D modeling, WASM enables use cases that were previously impossible in web applications.</p>
</body>
</html>`

	article3, err := wasmParser.Extract(ctx, "https://tech.example.com/wasm", techArticle)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Title: %s\n", article3.Title)
	fmt.Printf("Domain: %s\n", article3.Domain)
	fmt.Printf("Word Count: %d\n\n", article3.WordCount)

	// Example 4: Parse from JSON
	fmt.Println("=== Example 4: Parse from JSON ===")
	jsonData := []byte(`{
		"title": "Understanding JSON Parsing",
		"content": "<p>JSON is a lightweight data interchange format.</p>",
		"url": "https://example.com/json",
		"domain": "example.com",
		"word_count": 8,
		"direction": "ltr"
	}`)

	article4, err := p.ExtractFromJSON(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Title: %s\n", article4.Title)
	fmt.Printf("URL: %s\n", article4.URL)
	fmt.Printf("Content: %s\n\n", article4.Content)

	fmt.Println("=== All Examples Completed Successfully ===")
}
