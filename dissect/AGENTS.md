# Agent Guidelines for dissect

## Postmortems

### Postmortem: Double-Star Glob Pattern Bug (2025-10-12)

**Timeline:**

1. **Initial Implementation (Commit 588cf0d)**: Added glob support for file paths using `filepath.Glob()`. Documentation and examples claimed support for `pkg/**/*.go` pattern for recursive directory matching.

2. **Review Finding**: Code reviewer discovered that `filepath.Glob()` does NOT support `**` for recursive directory traversal. The `*` wildcard in Go's `filepath.Glob()` only matches non-separator characters, meaning it cannot cross directory boundaries. The pattern `pkg/**/*.go` would only match immediate subdirectories, not nested ones.

3. **Missing Tests**: No tests were written to verify `**` pattern behavior. The initial test suite only tested single-level glob patterns like `*.go`, which worked correctly with `filepath.Glob()`.

4. **Fix (Commit [current])**: Replaced `filepath.Glob()` with `doublestar.FilepathGlob()` from `github.com/bmatcuk/doublestar/v4`, which properly implements `**` for recursive matching. Added comprehensive test `MoveWithDoubleStarPattern` to verify recursive matching across multiple directory levels.

**Root Cause:**

- Assumption that `filepath.Glob()` supports `**` pattern without verifying in Go documentation
- Documentation was written based on desired behavior rather than tested behavior
- No tests were created to verify the documented functionality

**What Could Have Caught This Earlier:**

1. **Read the documentation**: Check `go doc filepath.Glob` and `go doc filepath.Match` before claiming support for a pattern
2. **Test what you document**: Every example in documentation should have a corresponding test
3. **Manual verification**: Manually test complex patterns (like `**`) before documenting them
4. **Integration tests**: Create test directories with nested structures to verify recursive behavior

**Lessons Learned:**

- Always verify library behavior against documentation before making claims
- Write tests for edge cases and complex patterns, especially when documenting them as supported
- Don't assume glob patterns work the same across all implementations
- Test the actual use case (nested directories) not just simplified scenarios
