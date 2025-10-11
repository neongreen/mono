# Release Workflow Implementation Summary

This document summarizes the automated release system implemented for Go projects in this monorepo.

## Problem Solved

You wanted to be able to:
1. Install tools from this repository, not just from main but also from PRs
2. Have versioned releases for each Go project
3. Distinguish between releases from main and different branches/PRs
4. Easily install specific versions

## Solution Delivered

### Automated Release Workflow

**File:** `.github/workflows/release.yml`

- **Triggers:** Push to main, Pull request activity (open, sync, reopen)
- **Path filters:** Only runs when Go files change (`*.go`, `go.mod`, `go.sum`)
- **Auto-detection:** Finds Go projects automatically (directories with `go.mod` + `main.go`)
- **Multi-platform:** Builds for Linux, macOS, and Windows (amd64 and arm64)
- **Version management:** Auto-increments versions per project and branch

### Version Naming Convention

Format: `<project>/<branch>.<number>`

**Examples:**
- `dissect/main.1`, `dissect/main.2`, `dissect/main.3` - Main branch releases
- `dissect/pr-42.1`, `dissect/pr-42.2` - PR #42 releases
- `markdown-format/main.1` - Markdown-format from main

Each project and branch combination has independent version numbering. PR releases are marked as pre-releases.

### Supported Platforms

Every release includes binaries for:
- **Linux:** amd64, arm64
- **macOS:** amd64 (Intel), arm64 (Apple Silicon)
- **Windows:** amd64

Total: 5 binaries per project per release.

### Installation Options

#### Option 1: Quick Install (Recommended)

```bash
# Install latest from main
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s dissect

# Install specific version
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s dissect main.5

# Install from PR (for testing)
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s dissect pr-42.1
```

#### Option 2: Install Script

```bash
# Download and run locally
wget https://raw.githubusercontent.com/neongreen/mono/main/install.sh
chmod +x install.sh

# Install latest from main
./install.sh markdown-format

# Install specific version
./install.sh dissect main.3
```

#### Option 3: Manual Download

1. Go to [Releases](https://github.com/neongreen/mono/releases)
2. Find your desired version (e.g., `dissect/main.1`)
3. Download the binary for your platform
4. Make executable and move to PATH:
   ```bash
   chmod +x dissect-main.1-linux-amd64
   sudo mv dissect-main.1-linux-amd64 /usr/local/bin/dissect
   ```

### Testing PR Changes

When you want to test changes from a PR before merging:

1. Open or update a PR
2. Workflow creates a release (e.g., `dissect/pr-123.1`)
3. Install it:
   ```bash
   ./install.sh dissect pr-123.1
   ```
4. Test the changes
5. If satisfied, merge the PR
6. A new main release is automatically created

## Files Created

### Core Files
- `.github/workflows/release.yml` - Main workflow (7.1 KB)
- `install.sh` - Installation script (4.0 KB)
- `.gitignore` - Build artifact exclusions

### Documentation
- `README.md` - Quick start guide (2.3 KB)
- `.github/workflows/RELEASE_WORKFLOW.md` - User guide (4.7 KB)
- `.github/workflows/RELEASE_ARCHITECTURE.md` - Technical docs (5.9 KB)
- `.github/CI_GUIDELINES.md` - Updated with release info

### Statistics
- Total files: 7
- Total documentation: ~17 KB
- Lines of workflow code: ~200
- Lines of documentation: ~400

## How It Works

### 1. Trigger
Workflow triggers when Go files change in a push to main or PR activity.

### 2. Detect
Scans the repository for Go projects (directories with `go.mod` and `main.go`).

### 3. Version
Calculates the next version number by:
- Finding existing tags for `<project>/<branch>.*`
- Extracting the highest number
- Incrementing by 1

### 4. Build
For each detected project:
- Determines build target (`./cmd` or `.`)
- Builds for 5 platforms in parallel
- Names binaries: `<project>-<version>-<os>-<arch>[.exe]`

### 5. Release
- Creates GitHub release with the version tag
- Attaches all binaries
- Includes installation instructions
- Marks PR releases as pre-releases

## Testing Performed

✅ YAML syntax validation  
✅ Project detection (correctly finds dissect and markdown-format)  
✅ Version calculation and auto-increment  
✅ Build testing for all platforms:
  - dissect: 5 platforms ✓
  - markdown-format: 5 platforms ✓
✅ Install script functionality  
⏳ End-to-end workflow (will run on first PR/push)

## Example Usage Scenarios

### Scenario 1: Installing the Latest Tool
```bash
./install.sh dissect
# Automatically finds and installs the latest main version
```

### Scenario 2: Testing a PR
```bash
# PR #42 has a new feature you want to test
./install.sh dissect pr-42.1
# Run tests with the PR version
dissect --help
# If it works, merge PR #42
# Then install from main:
./install.sh dissect
```

### Scenario 3: Pinning a Specific Version
```bash
# Install a known-good version
./install.sh markdown-format main.3
# System stays on this version until you update manually
```

### Scenario 4: Multiple Projects
```bash
# Install all Go tools
./install.sh dissect
./install.sh markdown-format
# Both installed from their latest main versions
```

## What Happens on Merge

When this PR is merged to main:

1. Workflow triggers (Go files changed)
2. Detects 2 projects: `dissect`, `markdown-format`
3. Calculates versions: Both will be `.1` (first releases)
4. Builds 10 binaries (5 platforms × 2 projects)
5. Creates 2 releases:
   - `dissect/main.1` with 5 binaries
   - `markdown-format/main.1` with 5 binaries

Then users can immediately install with:
```bash
./install.sh dissect
./install.sh markdown-format
```

## Future Releases

Every subsequent push to main will:
- Increment version numbers (`main.2`, `main.3`, etc.)
- Create new releases
- Build fresh binaries

Every PR will:
- Create pre-releases with PR number (`pr-N.1`, `pr-N.2`, etc.)
- Allow testing before merge
- Be marked as pre-release (not "latest")

## Maintenance

### Adding New Go Projects

No changes needed! The workflow automatically:
1. Detects any directory with `go.mod` and `main.go`
2. Builds it for all platforms
3. Creates releases

### Removing Projects

Simply delete the project directory. The workflow will stop building it.

### Updating Platforms

Edit `.github/workflows/release.yml` and modify the `PLATFORMS` array.

## Design Decisions

### Why separate releases per project?

Each project is independent and can have its own version cadence. Users only install what they need.

### Why tag format `project/branch.number`?

- Clearly identifies the project
- Clearly identifies the branch/PR
- Simple incrementing number
- Git-friendly (slashes allowed in tags)

### Why pre-release for PRs?

- Distinguishes experimental builds from stable ones
- Doesn't appear as "latest release"
- Still fully accessible for testing

### Why auto-detect projects?

- No manual configuration needed
- Works for future projects automatically
- Less maintenance overhead

## Comparison with Requirements

| Requirement | Implementation | Status |
|------------|----------------|--------|
| Install tools from repository | `install.sh` script + releases | ✅ |
| Install from main branch | `project/main.N` releases | ✅ |
| Install from PRs | `project/pr-N.N` releases | ✅ |
| Version identification | `project/branch.number` format | ✅ |
| Incremental versions | Auto-increment per project+branch | ✅ |
| Sanitized branch names | PR numbers instead of branch names | ✅ |
| Multiple projects | Matrix builds, independent releases | ✅ |
| Easy installation | Three installation options | ✅ |

## Conclusion

You now have a fully automated release system that:
- ✅ Creates releases for every Go project
- ✅ Works for both main and PRs
- ✅ Uses clear, incrementing version numbers
- ✅ Supports multiple platforms
- ✅ Provides easy installation
- ✅ Requires zero manual intervention

You can install any version of any tool from any branch or PR with a single command.
