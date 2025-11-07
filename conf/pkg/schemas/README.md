# Schema Files

This directory contains JSON schemas for configuration validation.

## Schema Source Policy

**Always use official schemas when available. Never write schemas by hand.**

## ⚠️ ABSOLUTELY NOT ALLOWED ⚠️

**Never claim a hand-written schema comes from an official source.**

This is misleading and will cause users to trust incorrect information. If you write a schema by hand, you MUST:
1. Clearly mark it as hand-written in ALL documentation
2. Add prominent warnings in the metadata file and the schema itself
3. Explain why an official schema couldn't be used
4. Never claim it came from SchemaStore.org or any official source

### Priority Order
1. **SchemaStore.org**: Check https://www.schemastore.org/ first
2. **Official project schemas**: Use schemas from the official project documentation
3. **Custom schemas**: Only as a last resort when no official schema exists, and MUST be clearly marked

### Metadata Files

Each schema file MUST have a corresponding `.meta` file documenting:
- **Source URL**: Where the schema was downloaded from (or "HAND-WRITTEN" if not downloaded)
- **Download Date**: When it was downloaded (YYYY-MM-DD format)
- **Modifications**: Any changes made, by whom, and why
- **Update History**: Track of all updates and changes
- **License**: Schema license information

### Schema Files in This Directory

- **jj.json**: Official schema from Jujutsu project
  - Source: https://jj-vcs.github.io/jj/latest/config-schema.json
  - Updated: 2025-10-28
  - Format: JSON Schema describing TOML configs
  - See: jj.json.meta

- **claude.json**: Official schema from SchemaStore.org
  - Source: https://json.schemastore.org/claude-code-settings.json
  - Format: JSON Schema describing JSON configs
  - See: claude.json.meta

- **mise.schema**: Custom TOML-based schema (used by parser)
  - Based on: https://mise.jdx.dev/configuration.html
  - Format: Custom TOML schema format
  - See: mise.schema.meta

- **mise.json**: Official JSON schema from Mise project (reference)
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
