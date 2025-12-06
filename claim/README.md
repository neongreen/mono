# claim

Lightweight claim-checking tool that uses Claude to verify claims are properly proven by their evidence.

## Overview

`claim` scans source files for `@claim` and `@lens` blocks embedded in comments, then uses Claude with structured output to judge whether each claim is proven by its bullets.

It's designed to catch false positives: claims that sound good but aren't actually proven by the code or reasoning provided.

**Note**: This tool makes real Claude API calls. Each `check` command costs ~$0.01-0.10. Development and testing is authorized up to $100/session - see [CLAUDE.md](CLAUDE.md) for details.

## Syntax

### Claims

A claim consists of a header and bullets:

```go
// @claim[unique-id]: Statement to be proven
// - First piece of evidence
// - Second piece of evidence
//   - Nested sub-evidence
// - @sorry
```

**Claim Header:**
- `@claim[id]` - Unique identifier (required)
- Tags like `@pedantic` - Optional, selects additional lenses
- `:` - Separates header from statement (required)
- Statement - What you're claiming is true

**Bullets:**
- Start with `-` after comment leader
- Indentation creates nesting (tabs = 2 spaces)
- Can reference other claims: `@claim[other-id]`
- Special: `@sorry` means "accepted without proof" (prevents "proven" result)

### Lenses

Lenses define how Claude checks claims. They're loaded from `claim/lenses.md` or via `--lens-file`:

```markdown
@lens[default]
You are a skeptical claim checker.
Never return proven if there's any plausible missing case.

@lens[pedantic]
Demand precise, unambiguous statements.
Reject vague quantifiers.
```

Lenses are selected by:
- `default` - always included if present
- Tags on claim header (e.g., `@pedantic`) match lens names

### Implicit Termination

Blocks end at the next `@claim` or `@lens` header, or EOF. No explicit end markers.

## Commands

### Index

Scan files and report claim/lens counts:

```bash
go run ./cmd/claim index
go run ./cmd/claim index --root fixtures
go run ./cmd/claim index --json  # Output index as JSON
```

Reports duplicate claim IDs (exits non-zero if found).

### Check

Check a specific claim:

```bash
go run ./cmd/claim check --claim t1
go run ./cmd/claim check --claim t1 --root fixtures
go run ./cmd/claim check --claim t1 --debug-prompt  # Show prompt sent to Claude
```

**Flags:**
- `--claim <id>` - Claim ID to check (required)
- `--root <dir>` - Root directory to scan (default: `.`)
- `--lens-file <path>` - Additional lens file to load
- `--max-ref-depth <n>` - Max depth for referenced claims (default: 3)
- `--debug-prompt` - Print the full prompt sent to Claude

**Exit codes:**
- `0` - Claim is unproven or sorry (expected for catching false positives)
- `1` - Tool error (parse error, missing claim, etc.)
- `2` - Claim is proven (test failure for golden suite)

**Environment:**
- `CLAIM_CLAUDE_CMD` - Override Claude command (default: `claude`)

### Golden

Run all test cases from `fixtures/cases.jsonl`:

```bash
go run ./cmd/claim golden
go run ./cmd/claim golden --root fixtures
```

Expects all claims to be unproven or sorry. Fails if any claim is proven.

## How It Works

1. **Scan** - Walk directory tree, read files, skip `.git`, `node_modules`, etc.
2. **Parse** - Extract `@claim` and `@lens` blocks using plain-text line scanning
3. **Index** - Build claim index, detect duplicates, track references
4. **Expand** - Recursively collect referenced claims (up to max depth)
5. **Prompt** - Build prompt with lenses, claim, bullets (with paths), and refs
6. **Run** - Execute `claude --print --output-format json --json-schema <schema> --tools ""`
7. **Validate** - Enforce post-rules:
   - Every bullet must have a verdict
   - Bullets with `@claim[ref]` must require that claim
   - Any `@sorry` bullet prevents "proven" result
8. **Report** - Print verdict and exit with appropriate code

## LLM Output Schema

Claude responds with structured JSON:

```json
{
  "claim_id": "string",
  "result": "proven|unproven|sorry",
  "bullets": [
    {
      "path": "string",
      "status": "ok|trivial|needs_split|needs_claim|contradicts|sorry",
      "required_claims": ["string"],
      "suggested_rewrite": ["string"]
    }
  ],
  "counterexample": "string"
}
```

Post-rules are enforced in code, so even if Claude makes a mistake, the tool catches it.

## Examples

### Basic Claim

```go
// @claim[no-panic]: This function never panics
// - All error cases return error values
// - No calls to panic() in the implementation
// - External calls are to panic-free standard library functions
```

### Claim with Reference

```go
// @claim[safe-server]: Server handles requests safely
// - Request validation prevents injection, see @claim[validate-input]
// - All panics are recovered in middleware
// - Timeouts prevent resource exhaustion

// @claim[validate-input]: Input validation is complete
// - SQL queries use parameterized statements
// - HTML output is escaped
// - File paths are sanitized
```

### Claim with Sorry

```go
// @claim[correct-algorithm]: Algorithm produces correct output
// - Base case returns correct result
// - @sorry
//   (Inductive case proof is complex, accepting without full verification)
```

## Testing

Run unit tests:

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

Test a specific fixture:

```bash
go run ./cmd/claim check --claim t1 --root fixtures
```

## Design Principles

1. **No AST parsing** - Works across languages using plain text
2. **Strict IDs** - Claim IDs are required and must be unique
3. **Implicit termination** - No end markers, blocks end at next header
4. **Post-rules in code** - Don't trust LLM, enforce rules programmatically
5. **Avoid false positives** - Default to unproven when uncertain

## License

Same as parent monorepo.
