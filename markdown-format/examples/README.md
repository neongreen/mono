# Integration Examples

This directory contains example configurations and files for integrating markdown-format with various formatting tools.

## Files

- `treefmt.toml` - Example configuration for treefmt
- `markdown-format-inplace.sh` - Wrapper script for treefmt integration (makes markdown-format work in-place)
- `dprint.json` - Example configuration for dprint
- `sample-input.md` - Sample markdown file before formatting
- `sample-output.md` - Sample markdown file after formatting with markdown-format

## Testing the Integration

### With treefmt

1. Install treefmt (https://github.com/numtide/treefmt)
2. Copy `treefmt.toml` and `markdown-format-inplace.sh` to your project root
3. Update the path in `markdown-format-inplace.sh` to point to your markdown-format binary
4. Run `treefmt` to format all markdown files

### With dprint

1. Install dprint (https://dprint.dev/)
2. Copy `dprint.json` to your project root
3. Update the command path to point to your markdown-format binary
4. Run `dprint fmt` to format all markdown files

## Verifying the Integration

You can test the integration by:

1. Creating a copy of `sample-input.md` in your project
2. Running your formatter (treefmt or dprint)
3. Comparing the result with `sample-output.md`

The output should match, with one sentence per line while preserving all markdown structure.
