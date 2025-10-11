# Release Workflow Architecture

## Overview

This document describes the technical architecture of the automated release system for Go projects.

## Workflow Flow

```
Trigger (Push to main OR Pull Request)
    |
    v
[Detect Job]
    |
    +-- Find directories with go.mod
    |
    +-- Check for main.go files
    |
    +-- Output: JSON array of projects
    |
    v
Decision: Has projects?
    |
    +-- NO --> Workflow ends
    |
    +-- YES --> Continue
    |
    v
[Release Job] (Matrix: one per project)
    |
    +-- Checkout code (fetch-depth: 0 for tags)
    |
    +-- Setup Go 1.24.7
    |
    v
[Determine Version]
    |
    +-- Extract branch/PR info
    |   |
    |   +-- Main: REF="main"
    |   +-- PR: REF="pr-NUMBER"
    |
    +-- Find existing tags: PROJECT/REF.*
    |
    +-- Calculate next version number
    |
    +-- Output: TAG="project/branch.N"
    |
    v
[Build Binaries]
    |
    +-- Detect build target (./cmd or .)
    |
    +-- Build for 5 platforms:
    |   +-- linux/amd64
    |   +-- linux/arm64
    |   +-- darwin/amd64
    |   +-- darwin/arm64
    |   +-- windows/amd64
    |
    +-- Output: dist/PROJECT-VERSION-OS-ARCH[.exe]
    |
    v
[Create Release]
    |
    +-- Create GitHub release
    |
    +-- Set tag: PROJECT/BRANCH.NUMBER
    |
    +-- Attach all binaries
    |
    +-- Mark as pre-release if from PR
    |
    v
Done
```

## Version Numbering

### Format
`<project>/<branch>.<number>`

### Examples

**Main Branch:**
- First push: `dissect/main.1`
- Second push: `dissect/main.2`
- Third push: `dissect/main.3`

**Pull Request #42:**
- First build: `dissect/pr-42.1`
- Updated PR: `dissect/pr-42.2`
- Another update: `dissect/pr-42.3`

**Different Projects:**
- `dissect/main.1` - dissect from main
- `markdown-format/main.1` - markdown-format from main
- Both can have `.1` as they're independent

### Version Determination Algorithm

```bash
# 1. Construct tag prefix
TAG_PREFIX="${PROJECT}/${REF_SAFE}."

# 2. Find all matching tags
EXISTING_TAGS=$(git tag -l "${TAG_PREFIX}*" | sort -V)

# 3. Calculate next number
if [ -z "$EXISTING_TAGS" ]; then
  VERSION_NUM=1
else
  LAST_TAG=$(echo "$EXISTING_TAGS" | tail -1)
  LAST_NUM=$(echo "$LAST_TAG" | sed "s|${TAG_PREFIX}||")
  VERSION_NUM=$((LAST_NUM + 1))
fi
```

## Build Process

### Build Target Detection

```bash
# Find main.go files
MAIN_FILES=$(find "$PROJECT" -name "main.go")

# Determine build path
if echo "$MAIN_FILES" | grep -q "cmd/main.go"; then
  BUILD_TARGET="./cmd"
else
  BUILD_TARGET="."
fi
```

### Cross-Compilation

Each project is built for 5 platform combinations:

```bash
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

for platform in "${PLATFORMS[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "$platform"
  OUTPUT_NAME="${PROJECT}-${VERSION}-${GOOS}-${GOARCH}"
  
  (cd "$PROJECT" && \
   GOOS=$GOOS GOARCH=$GOARCH \
   go build -o "$DIST_DIR/$OUTPUT_NAME" $BUILD_TARGET)
done
```

## Release Creation

### GitHub Release

Uses `softprops/action-gh-release@v2` to:
1. Create a new release
2. Tag it with the calculated version
3. Attach all built binaries
4. Generate installation instructions
5. Mark PR releases as pre-release

### Release Body Template

```markdown
Release of **{project}** version `{version}`

Built from: {main branch | PR #NUMBER}
Commit: {sha}

## Installation

Download the appropriate binary for your platform...

## Available Binaries

- Linux: amd64, arm64
- macOS: amd64 (Intel), arm64 (Apple Silicon)
- Windows: amd64
```

## Trigger Conditions

### Path Filters

The workflow only runs when Go-related files change:

```yaml
paths:
  - '**/*.go'
  - '**/go.mod'
  - '**/go.sum'
  - '.github/workflows/release.yml'
```

### Event Types

**Push to main:**
```yaml
push:
  branches:
    - main
```

**Pull Requests:**
```yaml
pull_request:
  types: [opened, synchronize, reopened]
```

## Project Detection

### Criteria

A directory is considered a Go project if:
1. It has a `go.mod` file
2. It contains at least one `main.go` file (anywhere in the tree)

### Detection Code

```bash
projects=()
for dir in */; do
  dir=${dir%/}
  if [ -f "$dir/go.mod" ]; then
    if find "$dir" -name "main.go" | grep -q .; then
      projects+=("$dir")
    fi
  fi
done
```

### Output Format

Projects are output as a JSON array for use in matrix builds:

```json
["dissect", "markdown-format"]
```

## Error Handling

### Build Failures

If a build fails for any platform:
```bash
if [ $? -ne 0 ]; then
  echo "Failed to build for $GOOS/$GOARCH"
  exit 1
fi
```

The entire job fails, preventing partial releases.

### No Projects Found

If no Go projects are detected:
- Set `has_projects=false`
- Skip the release job entirely
- Workflow completes successfully (no error)

## Security Considerations

### Token Permissions

Uses `GITHUB_TOKEN` provided by GitHub Actions:
- Has write access to create releases
- Scoped to the repository
- Automatically revoked after workflow completes

### Dependency Management

- No external dependencies downloaded at workflow runtime
- Uses official GitHub Actions (`actions/checkout`, `actions/setup-go`)
- Uses community action from softprops for release creation

## Performance Optimizations

### Matrix Strategy

Projects are built in parallel using matrix strategy:
```yaml
strategy:
  matrix:
    project: ${{ fromJson(needs.detect.outputs.projects) }}
```

For 2 projects, both are built simultaneously.

### Shallow Clones

The detect job uses shallow clone (default), but the release job needs full history:
```yaml
fetch-depth: 0  # Get all tags for version calculation
```

## Future Enhancements

Potential improvements:
- Cache Go modules between builds
- Generate checksums (SHA256) for binaries
- Sign binaries for additional security
- Support custom build flags per project
- Automated changelog generation from commits
- Notification on release creation
