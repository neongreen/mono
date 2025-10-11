# Release Workflow Fix

## Problem

The release workflow for Go projects was failing because the project detection script was not correctly building the JSON array of projects. While the script found the projects (`dissect` and `markdown-format`), it failed to pass them to the release job.

## Root Cause

Line 53 of `.github/workflows/release.yml` had a bug:

```bash
printf -v projects_json '%s\n' "${projects[@]}" | jq -R . | jq -s .
```

The issue was using `printf -v` (which stores formatted output in a variable) combined with pipes. When you use `printf -v variable`, the output goes into the variable, but then piping that doesn't work as expected. The variable gets the printf output, not the jq result.

Additionally, the jq output was multi-line, which doesn't work well with GitHub Actions output variables.

## Solution

Changed line 53 to:

```bash
projects_json=$(printf '%s\n' "${projects[@]}" | jq -R . | jq -s -c .)
```

Two key fixes:
1. **Replaced `printf -v` with command substitution `$()`** - This properly captures the final output of the pipe chain
2. **Added `-c` flag to jq** - This produces compact (single-line) JSON, which works correctly with GitHub Actions output variables

## Testing

Verified the fix locally:

```bash
# Before fix (produces empty array)
printf -v projects_json '%s\n' "${projects[@]}" | jq -R . | jq -s .
# projects_json is empty/wrong

# After fix (produces correct JSON)
projects_json=$(printf '%s\n' "${projects[@]}" | jq -R . | jq -s -c .)
# projects_json = ["dissect","markdown-format"]
```

## Expected Behavior After Fix

When the workflow runs:
1. Detect job will correctly output: `projects=["dissect","markdown-format"]`
2. Release job will trigger with a matrix strategy
3. Two parallel jobs will build and release:
   - `dissect--main.1` (or next version)
   - `markdown-format--main.1` (or next version)
4. Each release will include 4 binaries (Linux amd64/arm64, macOS amd64/arm64)

## Files Changed

- `.github/workflows/release.yml` - Line 53 only (minimal surgical fix)
