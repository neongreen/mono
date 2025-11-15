# Dagger

This document describes how to use Dagger functions in this repository.

## Linting

The `lint` function runs golangci-lint, uselesswrapper, and cobralint on Go projects.

### Lint all projects

```bash
dagger call lint
```

### Lint specific projects

Lint specific projects using the `--projects` flag:

```bash
dagger call lint --projects tk
dagger call lint --projects tk,conf
```

### How it works

- When no projects are specified, lints everything with `./...`
- When projects are specified, lints only those paths (e.g., `./tk/...`, `./conf/...`)
- `cobralint` only runs when linting all projects or when `tk` is specifically requested (since it only applies to the tk project)
