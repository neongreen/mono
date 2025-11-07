package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/neongreen/mono/lib/toml"
)

func main() {
	fmt.Println("=== TOML Library Basic Example ===")
	fmt.Println()

	// Example 1: Parse and Read
	fmt.Println("Example 1: Parse and Read")
	fmt.Println(strings.Repeat("-", 40))

	input := `# Server configuration
[server]
host = "localhost"  # The host to bind to
port = 8080         # The port to listen on
debug = false

# Database settings
[database]
url = "postgres://localhost/mydb"
max_connections = 100
`

	doc, err := toml.ParseString(input)
	if err != nil {
		log.Fatal(err)
	}

	// Read values
	host, _ := doc.Get("server.host")
	port, _ := doc.Get("server.port")
	dbURL, _ := doc.Get("database.url")

	fmt.Printf("Server Host: %v\n", host)
	fmt.Printf("Server Port: %v\n", port)
	fmt.Printf("Database URL: %v\n", dbURL)
	fmt.Println()

	// Example 2: Modify Values (Comments Preserved!)
	fmt.Println("Example 2: Modify Values")
	fmt.Println(strings.Repeat("-", 40))

	doc.Set("server.port", 9090)
	doc.Set("server.debug", true)
	doc.Set("server.tls_enabled", true) // New key

	fmt.Println("Modified TOML (notice comments are preserved):")
	fmt.Println(doc.String())
	fmt.Println()

	// Example 3: Working with Arrays
	fmt.Println("Example 3: Working with Arrays")
	fmt.Println(strings.Repeat("-", 40))

	doc.Set("server.allowed_hosts", []string{"localhost", "127.0.0.1", "example.com"})

	allowedHosts, _ := doc.Get("server.allowed_hosts")
	if hosts, ok := allowedHosts.([]any); ok {
		fmt.Println("Allowed hosts:")
		for i, h := range hosts {
			fmt.Printf("  %d. %v\n", i+1, h)
		}
	}
	fmt.Println()

	// Example 4: Working with Inline Tables
	fmt.Println("Example 4: Working with Inline Tables")
	fmt.Println(strings.Repeat("-", 40))

	doc.Set("admin", map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
		"role":  "superuser",
	})

	admin, _ := doc.Get("admin")
	if adminTable, ok := admin.(map[string]any); ok {
		fmt.Println("Admin info:")
		fmt.Printf("  Name: %v\n", adminTable["name"])
		fmt.Printf("  Email: %v\n", adminTable["email"])
		fmt.Printf("  Role: %v\n", adminTable["role"])
	}
	fmt.Println()

	// Example 5: Delete Keys
	fmt.Println("Example 5: Delete Keys")
	fmt.Println(strings.Repeat("-", 40))

	fmt.Println("Before deletion, database.url exists:", doc.Has("database.url"))
	doc.Delete("database.url")
	fmt.Println("After deletion, database.url exists:", doc.Has("database.url"))
	fmt.Println()

	// Example 6: Final TOML Output
	fmt.Println("Example 6: Final TOML Output")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println(doc.String())
}
