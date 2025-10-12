# prrun Troubleshooting Guide

## Common Issues

### 404 Error When Downloading Binary

**Symptoms:**
```
Looking for PR #27 in neongreen/mono (project: claude-trace)
Found release: claude-trace--pr-27.1
Downloading binary from https://github.com/neongreen/mono/releases/download/claude-trace--pr-27.1/claude-trace-pr-27.1-linux-amd64...
Error: download failed with status 404
```

**Possible Causes:**

#### 1. Release Has No Assets (Build Failed)

The release exists but the GitHub Actions workflow failed to build or upload the binaries.

**How to check:**
- Visit the release page on GitHub: `https://github.com/owner/repo/releases/tag/project--pr-N.1`
- Look at the "Assets" section - if it's empty, the build failed

**Solution:**
- Check the GitHub Actions workflow logs for build failures
- Fix any build errors in the project
- Re-run the workflow or push a new commit to trigger a rebuild

#### 2. Asset Name Mismatch

The asset exists but with a different name than expected.

**Expected formats:**
- `project-version-os-arch` (e.g., `claude-trace-pr-27.1-linux-amd64`)
- `project--version-os-arch` (e.g., `claude-trace--pr-27.1-linux-amd64`)

**How prrun will help:**
When searching for assets, prrun now displays all available assets:
```
Available assets (4):
  - claude-trace-pr-27.1-linux-amd64
  - claude-trace-pr-27.1-linux-arm64
  - claude-trace-pr-27.1-darwin-amd64
  - claude-trace-pr-27.1-darwin-arm64
```

**Solution:**
- prrun now handles both single-dash and double-dash naming formats automatically
- If you see assets listed but still get a 404, report this as a bug

#### 3. Private Release or Authentication Issue

The release exists but requires authentication to download.

**How to check:**
- Can you access the release URL in a web browser while logged out?
- Is the repository private?

**Solution:**
- Set a GitHub token: `export GITHUB_TOKEN=your_token_here`
- Or use the `gh` CLI which provides authentication automatically
- prrun will use the token for both API calls and downloads

#### 4. Project Not Found in Repository

The project directory doesn't exist or doesn't have the required files.

**Required for a project to be detected:**
- Directory with `go.mod` file
- At least one `main.go` file in the project

**Solution:**
- Ensure the project directory exists in the PR branch
- Verify `go.mod` and `main.go` are present
- Check that the GitHub Actions workflow ran successfully

## Release Workflow

### How Releases Are Created

1. **Trigger**: Push to main or PR activity
2. **Detect**: Find Go projects (directories with `go.mod` + `main.go`)
3. **Version**: Calculate next version number (e.g., `pr-27.1`)
4. **Build**: Build binaries for 4 platforms:
   - linux/amd64
   - linux/arm64
   - darwin/amd64 (macOS Intel)
   - darwin/arm64 (macOS Apple Silicon)
5. **Release**: Create GitHub release with tag and upload binaries

### Tag and Binary Naming

**Tag format:** `project--ref.number`
- Examples: `dissect--main.3`, `claude-trace--pr-27.1`
- Uses **double dash** (`--`) between project and ref

**Binary format:** `project-ref.number-os-arch`
- Examples: `dissect-main.3-linux-amd64`, `claude-trace-pr-27.1-darwin-arm64`
- Uses **single dash** (`-`) throughout

### Checking Workflow Logs

1. Go to the repository's Actions tab
2. Find the "Release Go Projects" workflow
3. Click on the specific workflow run for your PR
4. Check for errors in the "Build binaries" and "Create Release" steps

## Getting Help

If you encounter an issue not covered here:

1. Run prrun with verbose output (it will show all available assets)
2. Check the release page on GitHub to see what assets exist
3. Check the GitHub Actions logs for build errors
4. Create an issue with:
   - The exact command you ran
   - The full error output (including the assets list)
   - Link to the PR or release
   - Workflow logs if available
