# claim

Lightweight claim-checking tool that uses Claude to verify claims are properly proven by their proofs.

## Overview

`claim` scans source files for `@claim` and `@proof` blocks embedded in comments, then uses Claude with structured output to judge whether each claim is justified by its proof.

It's designed to catch false positives: claims that sound good but aren't actually proven by the reasoning provided.

**Note**: This tool makes real LLM API calls. Each `check` command costs ~$0.01-0.10. Development and testing is authorized up to $100/session - see [AGENTS.md](AGENTS.md) for details.

## Syntax

### Claims

A claim is an assertion to be proven:

```go
// @claim[unique-id]: Statement to be proven
```

### Proofs

A proof provides justification for a claim:

```go
// @proof[unique-id]:
// Proof body text goes here.
// Can be multiple lines.
// Can reference other claims with @see[other-id].
```

The proof:
- Must reference the same ID as the claim it proves
- Can be prose, bullets, or any combination (bullets have no special semantic meaning)
- Ends at the next `@claim`, `@proof`, or `@lens` header

### References (@see)

To reference another claim's statement as an axiom (assumed true):

```go
// @proof[foo]:
// Given @see[bar] (proven elsewhere), we know that X.
// Combined with @see[baz], this establishes Y.
// Therefore the claim follows.
```

When checking `foo`:
- `@see[bar]` inserts bar's **statement only** as an axiom
- The statement is marked "proven elsewhere, assume true"
- The proof of bar is NOT included - we don't re-verify it
- There is no recursive verification

### Context (@context)

To fetch additional code context for evidence verification:

```go
// @claim[my-claim]: The goroutine closes the channel on exit
// @context[my-claim]:
// function SafeGo in pkg/util/goroutine.go
// @proof[my-claim]:
// The goroutine is started with SafeGo which has panic recovery.
// The defer close() runs even if panic occurs because SafeGo's recover()
// catches panics but defer still executes after recover().
```

When checking `my-claim`:
- A separate Claude call fetches the content specified in @context
- The resolved content is included in the prompt for evidence verification
- Context is NOT used to narrow the claim's meaning - if context changes interpretation, the claim is marked unproven

**Important**: The claim must be unambiguous WITHOUT context. If "the function" could refer to multiple functions and context picks one, that's an ambiguous claim and will be rejected.

### Lenses

Lenses define how Claude checks claims. They're loaded from `lenses.md` or via `--lens-file`:

```markdown
@lens[default]
You are a skeptical claim checker.
Never return proven if there's any plausible missing case.

@lens[pedantic]
Demand precise, unambiguous statements.
Reject vague quantifiers.
```

## Quick Start

### Step 1: Write Your Claim

```go
// @claim[my-claim]: The function never returns nil
```

### Step 2: Write the Proof

```go
// @proof[my-claim]:
// The function has two return paths:
// - Line 10: returns a newly allocated struct (never nil)
// - Line 15: returns the cached instance (initialized in init(), never nil)
// Both paths return non-nil values.
```

### Step 3: Check It

```bash
go run ./claim check my-claim --root /path/to/project
```

## Commands

### check

Check a specific claim or all claims:

```bash
claim check my-claim-id                    # Check one claim
claim check my-claim-id --debug-prompt     # Show prompt sent to Claude
claim check --all                          # Check all claims
claim check --all --root ./src             # Check all in specific directory
```

**Flags:**
- `--root <dir>` - Root directory to scan (default: `.`)
- `--llm <name>` - LLM to use: `claude` or `codex` (default: `claude`)
- `--lens-file <path>` - Additional lens file to load
- `--debug-prompt` - Print the full prompt sent to LLM

**Exit codes:**
- `0` - Claim is proven
- `1` - Claim is unproven

### index

Scan files and report claim counts:

```bash
claim index
claim index --root fixtures
claim index --json
```

### golden

Run test cases from `fixtures/cases.jsonl`:

```bash
claim golden --root fixtures
```

## Examples

### Simple Claim

```go
// @claim[no-panic]: This function never panics
// @proof[no-panic]:
// All error cases return error values instead of panicking.
// No calls to panic() exist in the implementation.
// External calls are to panic-free standard library functions.
```

### Claim with Reference

```go
// @claim[channel-safe]: External code cannot close dataChan
// @proof[channel-safe]:
// By @see[field-unexported], external packages cannot access the field.
// The only way to close a channel is with close().
// Since external code has no access, it cannot call close().

// @claim[field-unexported]: dataChan is an unexported field
// @proof[field-unexported]:
// The field is declared with a lowercase name in Worker struct.
// Lowercase identifiers are unexported in Go.
```

## How It Works

1. **Scan** - Walk directory tree, read files
2. **Parse** - Extract `@claim`, `@proof`, and `@lens` blocks
3. **Index** - Build claim index, collect @see references
4. **Build Prompt** - Combine claim statement, axioms (from @see), and proof text
5. **Run LLM** - Execute Claude with structured output schema
6. **Report** - Print verdict with issues if unproven

## LLM Output Schema

Claude responds with:

```json
{
  "result": "proven|unproven",
  "issues": [
    {
      "title": "Short title",
      "description": "Explanation"
    }
  ]
}
```

## Design Principles

1. **Separation of concerns** - Claims state what, proofs explain why
2. **No recursive verification** - @see references are axioms, not re-verified
3. **One API call per claim** - Simple, predictable cost
4. **No semantic nesting** - Bullets in proofs are just formatting
5. **Avoid false positives** - Default to unproven when uncertain

## Claim Patterns

### Local vs Call-Tree Claims (for concurrency safety)

When proving properties like "no panic from send-on-closed-channel", use a two-level structure:

**Meta-definitions** (add as comments in your code):

```go
// === Meta-definitions for channel safety claims ===
//
// "local origin": A panic locally originates in function f if the innermost stack frame
// (the function whose source line executed the send) is f.
//
// "call-tree origin": A panic originates from the call tree of f if it occurs
// at any point before f returns, in f itself or any of its transitive callees.
```

**Local claim** (about one function's body):

```go
// @claim[MyFunc-no-send-on-closed-local]: No send-on-closed panic can locally originate in MyFunc
// @proof[MyFunc-no-send-on-closed-local]:
// MyFunc contains one send: `ch <- value` at line 50.
// The send is in a select with ctx.Done(), so it either succeeds or returns early.
// By @see[only-one-close-site], the channel is only closed after this function returns.
// Therefore no send-on-closed panic can locally originate in MyFunc.
```

**Call-tree claim** (about the entire call tree):

```go
// @claim[MyFunc-no-send-on-closed-call-tree]: No send-on-closed panic can originate from MyFunc's call tree
// @proof[MyFunc-no-send-on-closed-call-tree]:
// By @see[MyFunc-no-send-on-closed-local], no panic locally originates in MyFunc.
// Channel sends occur in these locations:
// 1. MyFunc itself (covered above)
// 2. helperFunc, called by MyFunc - By @see[helperFunc-no-send-on-closed-local], safe
// The channel is not passed to any other functions.
// Therefore no send-on-closed panic can originate from MyFunc's call tree.
```

### Channel Discipline Claims

For complex channel usage, document ownership:

```go
// @claim[channel-discipline]: resultChan has single-closer semantics
// @proof[channel-discipline]:
// - Created in NewWorker (one place)
// - Closed in exactly two sites: Run happy path, Cleanup error path
// - Both close sites are mutually exclusive (one runs, not both)
// - Goroutines have send-only access (chan<- Result type)
// - All senders respect context cancellation
```

## License

Same as parent monorepo.
