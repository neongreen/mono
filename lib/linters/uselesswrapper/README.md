# uselesswrapper

A Go static analysis tool that detects useless function wrappers.

## What it detects

A useless wrapper is a function that:
- Contains only a single return statement
- Returns the result of calling another function
- Passes all parameters unchanged (in the same order)
- Adds no additional logic, validation, or error handling

## Examples

### Detected as useless wrapper ❌

```go
func getCurrentUser() (string, error) {
    return utils.GetCurrentUser()
}

func getUserInfo() (*user.User, error) {
    return user.Current()
}
```

These wrappers add unnecessary indirection and should be replaced with direct calls to the underlying function.

### NOT detected (valid wrappers) ✅

```go
// Adds error handling
func getCurrentUserWithDefault() (string, error) {
    username, err := getUser()
    if err != nil {
        return "unknown", nil
    }
    return username, nil
}

// Transforms the result
func getCurrentUserUpper() string {
    user, _ := getUser()
    return strings.ToUpper(user)
}

// Has multiple statements (e.g., logging)
func getCurrentUserWithLogging() (string, error) {
    fmt.Println("Getting current user")
    return getUser()
}

// Transforms parameters
func swapParams(a, b int) (int, int) {
    return doSwap(b, a)
}

// Methods are skipped (not standalone functions)
func (h *Helper) getCurrentUser() (string, error) {
    return getUser()
}
```

## Usage

### As a standalone tool

```bash
# Build the tool (from repository root)
go build -o uselesswrapper ./lib/linters/uselesswrapper/cmd/uselesswrapper

# Run on a package
./uselesswrapper ./path/to/package

# Run on multiple packages
./uselesswrapper ./pkg1 ./pkg2 ./pkg3

# Run on entire monorepo
./uselesswrapper ./...
```

### Running the tests

```bash
# From repository root
go test ./lib/linters/uselesswrapper/...

# Or from the linter directory
cd lib/linters/uselesswrapper
go test -v
```

## Integration

This analyzer can be:
1. **Run as a standalone tool** (as shown above)
2. **Integrated into golangci-lint** as a custom linter (see below)
3. **Added as a mise task** for easy access
4. **Integrated into CI/CD pipelines** to prevent useless wrappers
5. **Used with other analysis tools** via the `go/analysis` framework

### Integrating with golangci-lint

To integrate this linter with golangci-lint in this repository:

1. **Build the linter plugin:**
   ```bash
   go build -buildmode=plugin -o uselesswrapper.so ./lib/linters/uselesswrapper/cmd/uselesswrapper
   ```

2. **Add to `.golangci.yml`:**
   ```yaml
   linters-settings:
     custom:
       uselesswrapper:
         path: ./uselesswrapper.so
         description: Detects useless function wrappers
         original-url: github.com/neongreen/mono/lib/linters/uselesswrapper
   ```

3. **Enable the linter:**
   ```yaml
   linters:
     enable:
       - uselesswrapper
   ```

Note: Custom analyzers in golangci-lint require building as Go plugins, which has platform limitations. For simpler integration, use as a standalone tool or add a mise task.

### Adding as a mise task

Add to `mise.toml`:
```toml
[tasks."lint:uselesswrapper"]
description = "Detect useless function wrappers"
run = "go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./..."
```

Then run with:
```bash
mise run lint:uselesswrapper
```

## Implementation

The analyzer uses Go's `go/analysis` framework and performs AST (Abstract Syntax Tree) analysis to detect the pattern. It:

1. Examines all function declarations
2. Skips methods (functions with receivers)
3. Checks if the function body contains exactly one statement
4. Verifies that statement is a return statement with a single function call
5. Ensures all function parameters are passed unchanged to the called function

## Why this matters

Useless wrappers:
- Add unnecessary indirection that makes code harder to follow
- Create duplicate function declarations across packages
- Increase maintenance burden (two places to update instead of one)
- Can cause confusion about which function to call
- Provide no abstraction value

Instead of creating wrappers, callers should directly call the underlying function.
