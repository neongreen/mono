# AI Maintenance Notes

This project is maintained by automated agents. Use these guardrails while making changes:

## Rendering Changes
- Whenever a change can affect rendered output (layout engine tweaks, styling updates, arrow logic, SVG serialization), regenerate the examples **before** committing.
  - `pnpm run examples`
  - `pnpm run examples:d2` (if D2 comparisons are part of the change)
- Never estimate text metrics by multiplying string length by a constant—always call `measureText` from `src/layout/text-measurement.ts` (exported via `diagram-dsl`) so layout math stays accurate.

## Lint Expansion Ideas
- **Arrow routing suggestions:** Recommend alternate attachment sides when arrows bend sharply or travel long diagonals.
- **Connector spacing bands:** Warn when multiple parallel arrows lack consistent spacing, prompting grouping or reflow.

Feel free to expand this document with additional automation guidance as the rule set evolves.
