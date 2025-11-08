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
# Build the tool
cd lib/linters/uselesswrapper
go build -o uselesswrapper ./cmd/uselesswrapper

# Run on a package
./uselesswrapper ./path/to/package

# Run on multiple packages
./uselesswrapper ./pkg1 ./pkg2 ./pkg3
```

### Running the tests

```bash
cd lib/linters/uselesswrapper
go test -v
```

## Integration

This analyzer can be:
1. Run as a standalone tool (as shown above)
2. Integrated into golangci-lint as a custom linter
3. Integrated into other analysis tools using the `go/analysis` framework

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
