# CI Guidelines for Monorepo

This document defines the rules and best practices for GitHub Actions CI workflows in this monorepo.

## Core Principle: Path-Based Triggering

**All CI workflows MUST be limited to changes in their respective directories.**

This is a monorepo containing multiple independent projects. CI pipelines should only run when files in their specific project directory change, to avoid unnecessary CI runs and resource waste.

## Workflow Structure

### Required Path Filtering

Every workflow must include path filters for both `push` and `pull_request` events:

```yaml
on:
  push:
    paths:
      - 'project-name/**'
  pull_request:
    paths:
      - 'project-name/**'
```

### Working Directory

Set the `working-directory` for steps that need to run commands in the project directory:

```yaml
- name: Run tests
  working-directory: project-name
  run: go test -v
```

## Example Workflows

### Go Project

```yaml
name: project-name

on:
  push:
    paths:
      - 'project-name/**'
  pull_request:
    paths:
      - 'project-name/**'

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24.7'
      
      - name: Run tests
        working-directory: project-name
        run: go test ./... -v
```

### Node.js Project

```yaml
name: project-name

on:
  push:
    paths:
      - 'project-name/**'
  pull_request:
    paths:
      - 'project-name/**'

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Install dependencies
        working-directory: project-name
        run: npm ci
      
      - name: Run tests
        working-directory: project-name
        run: npm test
```

## Current Projects

| Project | Language | Workflow File | Status |
|---------|----------|---------------|--------|
| `dissect` | Go | `.github/workflows/dissect.yml` | ✅ Active |
| `markdown-format` | Go | `.github/workflows/markdown-format.yml` | ✅ Active |
| `diagram-dsl` | TypeScript | `.github/workflows/diagram-dsl.yml` | ✅ Active |
| `want` | - | - | 🔜 Pending |

## Release Workflow

All Go projects in the monorepo are automatically released via `.github/workflows/release.yml`.

- **Triggers**: Push to main, pull request activity
- **Auto-detection**: Finds all Go projects with `go.mod` and `main.go`
- **Platforms**: Linux, macOS, Windows (amd64 and arm64)
- **Versioning**: `<project>--<branch>.<number>` (e.g., `dissect--main.1`, `dissect--pr-42.1`)
- **Documentation**: See `.github/workflows/RELEASE_WORKFLOW.md` for details

## Adding a New Project

When adding a new project to the monorepo:

1. **Create workflow file**: `.github/workflows/<project-name>.yml`
2. **Add path filters**: Include `<project-name>/**` in both `push` and `pull_request` events
3. **Set working directory**: Use `working-directory: <project-name>` for all project-specific steps
4. **Update this document**: Add the project to the "Current Projects" table above

## Special Cases

### Shared Dependencies

If multiple projects share common code in a shared directory:

```yaml
on:
  push:
    paths:
      - 'project-name/**'
      - 'shared/**'
  pull_request:
    paths:
      - 'project-name/**'
      - 'shared/**'
```

### Workflow Changes

Changes to the workflow file itself should trigger the workflow:

```yaml
on:
  push:
    paths:
      - 'project-name/**'
      - '.github/workflows/project-name.yml'
  pull_request:
    paths:
      - 'project-name/**'
      - '.github/workflows/project-name.yml'
```

## Benefits

✅ **Efficient CI runs** - Only test what changed  
✅ **Faster feedback** - Reduced queue times  
✅ **Clear separation** - Each project is independent  
✅ **Cost effective** - Minimize CI minutes usage  
✅ **Easier debugging** - Workflow logs are project-specific

## Enforcement

- All new workflows must follow these guidelines
- Existing workflows should be updated to comply
- Pull requests adding workflows without proper path filtering will be rejected

## Related Documentation

- [GitHub Actions: Workflow syntax - paths](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onpushpull_requestpull_request_targetpathspaths-ignore)
- [Monorepo CI best practices](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#running-your-workflow-only-when-a-push-affects-specific-files)
