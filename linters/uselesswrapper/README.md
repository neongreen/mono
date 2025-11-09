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

### Via mise task (integrated)

```bash
# Run the linter on the entire repository
mise run lint:uselesswrapper
```

### As a standalone tool

```bash
# From repository root
go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./...

# Or build and run
go build -o uselesswrapper ./lib/linters/uselesswrapper/cmd/uselesswrapper
./uselesswrapper ./path/to/package
```

### Running the tests

```bash
go test ./lib/linters/uselesswrapper/...
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
