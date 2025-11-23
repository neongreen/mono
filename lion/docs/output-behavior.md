# Output Behavior

## Stale files

Generate overwrites files for topics that still exist but does not delete files for topics that disappeared; stale topic files remain even though the index links only current topics, so clean them up yourself if needed.

*Source: `lion/internal/generator/generator.go:20`*

