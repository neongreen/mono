# Agent Instructions for claim

## Overview

`claim` is a lightweight claim-checking tool that uses Claude to verify that claims embedded in source code are properly proven by their supporting evidence (bullets).

## Budget Authorization

**AUTHORIZED: Up to $100 per session for Claude API usage**

This is an explicit authorization to spend up to $100 per session on Claude API calls for development, testing, and verification of the `claim` tool.

## Expected Claude Usage

**IMPORTANT**: This project is expected to make real Claude API calls with real costs when:

1. **Running `claim check`** - Each check makes one Claude API call to verify a claim (~$0.01-0.10)
2. **Running `claim golden`** - Runs checks on all 9 fixtures (9 Claude API calls ~$0.10-1.00)
3. **Testing during development** - End-to-end tests require real Claude calls
4. **Debugging and iteration** - Test as many times as needed

**When working on this project, agents should:**
- Use real Claude API calls freely - you have a $100/session budget
- Not hesitate to spend tokens on testing - it's expected and necessary
- Test end-to-end functionality with actual Claude calls, not just mocks
- Run multiple iterations to verify fixes work correctly
- Document findings from real Claude runs
- Use `--model haiku` for speed/cost optimization only when appropriate
- Default to quality (sonnet) over cost savings

## Architecture

### Key Components

1. **Scanner** (`internal/scan`) - File traversal with ignore patterns
2. **Parser** (`internal/parse`) - Plain-text parsing of `@claim` and `@lens` blocks
3. **Index** (`internal/index`) - Build claim index, detect duplicates, track references
4. **Prompt Builder** (`internal/prompt`) - Construct prompts with lenses, claims, and source context
5. **Claude Runner** (`internal/runner`) - Execute `claude --print --output-format json --json-schema`
6. **Check Logic** (`internal/check`) - Enforce post-rules and validate results

### Key Design Decisions

**No AST parsing**: Works across all languages using plain text scanning

**Comment leader stripping**: Preserves indentation after comment markers for nested bullets
- `//   - child` becomes `   - child` (preserves 3 spaces)
- Important for building bullet trees with correct nesting

**Source code context**: Claims include 10 lines before/after for Claude to see actual code
- Added because Claude needs to verify bullets against real implementation
- Prevents "can't read file" errors when tools are disabled

**Filename anonymization**: Golden tests use `<source>` instead of real filenames
- Prevents Claude from cheating by seeing names like `t1_send_after_close.go`
- Regular checks show real filenames for debugging

**Post-rule enforcement**: Code validates Claude's output, doesn't blindly trust it
- Every bullet must have a verdict
 - Bullets with `@claim[ref]` must list ref in `required_claims`
 - Any `@later` bullet causes exit code 2 (proven but has deferred work)

## Testing Strategy

### Unit Tests
- Parser tests verify claim/lens extraction, nesting, @later detection
- Run with: `go test ./...`

### Integration Tests
- Use `MockRunner` to test check logic without Claude calls
- Fast, deterministic, no API costs

### End-to-End Tests
- **Must use real Claude** - This is expected and necessary
- Test with: `/tmp/claim check --claim t2 --root fixtures --lens-file lenses.md`
- Example output:
  ```
  Claim: t2
  Result: unproven
  Bullet Verdicts:
    [0] status=contradicts
    [1] status=contradicts
  Counterexample: The function abc() calls close(ch) twice...
  ```

### Golden Suite
- 9 deliberately flawed fixtures in `fixtures/`
- All should return "unproven", never "proven"
- Run with: `claim golden --root fixtures`
- **Will make 9 Claude API calls** - this is expected

## Common Issues

### "Missing verdict for bullet path"
Claude didn't return verdicts for all bullets. This is caught by post-rule enforcement.
- May happen on complex claims
- Post-rules are working correctly by catching this

### "No lenses loaded"
Lenses aren't in the scanned directory.
- Solution: Use `--lens-file claim/lenses.md`
- Or copy `lenses.md` to the fixtures directory

### Lenses

Two lenses are provided:
- **default**: Skeptical, avoids false positives
- **pedantic**: Extremely strict, demands precision

Add `@pedantic` tag to claim header to use pedantic lens.

## Development Workflow

When adding features or fixing bugs:

1. Create tk tasks: `tk new "Feature description" --project claim`
2. Implement changes
3. Run unit tests: `go test ./...`
4. **Test with real Claude**: `claim check --claim <id> ...`
5. Mark task done: `tk mark claim-X done`

## Example Claude Session

```bash
# Build
cd claim && go build

# Test on one fixture
./claim check --claim t2 --root fixtures --lens-file lenses.md

# See the full prompt sent to Claude
./claim check --claim t2 --root fixtures --lens-file lenses.md --debug-prompt

# Run full golden suite (9 Claude calls)
./claim golden --root fixtures --lens-file lenses.md
```

## Cost Estimates

Approximate costs per run (varies by model):
- Single check: ~$0.01-0.10
- Golden suite (9 checks): ~$0.10-1.00

These are rough estimates. Actual costs depend on:
- Claim complexity
- Source context size
- Number of referenced claims
- Model used (can override with env: `CLAIM_CLAUDE_CMD="claude --model haiku"`)
