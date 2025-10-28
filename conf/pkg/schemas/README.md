# Schema Files

This directory contains JSON schemas for configuration validation.

## Schema Source Policy

**Always use official schemas when available. Never write schemas by hand.**

### Priority Order
1. **SchemaStore.org**: Check https://www.schemastore.org/ first
2. **Official project schemas**: Use schemas from the official project documentation
3. **Custom schemas**: Only as a last resort when no official schema exists

### Metadata Files

Each schema file MUST have a corresponding `.meta` file documenting:
- **Source URL**: Where the schema was downloaded from
- **Download Date**: When it was downloaded (YYYY-MM-DD format)
- **Modifications**: Any changes made, by whom, and why
- **Update History**: Track of all updates and changes
- **License**: Schema license information

### Schema Files in This Directory

- **jj.json**: Official schema from Jujutsu project
  - Source: https://jj-vcs.github.io/jj/latest/config-schema.json
  - See: jj.json.meta

- **claude.json**: Schema from SchemaStore.org
  - Source: https://www.schemastore.org/claude-code-settings.json
  - See: claude.json.meta

- **mise.schema**: Custom schema (no official schema available)
  - Based on: https://mise.jdx.dev/configuration.html
  - See: mise.schema.meta

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
