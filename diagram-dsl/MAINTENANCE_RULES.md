# Maintenance Rules

Guidelines to keep rendered assets and examples consistent.

1. **Regenerate examples after rendering changes.** Whenever code alters how diagrams render (layout calculations, styling, SVG output, arrow logic, etc.), rerun `pnpm run examples` (and `pnpm run examples:d2` if D2 comparisons apply) before committing so the checked-in SVGs stay in sync.

Feel free to extend this list with additional maintenance expectations as the project evolves.
