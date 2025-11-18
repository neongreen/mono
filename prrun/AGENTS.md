# Agent Guidelines for prrun

## Build, Test, and Run Commands

**All commands must be run from the repository root (`/home/user/mono`).**

```bash
# Build
go build ./prrun

# Test
go test ./prrun/...

# Run
go run ./prrun [args...]

# Install (builds and places in $GOPATH/bin)
go install ./prrun
```

**Important:** Use `go` commands directly. Do not use `mise` for building or running prrun.

## Postmortems

### Postmortem: Missing Releases Due to API Pagination (2025-01-13)

**Timeline:**
1. Initial implementation: `findPRRelease()` and `findAllPRReleases()` fetched releases from `/repos/:owner/:repo/releases` endpoint without pagination
2. The code worked correctly when the repository had fewer than 30 releases
3. As the repository grew beyond 30 releases, older PR releases stopped being found
4. User reported: "Error: no releases found for PR #89" despite a prerelease existing for that PR
5. Investigation revealed that the GitHub API returns 30 releases per page by default
6. The release `printpdf--pr-89.1` existed but was beyond the first page of results

**Root Cause:**
- GitHub's releases API endpoint (`/repos/:owner/:repo/releases`) returns paginated results
- By default, only 30 releases are returned per page
- The code only fetched the first page, never checking for additional pages
- As the repository accumulated more releases, older releases became invisible to prrun

**What Could Have Caught This Earlier:**
1. Integration tests with a repository containing more than 30 releases
2. Warning comments in code noting the pagination limitation
3. Debug logging to show how many releases were fetched (now implemented with --debug flag)
4. Checking GitHub API documentation for pagination during initial implementation

**Lessons Learned:**
- Always check GitHub API documentation for pagination behavior
- When fetching lists from GitHub API, always implement pagination support
- Add debug flags early to help diagnose issues in production
- Test with repositories at scale, not just small test cases

**Prevention Measures Added:**
1. Implemented `fetchAllReleases()` helper that handles pagination automatically
2. Added `--debug` flag to show exactly what's happening during execution
3. This postmortem serves as documentation for future developers

**Future Considerations:**
- Consider adding integration tests that work with actual GitHub repositories
- Consider adding telemetry or logging to track how many releases are being fetched
- Document pagination behavior in code comments
