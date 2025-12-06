# claim

Lightweight claim-checking tool that uses Claude to verify claims are properly proven by their evidence.

## Overview

`claim` scans source files for `@claim` and `@lens` blocks embedded in comments, then uses Claude with structured output to judge whether each claim is proven by its bullets.

It's designed to catch false positives: claims that sound good but aren't actually proven by the code or reasoning provided.

**Note**: This tool makes real LLM API calls. Each `check` command costs ~$0.01-0.10 depending on the LLM used. Development and testing is authorized up to $100/session - see [CLAUDE.md](CLAUDE.md) for details.

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

## Quick Start: Adding Claims to Your Project

### Step 1: Write Your First Claim

Add a claim as a comment in your source code:

```go
// @claim[unique-id]: Your statement about the code
// - First piece of evidence (bullet point)
// - Second piece of evidence
//   - Sub-evidence (nested with indentation)
```

**Key rules:**
- Claim ID must be unique across your project
- Use `:` to separate the claim ID/tags from the statement
- Bullets start with `-` after the comment marker
- Indent nested bullets with 2 spaces or 1 tab
- Claims work in any language (Go, TypeScript, Python, etc.)

### Step 2: Create a Lens File (Optional)

Create `lenses.md` in your project root to define checking behavior:

```markdown
@lens[default]
You are a skeptical claim checker.
Never return proven if there's any plausible missing case.
If a bullet is vague, mark it as needs_split.

@lens[pedantic]
Demand precise, unambiguous statements.
Reject vague quantifiers like "usually" or "often".
```

The `default` lens is always applied. Other lenses are applied when you tag claims (e.g., `@claim[id] @pedantic:`).

### Step 3: Check Your Claim

From the monorepo root:

```bash
# With Claude (default)
go run ./claim check --claim unique-id --root /path/to/your/project

# With Codex
go run ./claim check --claim unique-id --root /path/to/your/project --llm codex

# With your lens file
go run ./claim check --claim unique-id --root /path/to/your/project --lens-file /path/to/lenses.md
```

### Step 4: Check All Claims

Check all claims in your project:

```bash
# First, get all claim IDs
go run ./claim index --root /path/to/your/project

# Then check each one
for claim in id1 id2 id3; do
  echo "=== Checking $claim ==="
  go run ./claim check --claim $claim --root /path/to/your/project --llm codex
done
```

### Expected Results

Claims should be **unproven** or **sorry** - that's the goal! This tool catches false positives where you claim something but don't actually prove it.

- `unproven` - The LLM found a counterexample or missing case
- `sorry` - You explicitly marked a bullet as `@sorry` (accepted without proof)
- `proven` - All bullets are acceptable (rare, means you really did prove it)

## Commands

### Index

Scan files and report claim/lens counts:

```bash
go run ./claim index
go run ./claim index --root fixtures
go run ./claim index --json  # Output index as JSON
```

Reports duplicate claim IDs (exits non-zero if found).

### Check

Check a specific claim:

```bash
go run ./claim check --claim t1
go run ./claim check --claim t1 --root fixtures --llm codex
go run ./claim check --claim t1 --debug-prompt  # Show prompt sent to LLM
```

**Flags:**
- `--claim <id>` - Claim ID to check (required)
- `--root <dir>` - Root directory to scan (default: `.`)
- `--llm <name>` - LLM to use: `claude` or `codex` (default: `claude`)
- `--lens-file <path>` - Additional lens file to load
- `--max-ref-depth <n>` - Max depth for referenced claims (default: 3)
- `--debug-prompt` - Print the full prompt sent to LLM

**Exit codes:**
- `0` - Claim is unproven or sorry (expected for catching false positives)
- `1` - Tool error (parse error, missing claim, etc.)
- `2` - Claim is proven (test failure for golden suite)

**Environment:**
- `CLAIM_CLAUDE_CMD` - Override Claude command (default: `claude`)
- `CLAIM_CODEX_CMD` - Override Codex command (default: `codex`)

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
6. **Run** - Execute LLM (Claude or Codex) with structured output schema
   - Claude: `claude --verbose --print --output-format stream-json --json-schema <schema>`
   - Codex: `codex exec --output-schema <file> --json`
7. **Validate** - Enforce post-rules:
   - Every bullet must have a verdict
   - Bullets with `@claim[ref]` must require that claim
   - Any `@sorry` bullet prevents "proven" result
   - Normalize paths to handle LLM variations
8. **Report** - Print verdict with colors and exit with appropriate code

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
