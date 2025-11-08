# Integration Guide for uselesswrapper Linter

This guide shows how to integrate the `uselesswrapper` linter into the monorepo's existing linting infrastructure.

## Monorepo Structure

The uselesswrapper linter is part of the monorepo and uses the root `go.mod`:
- **Location**: `lib/linters/uselesswrapper/`
- **Module**: `github.com/neongreen/mono` (same as repo root)
- **Dependencies**: Uses `golang.org/x/tools` from root `go.mod`

No separate go.mod is needed - it's a standard Go package within the monorepo.

## Integration Options

### Option 1: Add as a mise task (Recommended)

Add to the root `mise.toml` file:

```toml
[tasks."lint:uselesswrapper"]
description = "Detect useless function wrappers"
run = "go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./..."

[tasks."lint:uselesswrapper:tk"]
description = "Detect useless function wrappers in tk package"
run = "go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./tk/..."
```

Usage:
```bash
mise run lint:uselesswrapper        # Check entire repo
mise run lint:uselesswrapper:tk     # Check just tk package
```

### Option 2: Run directly with go run

```bash
# Check entire repo
go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./...

# Check specific package
go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./tk/...

# Check multiple packages
go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./tk/... ./lib/... ./want/...
```

### Option 3: Build once and reuse

```bash
# Build the binary
go build -o bin/uselesswrapper ./lib/linters/uselesswrapper/cmd/uselesswrapper

# Add to .gitignore if needed
echo "bin/uselesswrapper" >> .gitignore

# Use it
./bin/uselesswrapper ./...
```

### Option 4: Integrate with golangci-lint (Advanced)

While golangci-lint supports custom linters, it requires building as a Go plugin which has platform limitations. For this monorepo, Options 1-3 are simpler and more portable.

If you still want to try:

1. **Note**: This requires CGO and has platform-specific limitations
2. Build as plugin: `go build -buildmode=plugin -o uselesswrapper.so ./lib/linters/uselesswrapper/cmd/uselesswrapper`
3. Add to `.golangci.yml`:
   ```yaml
   linters-settings:
     custom:
       uselesswrapper:
         path: ./uselesswrapper.so
         description: Detects useless function wrappers
   ```

However, the mise task approach (Option 1) is recommended for this repo.

## CI/CD Integration

### GitHub Actions

Add to your workflow file (e.g., `.github/workflows/lint.yml`):

```yaml
- name: Run uselesswrapper linter
  run: go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./...
```

Or if using mise:

```yaml
- name: Setup mise
  uses: jdx/mise-action@v2

- name: Run uselesswrapper linter
  run: mise run lint:uselesswrapper
```

### Pre-commit Hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
echo "Running uselesswrapper linter..."
go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./...
if [ $? -ne 0 ]; then
    echo "Useless wrappers detected! Please fix before committing."
    exit 1
fi
```

## Current Findings

As of now, the linter detects 13 useless wrappers in the tk package:
- 3x `getCurrentUser()` duplicates
- 6x `save()` wrappers around `SaveJSON()`
- 2x type constructors (`NewProjectRef`, `NewTaskRef`)
- 2x misc wrappers

To see the full list:
```bash
go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./tk/...
```

## Recommended Next Steps

1. **Add as mise task** (Option 1) for easy access
2. **Run on specific packages** when refactoring
3. **Consider adding to CI** to prevent new useless wrappers
4. **Gradually fix existing issues** as you work on related code

## Notes

- The linter has exit code 3 when issues are found (standard for analysis tools)
- It can be safely run on any package without modifying code
- False positives are rare but possible - use judgment when fixing
- The linter skips methods (functions with receivers) as they may be part of an interface
