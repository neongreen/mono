# tomlcp - Comment-Preserving TOML Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/neongreen/mono/tomlcp.svg)](https://pkg.go.dev/github.com/neongreen/mono/tomlcp)

`tomlcp` is a Go library for parsing, modifying, and serializing TOML documents while **preserving all comments, formatting, and declaration order**. It wraps the excellent [creachadair/tomledit](https://github.com/creachadair/tomledit) library to provide a stable, user-friendly API.

## Features

- ✅ **Comment Preservation**: All comments (block, inline, and trailing) are preserved during read-modify-write operations
- ✅ **Format Preservation**: Original formatting and whitespace are maintained
- ✅ **Order Preservation**: Declaration order of keys and sections is preserved
- ✅ **Simple API**: Easy-to-use interface with `Get`, `Set`, `Delete`, and `Has` methods
- ✅ **Dotted Path Support**: Access nested values using dotted paths (e.g., `server.database.host`)
- ✅ **Full TOML v1.0.0 Support**: Supports all TOML features including arrays, inline tables, and multiline strings
- ✅ **Extensively Tested**: Comprehensive test suite with 27+ test cases

## Installation

```bash
go get github.com/neongreen/mono/tomlcp
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/neongreen/mono/tomlcp"
)

func main() {
	// Parse a TOML document
	input := `# Server configuration
[server]
host = "localhost"  # The host to bind to
port = 8080         # The port to listen on

[database]
url = "postgres://localhost/db"
`

	doc, err := tomlcp.ParseString(input)
	if err != nil {
		log.Fatal(err)
	}

	// Read values
	host, _ := doc.Get("server.host")
	fmt.Println("Host:", host) // Output: Host: localhost

	// Modify values (comments are preserved!)
	doc.Set("server.port", 9090)
	doc.Set("server.debug", true)

	// Delete values
	doc.Delete("database.url")

	// Check if a value exists
	if doc.Has("server.host") {
		fmt.Println("Host is configured")
	}

	// Write back to TOML (with all comments preserved)
	fmt.Println(doc.String())
}
```

## API Documentation

### Parsing

#### `Parse(input []byte) (*Document, error)`

Parses a TOML document from a byte slice.

```go
data, _ := os.ReadFile("config.toml")
doc, err := tomlcp.Parse(data)
```

#### `ParseString(input string) (*Document, error)`

Parses a TOML document from a string.

```go
doc, err := tomlcp.ParseString(`
name = "myapp"
version = 1
`)
```

### Reading Values

#### `Get(path string) (interface{}, error)`

Retrieves a value at the given dotted path. Returns `nil` if the path doesn't exist.

```go
value, err := doc.Get("server.port")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Port: %v\n", value)
```

Supported return types:
- `string` - for string values
- `int64` - for integer values
- `float64` - for floating-point values
- `bool` - for boolean values
- `[]interface{}` - for arrays
- `map[string]interface{}` - for inline tables

#### `Has(path string) bool`

Returns true if the given path exists in the document.

```go
if doc.Has("server.port") {
    fmt.Println("Port is configured")
}
```

### Modifying Values

#### `Set(path string, value interface{}) error`

Sets a value at the given dotted path, creating intermediate sections if necessary. If the key already exists, its comments are preserved.

Supported value types:
- `string`, `int`, `int64`, `float64`, `bool`
- `[]interface{}`, `[]string`, `[]int`
- `map[string]interface{}`

```go
// Set a simple value
doc.Set("server.port", 8080)

// Set a nested value (creates [server] section if needed)
doc.Set("server.tls.enabled", true)

// Set an array
doc.Set("server.hosts", []string{"localhost", "example.com"})

// Set an inline table
doc.Set("person", map[string]interface{}{
    "name": "Alice",
    "age":  30,
})
```

#### `Delete(path string) error`

Removes a key at the given dotted path.

```go
doc.Delete("server.debug")
```

### Serialization

#### `String() string`

Serializes the document back to TOML format as a string, preserving all comments and formatting.

```go
output := doc.String()
fmt.Println(output)
```

#### `Bytes() []byte`

Serializes the document back to TOML format as a byte slice.

```go
data := doc.Bytes()
os.WriteFile("config.toml", data, 0644)
```

## Examples

### Example 1: Configuration File Management

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/neongreen/mono/tomlcp"
)

func main() {
	// Read existing config
	data, err := os.ReadFile("app.toml")
	if err != nil {
		log.Fatal(err)
	}

	doc, err := tomlcp.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	// Update configuration
	doc.Set("app.version", "2.0.0")
	doc.Set("app.debug", false)
	doc.Set("server.max_connections", 1000)

	// Save back to file (comments preserved!)
	err = os.WriteFile("app.toml", doc.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Configuration updated successfully")
}
```

### Example 2: Comment Preservation

```go
package main

import (
	"fmt"

	"github.com/neongreen/mono/tomlcp"
)

func main() {
	input := `# Application settings
app_name = "myapp"    # The name of the application
version = 1           # Current version

# Database configuration
[database]
# Connection settings
host = "localhost"
port = 5432
`

	doc, _ := tomlcp.ParseString(input)

	// Modify the version (comment is preserved!)
	doc.Set("version", 2)

	// Modify database port (comments are preserved!)
	doc.Set("database.port", 5433)

	fmt.Println(doc.String())
	// Output still contains:
	// # Application settings
	// app_name = "myapp"    # The name of the application
	// version = 2           # Current version
	// ... etc
}
```

### Example 3: Working with Complex Structures

```go
package main

import (
	"fmt"

	"github.com/neongreen/mono/tomlcp"
)

func main() {
	doc, _ := tomlcp.ParseString("")

	// Create a complex nested structure
	doc.Set("servers.alpha.ip", "10.0.0.1")
	doc.Set("servers.alpha.dc", "eqdc10")
	doc.Set("servers.beta.ip", "10.0.0.2")
	doc.Set("servers.beta.dc", "eqdc10")

	// Set arrays
	doc.Set("database.ports", []int{8001, 8002, 8003})

	// Set inline tables
	doc.Set("owner", map[string]interface{}{
		"name":  "Tom",
		"email": "tom@example.com",
	})

	fmt.Println(doc.String())
}
```

### Example 4: Migrating Configuration Values

```go
package main

import (
	"fmt"

	"github.com/neongreen/mono/tomlcp"
)

func main() {
	input := `
[old_section]
setting = "value"

[server]
host = "localhost"
`

	doc, _ := tomlcp.ParseString(input)

	// Migrate old setting to new location
	if oldValue, _ := doc.Get("old_section.setting"); oldValue != nil {
		doc.Set("server.setting", oldValue)
		doc.Delete("old_section.setting")
	}

	fmt.Println(doc.String())
}
```

## Advanced Usage

### Accessing Arrays

```go
doc, _ := tomlcp.ParseString(`
tags = ["go", "toml", "parser"]
`)

tags, _ := doc.Get("tags")
if arr, ok := tags.([]interface{}); ok {
	for i, tag := range arr {
		fmt.Printf("Tag %d: %v\n", i, tag)
	}
}
```

### Accessing Inline Tables

```go
doc, _ := tomlcp.ParseString(`
person = { name = "Alice", age = 30 }
`)

person, _ := doc.Get("person")
if table, ok := person.(map[string]interface{}); ok {
	fmt.Println("Name:", table["name"])
	fmt.Println("Age:", table["age"])
}
```

### Type Assertions

When retrieving values, you'll need to assert the type:

```go
// String
if str, ok := value.(string); ok {
	fmt.Println(str)
}

// Integer (always returned as int64)
if num, ok := value.(int64); ok {
	fmt.Println(num)
}

// Float
if f, ok := value.(float64); ok {
	fmt.Println(f)
}

// Boolean
if b, ok := value.(bool); ok {
	fmt.Println(b)
}
```

## Testing

The library includes a comprehensive test suite with 27+ test cases covering:

- Parsing various TOML formats
- Getting and setting values
- Deleting values
- Comment preservation
- Round-trip parsing and serialization
- Arrays and inline tables
- Edge cases and error handling
- Unicode support
- Large and complex documents

Run tests with:

```bash
go test -v ./...
```

## Comparison with Other Libraries

| Feature | tomlcp | go-toml/v2 | BurntSushi/toml |
|---------|--------|------------|-----------------|
| Comment preservation | ✅ | ❌ | ❌ |
| Format preservation | ✅ | ❌ | ❌ |
| Order preservation | ✅ | ❌ | ❌ |
| Marshal/Unmarshal | ❌ | ✅ | ✅ |
| Struct tags | ❌ | ✅ | ✅ |
| Use case | Config editing | Data serialization | Data serialization |

**When to use tomlcp:**
- You need to edit TOML files while preserving comments
- You're building a configuration management tool
- You need to maintain human-readable formatting
- You want to programmatically update config files without losing documentation

**When to use go-toml/v2 or BurntSushi/toml:**
- You just need to deserialize TOML into Go structs
- Comments and formatting don't matter
- You need the fastest parsing performance

## Implementation Details

`tomlcp` is built on top of [creachadair/tomledit](https://github.com/creachadair/tomledit), which provides the low-level AST-based parsing and formatting capabilities. `tomlcp` adds:

- A higher-level, more ergonomic API
- Automatic comment preservation when updating values
- Simplified path-based access to nested values
- Type conversion helpers

## Limitations

- The library works at the AST level, not the semantic level. It preserves the syntactic structure but doesn't validate semantic constraints (like duplicate keys).
- When creating new values programmatically, they're added without comments unless you modify the underlying AST.
- Marshal/unmarshal to Go structs is not supported (use go-toml/v2 or BurntSushi/toml for that).

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Acknowledgments

- [creachadair/tomledit](https://github.com/creachadair/tomledit) - The underlying TOML parser and formatter
- [TOML](https://toml.io) - The TOML specification
