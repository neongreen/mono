# prrun Troubleshooting Guide

## Debug Mode

When encountering issues, use the `--debug` flag to see detailed information about what prrun is doing:

```bash
prrun https://github.com/owner/repo/pull/123 --debug
```

Debug output includes:
- Number of releases fetched from each API page
- Total releases found
- Which releases match the PR number
- Which release is selected
- Platform detection and binary selection

Example debug output:
```
[DEBUG] Fetching releases page 1 from: https://api.github.com/repos/owner/repo/releases?per_page=100&page=1
[DEBUG] Found 100 releases on page 1
[DEBUG] Fetching releases page 2 from: https://api.github.com/repos/owner/repo/releases?per_page=100&page=2
[DEBUG] Found 23 releases on page 2
[DEBUG] Total releases fetched: 123
[DEBUG] Looking for PR #89 in owner/repo (project: )
[DEBUG] Found matching release: printpdf--pr-89.1 (prerelease=true)
[DEBUG] Using release: printpdf--pr-89.1
```

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

### No Releases Found for PR

**Symptoms:**
```
Error: no releases found for PR #123
```

**Possible Causes:**

#### 1. Release Workflow Not Run or Pending Approval

The GitHub Actions workflow may not have run yet or is waiting for approval.

**How to check:**
- Go to the PR on GitHub
- Click the "Checks" tab
- Look for the "Release Go Projects" workflow

**Solution:**
- If the workflow is pending approval, approve it
- If the workflow hasn't run, push a new commit or manually trigger it
- Wait for the workflow to complete before running prrun

#### 2. PR Beyond First Page of Releases (Fixed in Latest Version)

In older versions of prrun, only the first 30 releases were searched. This is now fixed with automatic pagination.

**Solution:**
- Update to the latest version of prrun
- Use `--debug` flag to see how many releases are being fetched:
  ```bash
  prrun https://github.com/owner/repo/pull/123 --debug
  ```

#### 3. Project Has No Go Code to Build

The PR doesn't modify any Go projects, so no binaries are built.

**Solution:**
- Check if the PR includes changes to Go projects
- Ensure the project has `go.mod` and `main.go` files
- Specify the project name explicitly if multiple projects exist

## Getting Help

If you encounter an issue not covered here:

1. Run prrun with `--debug` flag to see detailed information
2. Check the release page on GitHub to see what releases exist
3. Check the GitHub Actions logs for build errors
4. Create an issue with:
   - The exact command you ran (including `--debug` output)
   - The full error output
   - Link to the PR or release
   - Workflow logs if available
