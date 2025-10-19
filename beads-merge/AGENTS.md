# beads-merge Agent Guidelines

## Purpose

beads-merge is a 3-way merge tool specifically for beads `.jsonl` issue files. It implements intelligent merging based on issue identity and field-specific merge strategies.

## Testing Requirements

- All changes to merge logic must include unit tests
- Test both successful merge cases and conflict scenarios
- Integration tests should cover real-world merge patterns from `.beads/issues.jsonl`

## Merge Algorithm Constraints

### DO NOT change these without discussion:

1. Issue identity is based on `.id`, `.created_at`, and `.created_by` only
2. Timestamps (updated_at, closed_at) always use max value
3. Dependencies are always merged (deduplicated by issue_id + depends_on_id + type)
4. String fields use 3-way merge (prefer the side that changed from base)

### Fields that need special handling:

- Priority: integer field, handle like strings but compare as numbers
- Dependencies: array, must deduplicate based on composite key

## Common Issues

### Issue not merging when expected
Check that `.id`, `.created_at`, and `.created_by` match exactly in all three files. Even minor differences (timezone, precision) will cause issues to be treated as different.

### Dependencies being duplicated
Ensure the deduplication key includes all three fields: issue_id, depends_on_id, and type. Two dependencies with same issue_id and depends_on_id but different types are distinct.

### Incorrect conflict markers
Conflicts should only be generated when both sides changed the same field to different values (not when base == left or base == right).
