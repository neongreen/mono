# Agent Guidelines for conf

## Project-Specific Rules

- **Schema handling**: Use `santhosh-tekuri/jsonschema/v5` for JSON schema parsing
- **TOML editing**: Use `github.com/neongreen/mono/lib/toml` for surgical editing with comment preservation, including conf's own files under `~/.config/conf/`
- **JSON editing**: Use `encoding/json` with `MarshalIndent` for pretty-printing JSON target configs
- **TOML serialization**: Use `pelletier/go-toml/v2` for struct marshaling/unmarshaling in internal config files
- **No validation**: conf should surgically set/unset keys without validating surrounding config
- **Global configs only**: Only handle user-level configuration files, not project-specific ones

## Target Configuration Files

- **jj**: `~/.jjconfig.toml` (TOML format, schema: embedded jj.json)
- **mise**: `~/.config/mise/config.toml` (TOML format, custom schema)
- **starship**: `~/.config/starship.toml` (TOML format)
- **claude**: `~/.config/claude/config.json` (JSON format, schema: embedded claude.json)
- **conf itself**: `~/.config/conf/config.toml` (TOML format)

## Schema Sources

**Policy**: Always use official schemas from SchemaStore.org (https://www.schemastore.org/) when available. Never write schemas by hand.

### ⚠️ ABSOLUTELY NOT ALLOWED ⚠️

**Never claim a hand-written schema comes from an official source.**

This is misleading and will cause users to trust incorrect information. Examples of what is NOT allowed:
- Writing a schema by hand and claiming it came from SchemaStore.org
- Creating a schema based on documentation and calling it "official"
- Modifying an official schema without clearly documenting all changes
- Failing to track the actual provenance of a schema

**If you write a schema by hand:**
1. Clearly mark it as hand-written in all documentation
2. Add prominent warnings in the metadata file
3. Explain why an official schema couldn't be used
4. Create a plan to replace it with an official schema ASAP

### Schema Tracking Requirements

For each schema file:
- Track download date and source URL in `.meta` file alongside the schema
- Document any modifications with date, author, and reason
- Check for updates periodically
- Be 100% truthful about the schema's actual origin

**Current schemas**:
- **jj**: Download from `https://jj-vcs.github.io/jj/latest/config-schema.json`
- **mise**: Mise has an official schema at `https://mise.jdx.dev/schema/mise.json` (to be integrated)
- **claude**: Official schema from `https://json.schemastore.org/claude-code-settings.json` (see `claude.json.meta` for details)
- **Embedded**: Bundle schemas in binary for offline use

## CLI Interface Pattern

```bash
conf <tool> <dotted.config.path> <value>
```

Examples:
- `conf jj snapshot.max-new-file-size 0`
- `conf jj user.name "Alice"`  
- `conf mise tasks.python.run "python3 {}"`

## Testing Strategy

- **Unit tests**: TOML editing, schema parsing, completion generation
- **Integration tests**: Real config file manipulation for jj and mise
- **CLI tests**: End-to-end command workflow verification

## Development Notes

- Follow monorepo patterns: standard `mise.toml`, CI workflow, etc.
- Generate shell completion scripts for bash/zsh/fish
- Preserve existing TOML formatting, comments, and structure
- Use surgical editing - modify only the target key/value pair
