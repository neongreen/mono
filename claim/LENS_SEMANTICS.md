# Lens Semantics

This document describes how lenses work in the claim checker.

## What is a Lens?

A lens is a set of instructions that guide how Claude checks claims. Different lenses have different "personalities" - some are more skeptical, some are more pedantic, some refuse to chase context.

## Lens Definition

Lenses are defined in markdown files using the `@lens[name]` syntax:

```markdown
@lens[default]
You are a skeptical claim checker.
Your job is to avoid false positives...

@lens[pedantic]
You are an extremely strict claim checker...
```

## Lens Loading

Lenses are loaded from two sources:

1. **Scanned directory**: Any `@lens[...]` blocks found in scanned files
2. **Lens file**: Additional lenses from `--lens-file <path>`

## Lens Selection

Which lenses are applied to a claim:

1. **Default lens**: Always included if a lens named `default` exists
2. **Tagged lenses**: If a claim has tags (e.g., `@claim[id] @pedantic`), the matching lens is included

Example:
```go
// @claim[my-claim] @pedantic: Something is true
```

This claim would get both the `default` lens and the `pedantic` lens.

## Current Limitations

### No Subgoal Scoping

Lenses apply to the entire claim, not individual bullets. You cannot:
- Apply a lens to only one bullet
- Apply different lenses to different parts of the proof

### No Specialized Reviewers

All lenses must give a verdict on the entire claim. You cannot:
- Have a lens that only comments on certain topics (e.g., "database expert")
- Have a lens that says "no opinion" on parts it doesn't understand

### No Lens Composition

Lenses are simply concatenated. There's no way to:
- Define a lens that extends another
- Override parts of a lens

## Future Considerations

If lenses need more structure (scope declarations, specialization), consider:
- Moving to YAML/TOML format with explicit fields
- Adding a `scope` field to limit what a lens comments on
- Adding a `extends` field for lens composition

## Available Lenses

See `lenses.md` for the current lens definitions:

- **default**: Skeptical, avoids false positives, prefers local reasoning
- **pedantic**: Extremely strict, demands precision, constructs counterexamples
- **local**: Refuses to chase context, only verifies from immediate bullets