# Useless Wrapper Linter - Detection Results

This document shows the results of running the `uselesswrapper` linter on the tk codebase.

## How to Run

```bash
# From repository root, build the linter
go build -o uselesswrapper ./lib/linters/uselesswrapper/cmd/uselesswrapper

# Run on the tk package
./uselesswrapper ./tk/...

# Or run without building (slower but simpler)
go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./tk/...
```

## Results

The linter detected 13 useless wrapper functions in the tk codebase:

```
tk/internal/types/ids.go:46:1: useless wrapper function: directly call ProjectRef() instead
tk/internal/types/ids.go:115:1: useless wrapper function: directly call TaskRef() instead
tk/cmd/project/helpers.go:7:1: useless wrapper function: directly call utils.GetCurrentUser() instead
tk/cmd/relate/helpers.go:7:1: useless wrapper function: directly call utils.GetCurrentUser() instead
tk/internal/remote/export.go:192:1: useless wrapper function: directly call function instead
tk/internal/remote/export.go:197:1: useless wrapper function: directly call SaveJSON() instead
tk/internal/remote/index.go:12:1: useless wrapper function: directly call function instead
tk/internal/remote/index.go:17:1: useless wrapper function: directly call SaveJSON() instead
tk/internal/remote/ingest.go:212:1: useless wrapper function: directly call function instead
tk/internal/remote/ingest.go:217:1: useless wrapper function: directly call SaveJSON() instead
tk/cmd/root.go:14:1: useless wrapper function: directly call rootCmd.Execute() instead
tk/cmd/utils.go:29:1: useless wrapper function: directly call utils.GetCurrentUser() instead
tk/cmd/ls_test.go:94:1: useless wrapper function: directly call slices.Contains() instead
```

## Key Examples

### Example 1: getCurrentUser() wrappers (PR #259)

The original comment from PR #259 was about this pattern:

```go
// tk/cmd/project/helpers.go:7
func getCurrentUser() (string, error) {
    return utils.GetCurrentUser()
}
```

This wrapper appears in 3 different locations:
- `tk/cmd/project/helpers.go`
- `tk/cmd/relate/helpers.go`  
- `tk/cmd/utils.go`

**Fix**: Replace all usages with direct calls to `utils.GetCurrentUser()`.

### Example 2: Type constructor wrappers

```go
// tk/internal/types/ids.go:46
func NewProjectRef(s string) ProjectRef {
    return ProjectRef(s)
}

// tk/internal/types/ids.go:115
func NewTaskRef(s string) TaskRef {
    return TaskRef(s)
}
```

These are just type conversions wrapped in functions. Callers should use `ProjectRef(s)` and `TaskRef(s)` directly.

### Example 3: JSON save wrappers

Multiple files have wrappers around SaveJSON:

```go
// tk/internal/remote/export.go:197
func (s *state) save() error {
    return s.SaveJSON()
}

// tk/internal/remote/index.go:17
func (i *Index) save() error {
    return i.SaveJSON()
}

// tk/internal/remote/ingest.go:217
func (s *state) save() error {
    return s.SaveJSON()
}
```

These add no value - callers should call `.SaveJSON()` directly.

## Benefits of Removing Useless Wrappers

1. **Reduces indirection** - makes code easier to follow
2. **Eliminates duplication** - one function instead of multiple wrappers
3. **Reduces maintenance burden** - fewer places to update
4. **Clarifies intent** - no confusion about which function to call
5. **Improves discoverability** - easier to find the actual implementation

## Integration Options

This linter can be:
1. Run manually as shown above
2. Added to CI/CD pipelines to catch new wrappers
3. Integrated into golangci-lint as a custom linter
4. Run as part of pre-commit hooks
