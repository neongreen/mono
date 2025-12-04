# lion

Documentation extraction tool for Go and TypeScript code that generates markdown files from special comments.

## Overview

lion scans source code for special documentation comments and generates organized markdown documentation. This allows you to build a book-like documentation structure by adding comments throughout your codebase that contribute to different chapters/topics.

Supported languages:
- **Go**: Uses `//lion:topic-name` comments
- **TypeScript/TSX**: Uses `@lion topic-name` tags in JSDoc comments

## Installation

```bash
go install github.com/neongreen/mono/lion@latest
```

For TypeScript support, you also need to build the TypeScript helper:

```bash
cd lion/ts-helper
npm install
npm run build
```

## Usage

### Go Documentation Comments

Add `//lion:topic-name` comments to your Go code:

```go
//lion:architecture This is the main entry point for the application.
//lion:architecture It handles initialization and routing.
func main() {
    // ...
}

//lion:api The Config struct holds application configuration.
//lion:api Fields can be loaded from environment variables or config files.
type Config struct {
    Port int
    Debug bool
}
```

### TypeScript Documentation Comments

Add `@lion topic-name` tags in JSDoc comments to your TypeScript code:

```typescript
/**
 * @lion api
 * The User interface defines user data.
 * It contains the basic user information.
 */
interface User {
  name: string;
  email: string;
}

/**
 * @lion api title="API Reference" section="User Functions"
 * Creates a new user with the given name.
 * @param name - The user's name
 * @returns A new User object
 */
function createUser(name: string): User {
  return { name, email: "" };
}
```

### Generate Documentation

```bash
# Generate docs from current directory to ./docs
lion generate

# Specify input directory and output directory
lion generate ./myproject --output ./documentation

# List all topics found in code
lion topics
```

### Output

The tool generates:
- One markdown file per topic (e.g., `architecture.md`, `api.md`)
- An `index.md` file listing all topics
- Each entry includes source file reference for traceability

## Comment Formats

### Go Comment Formats

lion supports three comment formats for Go:

#### 1. Single-line format (marker first)

```
//lion:topic-name Optional content describing this code element
```

- **topic-name**: Hyphen-separated identifier (becomes filename)
- **Content**: Optional markdown-formatted text
- Multiple consecutive `//lion:topic` lines are combined into one entry

#### 2. Block comment format

```go
/*lion:topic-name
Multi-line content can go here
without repeating the topic name.
This makes documentation cleaner.
*/
```

- **topic-name**: Specified once at the beginning
- **Content**: All subsequent lines in the block comment
- Cleaner for longer documentation blocks

All formats can be attached to functions, types, constants, variables, and package declarations.

### TypeScript Comment Format

TypeScript uses JSDoc-style comments with `@lion` tags:

```typescript
/**
 * @lion topic-name
 * Documentation content here.
 * Multiple lines are supported.
 */
```

Optional metadata can be added:

```typescript
/**
 * @lion topic-name title="Custom Title" section="Section Name"
 * Documentation content.
 */
```

- **title**: Overrides the topic's display title in the generated markdown
- **section**: Overrides the section heading for this entry

Supported TypeScript declarations:
- Functions (regular, async, arrow)
- Classes (including abstract)
- Interfaces
- Type aliases
- Enums
- Variables (const, let, var)
- Class methods and properties

## Examples

### Go Example

Marker-first format (recommended):

```go
//lion:getting-started
//
// lion is a documentation extraction tool.
// Add lion comments to generate markdown docs.
// This is the cleanest syntax for multi-line docs.
package main

//lion:architecture
//
// The main function initializes the app.
// It handles setup and configuration.
func main() {
    // ...
}
```

Single-line format:

```go
//lion:getting-started lion is a documentation extraction tool.
//lion:getting-started Add lion comments to generate markdown docs.
package main

//lion:architecture The main function initializes the app.
func main() {
    // ...
}
```

Block comment format:

```go
/*lion:getting-started
lion is a documentation extraction tool.
Add lion comments to generate markdown docs.
No need to repeat the topic on each line.
*/
package main

/*lion:architecture
The main function initializes the app.
It handles setup and configuration.
*/
func main() {
    // ...
}
```

Running `lion generate` creates:
- `docs/getting-started.md`
- `docs/architecture.md`
- `docs/index.md`

### TypeScript Example

```typescript
/**
 * @lion getting-started
 * lion is a documentation extraction tool.
 * Add lion comments to generate markdown docs.
 */

/**
 * @lion architecture
 * The UserService class handles user operations.
 * It provides methods for CRUD operations.
 */
class UserService {
  /**
   * @lion api section="Create User"
   * Creates a new user with the given data.
   */
  async createUser(data: UserData): Promise<User> {
    // ...
  }
}

/**
 * @lion types
 * The User interface defines the user structure.
 */
interface User {
  id: string;
  name: string;
}
```

## Status

**Alpha** - Core functionality works. Comment format and CLI may change.

## Self-Documentation

lion uses itself to generate its own documentation! The `docs/` directory contains markdown files extracted from lion comments in the source code.

To regenerate the documentation:
```bash
./regenerate-docs.sh
```

Or manually:
```bash
go run . generate . --output ./docs
```
