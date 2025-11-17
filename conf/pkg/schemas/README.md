# Schema Files

This directory contains schema parsers and configuration for `conf`.

## Schema Source

**JSON schemas are now centralized in `lib/configschema`.**

The embedded schemas (jj, mise, starship, claude) are sourced from `lib/configschema` which maintains versioned schemas. To update schemas:

```bash
mise run //:schemas:update
```

See `lib/configschema/README.md` for more information.

## Files in This Directory

### Schema Parsers
- `jj_parser.go` - JJ schema parser and completion logic
- `claude_parser.go` - Claude schema parser
- `mise_json_parser.go` - Mise JSON schema parser
- `mise.go` - Mise TOML schema parser (custom format)
- `starship_parser.go` - Starship schema parser
- `jsonschema_parser.go` - Generic JSON schema parser

### Embed Files
- `embed.go` - Embeds JJ schema from lib/configschema
- `claude_embed.go` - Embeds Claude schema from lib/configschema
- `mise_json_embed.go` - Embeds Mise JSON schema from lib/configschema
- `starship_embed.go` - Embeds Starship schema from lib/configschema
- `mise_embed.go` - Embeds Mise custom TOML schema (local file)

### Local Schema Files
- `mise.schema` - Custom TOML-based schema for Mise
  - Format: Custom TOML schema format (not JSON Schema)
  - Used by Mise TOML parser
  - See: mise.schema.meta
  - Source: https://mise.jdx.dev/schema/mise.json
  - Downloaded: 2025-10-28
  - Format: JSON Schema describing TOML configs
  - Note: Mise configs are TOML; this JSON schema describes their structure
  - See: mise.schema.meta

- **starship.json**: Official schema from Starship project
  - Source: https://starship.rs/config-schema.json
  - Downloaded: 2025-11-07
  - Format: JSON Schema (Draft 2020-12) describing TOML configs
  - See: starship.json.meta

### Adding New Schemas

When adding a new schema:

1. Check SchemaStore.org first: https://www.schemastore.org/api/json/catalog.json
2. If not on SchemaStore, check the official project documentation
3. Download the schema (never write it by hand)
4. Create a `.meta` file with full provenance information
5. Add the `$id` and `$comment` fields to the JSON schema if not present
6. Update AGENTS.md with the new schema information

### Updating Schemas

To update an existing schema:

1. Download the latest version from the source URL
2. Update the `.meta` file with the new download date
3. Document any changes in the update history
4. Run tests to ensure compatibility
5. Commit with a clear message about what was updated

See `../../../AGENTS.md` for additional guidelines.
