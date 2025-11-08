package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/neongreen/mono/lib/postlight"
)

func main() {
	// Parse command-line flags
	urlFlag := flag.String("url", "", "URL to parse")
	htmlFlag := flag.String("html", "", "HTML content to parse (if not fetching from URL)")
	jsonFlag := flag.Bool("json", false, "Output as JSON")
	flag.Parse()

	if *urlFlag == "" {
		log.Fatal("Usage: example -url <URL> [-html <HTML>] [-json]")
	}

	ctx := context.Background()

	// Create parser
	parser, err := postlight.NewParser(ctx)
	if err != nil {
		log.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	// Parse the article
	var article *postlight.Article
	if *htmlFlag != "" {
		// Parse provided HTML
		article, err = parser.Parse(ctx, *urlFlag, *htmlFlag)
	} else {
		// Fetch and parse URL
		article, err = parser.ParseURL(ctx, *urlFlag)
	}

	if err != nil {
		log.Fatalf("Failed to parse article: %v", err)
	}

	// Output results
	if *jsonFlag {
		// JSON output
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(article); err != nil {
			log.Fatalf("Failed to encode JSON: %v", err)
		}
	} else {
		// Human-readable output
		fmt.Println("=== Article Content ===")
		fmt.Println()
		fmt.Printf("Title:      %s\n", article.Title)
		fmt.Printf("Author:     %s\n", article.Author)
		fmt.Printf("Published:  %s\n", article.DatePublished)
		fmt.Printf("Domain:     %s\n", article.Domain)
		fmt.Printf("Word Count: %d\n", article.WordCount)
		if article.LeadImageURL != "" {
			fmt.Printf("Lead Image: %s\n", article.LeadImageURL)
		}
		fmt.Println()

		if article.Excerpt != "" {
			fmt.Println("=== Excerpt ===")
			fmt.Println(article.Excerpt)
			fmt.Println()
		}

		fmt.Println("=== Content ===")
		// Truncate content for display
		content := article.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		fmt.Println(content)
	}
}
