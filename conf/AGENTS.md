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

For each schema file:
- Track download date and source URL in `.meta` file alongside the schema
- Document any modifications with date, author, and reason
- Check for updates periodically

**Current schemas**:
- **jj**: Download from `https://jj-vcs.github.io/jj/latest/config-schema.json`
- **mise**: Custom schema definition based on documentation
- **claude**: From SchemaStore.org at `https://www.schemastore.org/claude-code-settings.json` (see `claude.json.meta` for details)
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
