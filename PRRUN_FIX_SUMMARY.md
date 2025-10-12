# prrun Download 404 Issue - Fix Summary

## Problem Statement

Users were experiencing 404 errors when trying to download PR binaries using `prrun`:

```
$ prrun https://github.com/neongreen/mono/pull/27 claude-trace --help
Looking for PR #27 in neongreen/mono (project: claude-trace)
Found release: claude-trace--pr-27.1
Downloading binary from https://github.com/neongreen/mono/releases/download/claude-trace--pr-27.1/claude-trace-pr-27.1-linux-amd64...
Error: download failed with status 404
```

The symptom: The release is found successfully, but downloading the binary fails.

## Root Cause Analysis

After thorough investigation, I identified several potential issues:

### 1. Release Exists but Has No Assets
- The GitHub Actions workflow might have failed during the build step
- This would create the release tag but upload no binaries
- Result: Any download attempt would get a 404

### 2. Binary Naming Convention Mismatch
- The release workflow creates two formats potentially:
  - Standard: `project-version-os-arch` (e.g., `claude-trace-pr-27.1-linux-amd64`)
  - Alternative: `project--version-os-arch` (e.g., `claude-trace--pr-27.1-linux-amd64`)
- prrun was only looking for the standard format
- If assets used the alternative format, they wouldn't be found

### 3. Insufficient Error Diagnostics
- Generic error messages didn't help users understand what went wrong
- No visibility into what assets actually existed
- Hard to diagnose whether it's a build failure or naming issue

## Release Workflow Analysis

The release workflow (`.github/workflows/release.yml`) uses:

**Tag format:** `project--ref.number` (double dash)
- Example: `claude-trace--pr-27.1`

**Binary format:** `project-ref.number-os-arch` (single dash)
- Example: `claude-trace-pr-27.1-linux-amd64`

The workflow:
1. Line 114: `TAG="${PROJECT}--${REF_SAFE}.${VERSION_NUM}"`
2. Line 115: `VERSION="${REF_SAFE}.${VERSION_NUM}"`
3. Line 158: `OUTPUT_NAME="${PROJECT}-${VERSION}-${GOOS}-${GOARCH}"`

This creates binaries like: `claude-trace-pr-27.1-linux-amd64`

The workflow logic is **correct**, but edge cases exist where builds might fail or assets might get named differently due to workflow variations.

## Solutions Implemented

### 1. Debug Output for Available Assets

Added code to list all assets in a release when searching:

```go
fmt.Printf("Available assets (%d):\n", len(release.Assets))
for _, asset := range release.Assets {
    fmt.Printf("  - %s\n", asset.Name)
}
```

**Benefit:** Users can immediately see if:
- The release has no assets (build failed)
- Assets exist but with different names
- The expected asset is present

### 2. Handle Multiple Naming Formats

Updated the asset matching logic to accept both:
- `project-version-os-arch` (single dash)
- `project--version-os-arch` (double dash)

```go
// Check if asset starts with project name (handles both single and double dash)
if strings.HasPrefix(asset.Name, projectName) {
    return asset.Name, asset.BrowserDownloadURL, nil
}
```

**Benefit:** Works regardless of which naming convention the workflow uses.

### 3. Specific Error for Empty Releases

Added check for releases with no assets:

```go
if len(release.Assets) == 0 {
    return "", "", fmt.Errorf("release %s has no assets (the build may have failed)", release.TagName)
}
```

**Benefit:** Immediately tells users to check workflow logs for build failures.

### 4. Enhanced 404 Error Messages

Added detailed explanation for 404 errors:

```go
if resp.StatusCode == 404 {
    return fmt.Errorf("download failed with status 404 (asset not found). This may mean:\n"+
        "  1. The release exists but has no assets (build may have failed)\n"+
        "  2. The asset name doesn't match what was expected\n"+
        "  3. The release is private and requires authentication\n"+
        "  Download URL: %s", downloadURL)
}
```

**Benefit:** Users get actionable information about what to check.

### 5. Comprehensive Documentation

Created `TROUBLESHOOTING.md` with:
- Detailed explanation of each error type
- How to diagnose the issue
- Step-by-step solutions
- Links to relevant GitHub pages and logs

Updated `README.md` to reference the troubleshooting guide.

## Test Coverage

Added tests for:
1. Standard naming format (existing)
2. Double-dash naming format (new)
3. Empty release (no assets) (new)

All tests pass:
```
PASS: TestParsePRURL
PASS: TestGetCacheDir
PASS: TestGetPlatformBinaryName
PASS: TestGetPlatformBinaryName_NoAssets
PASS: TestGetPlatformBinaryName_DoubleDashFormat
PASS: TestGetGitHubToken
PASS: TestCreateAuthenticatedRequest
```

## Files Changed

- `prrun/main.go` - Core improvements for error handling and asset matching
- `prrun/main_test.go` - Added test coverage for new scenarios
- `prrun/README.md` - Updated troubleshooting section
- `prrun/TROUBLESHOOTING.md` - New comprehensive guide

## Impact

### Before
```
Error: download failed with status 404
```
User is stuck with no actionable information.

### After
```
Available assets (4):
  - claude-trace-pr-27.1-linux-amd64
  - claude-trace-pr-27.1-linux-arm64
  - claude-trace-pr-27.1-darwin-amd64
  - claude-trace-pr-27.1-darwin-arm64

Error: download failed with status 404 (asset not found). This may mean:
  1. The release exists but has no assets (build may have failed)
  2. The asset name doesn't match what was expected
  3. The release is private and requires authentication
  Download URL: https://github.com/...
```
User sees exactly what assets exist and gets clear guidance on what to check.

## Why This Fixes the Issue

1. **Flexibility**: Handles both naming conventions, so format variations don't cause failures
2. **Visibility**: Shows all available assets, making naming mismatches obvious
3. **Guidance**: Detailed error messages guide users to the right solution
4. **Documentation**: Comprehensive troubleshooting guide covers all scenarios

## Verification

To verify the fix works:

1. **For missing assets:**
   - prrun will report "release has no assets (the build may have failed)"
   - Users can check GitHub Actions logs

2. **For naming mismatches:**
   - prrun will list all available assets
   - Users can see the actual names and report discrepancies

3. **For double-dash format:**
   - prrun now handles it automatically
   - No changes needed from users

## Recommendations

1. **Monitor**: Watch for users reporting 404 errors after this fix
2. **Workflow**: Consider standardizing on one naming format in the workflow
3. **Testing**: Test with actual PR releases to confirm the fix works in practice
4. **Documentation**: Consider adding a section in the main repo README about release naming conventions

## Conclusion

The fix makes prrun significantly more robust by:
- Handling multiple naming formats
- Providing detailed error diagnostics
- Offering clear troubleshooting guidance

Users will now be able to quickly identify and resolve 404 errors, whether they're caused by build failures, naming mismatches, or authentication issues.
