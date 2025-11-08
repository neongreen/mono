# Quick Start: Integrating uselesswrapper with mise

This is a quick reference for the most common way to use the uselesswrapper linter in this monorepo.

## Recommended: Add as mise task

Add these lines to the root `mise.toml` file:

```toml
[tasks."lint:uselesswrapper"]
description = "Detect useless function wrappers in entire repo"
run = "go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./..."

[tasks."lint:uselesswrapper:tk"]
description = "Detect useless function wrappers in tk package"
run = "go run ./lib/linters/uselesswrapper/cmd/uselesswrapper ./tk/..."
```

## Usage

```bash
# Check entire repository
mise run lint:uselesswrapper

# Check just the tk package
mise run lint:uselesswrapper:tk

# List all lint tasks
mise tasks | grep lint
```

## Example Output

```bash
$ mise run lint:uselesswrapper:tk
/home/user/mono/tk/internal/types/ids.go:46:1: useless wrapper function: directly call ProjectRef() instead
/home/user/mono/tk/internal/types/ids.go:115:1: useless wrapper function: directly call TaskRef() instead
/home/user/mono/tk/cmd/project/helpers.go:7:1: useless wrapper function: directly call utils.GetCurrentUser() instead
...
```

## Why use go run instead of building?

Using `go run` in the mise task means:
- No need to track a binary in git
- Always uses the latest version of the linter
- Simpler setup - just add to mise.toml and it works
- Slightly slower first run, but cached thereafter

If you prefer a pre-built binary, you can change the task to:
```toml
[tasks."lint:uselesswrapper:build"]
description = "Build uselesswrapper linter"
run = "go build -o bin/uselesswrapper ./lib/linters/uselesswrapper/cmd/uselesswrapper"

[tasks."lint:uselesswrapper"]
description = "Detect useless function wrappers"
depends = ["lint:uselesswrapper:build"]
run = "./bin/uselesswrapper ./..."
```

## Full documentation

See [INTEGRATION.md](./INTEGRATION.md) for complete integration options including:
- CI/CD integration
- Pre-commit hooks
- golangci-lint integration
- Running without mise
