# Dagger Integration Guide

This document covers the Dagger-specific setup steps for new Go projects in the monorepo.

## Add to dagger.json include list

Add the project directory to the `include` array in `dagger.json`:

```json
{
  "include": [
    "go.mod",
    "go.sum",
    ".golangci.yml",
    ".dagger/**",
    "linters/**",
    "tk/**",
    ...
    "{project-dir}/**"  // Add this line
  ]
}
```

For example: `"my-tool/**"` or `"lib/my-lib/**"`

This ensures Dagger has access to the project files during builds.

## Create Dagger Project File

Create a Dagger project file at `.dagger/project_{sanitized-name}.go`.

**File naming convention:**
- CLI Tool: `project_{projectname}.go` (e.g., `project_mytool.go`)
- Library: `project_lib_{libname}.go` (e.g., `project_lib_mylib.go`)
- Linter: `project_linters_{lintername}.go` (e.g., `project_linters_mylinter.go`)

**Template for simple projects:**

```go
package main

import (
	"context"

	"dagger/internal/dagger"
)

type {CamelCaseName}Project struct {
	Dagger *Dagger // +private
}

func (p *{CamelCaseName}Project) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("{project-dir}", dag.CurrentModule().Source().Directory(".."))
	return buildProject(ctx, "{project-dir}", ".", src)
}

func (p *{CamelCaseName}Project) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("{project-dir}", dag.CurrentModule().Source().Directory(".."))
	return testProject(ctx, "{project-dir}", format, src)
}

func (p *{CamelCaseName}Project) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("{project-dir}", dag.CurrentModule().Source().Directory(".."))
	return coverageFile(ctx, "{project-dir}", format, src)
}
```

Replace `{CamelCaseName}` with the project name in CamelCase (e.g., `MyTool`, `MyLib`).
Replace `{project-dir}` with the actual directory (e.g., `my-tool`, `lib/my-lib`).

**For projects with special test requirements** (like `tk` which needs DuckDB), use `testProjectWithSetup` instead. See @.dagger/project_tk.go for an example.

## Verifying Dagger Integration

After setting up Dagger integration, verify it works:

```bash
# Test Dagger build
dagger call project {project-name} build

# Test Dagger tests (if applicable)
dagger call project {project-name} test
```
