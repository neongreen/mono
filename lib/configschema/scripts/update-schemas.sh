#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA_DIR="$(cd "$SCRIPT_DIR/.." && pwd)/schemas"

echo "Updating configuration schemas..."

# Download JJ latest schema
echo "  Downloading JJ latest schema..."
curl -sSL "https://jj-vcs.github.io/jj/latest/config-schema.json" \
  | jq '.' > "$SCHEMA_DIR/jj/latest.json"

# Download JJ v0.34.0 schema (pinned) and fix duplicate enum
echo "  Downloading JJ v0.34.0 schema (pinned)..."
curl -sSL "https://jj-vcs.github.io/jj/v0.34.0/config-schema.json" \
  | jq '.properties["--when"].properties.platforms.items.enum |= unique' \
  > "$SCHEMA_DIR/jj/v0.34.0.json"

# Download Mise schema
echo "  Downloading Mise schema..."
curl -sSL "https://mise.jdx.dev/schema/mise.json" \
  | jq '.' > "$SCHEMA_DIR/mise/latest.json"

# Download Starship schema
echo "  Downloading Starship schema..."
curl -sSL "https://starship.rs/config-schema.json" \
  | jq '.' > "$SCHEMA_DIR/starship/latest.json"

# Download Claude schema
echo "  Downloading Claude schema..."
curl -sSL "https://json.schemastore.org/claude-code-settings.json" \
  | jq '.' > "$SCHEMA_DIR/claude/latest.json"

# Update metadata
DATE=$(date -u +"%Y-%m-%d")

cat > "$SCHEMA_DIR/jj/metadata.toml" <<EOF
[pinned]
version = "v0.34.0"
url = "https://jj-vcs.github.io/jj/v0.34.0/config-schema.json"
downloaded = "$DATE"
fixes = ["Removed duplicate 'hermit' from platforms enum"]

[latest]
url = "https://jj-vcs.github.io/jj/latest/config-schema.json"
downloaded = "$DATE"
EOF

cat > "$SCHEMA_DIR/mise/metadata.toml" <<EOF
[latest]
url = "https://mise.jdx.dev/schema/mise.json"
downloaded = "$DATE"
EOF

cat > "$SCHEMA_DIR/starship/metadata.toml" <<EOF
[latest]
url = "https://starship.rs/config-schema.json"
downloaded = "$DATE"
EOF

cat > "$SCHEMA_DIR/claude/metadata.toml" <<EOF
[latest]
url = "https://json.schemastore.org/claude-code-settings.json"
downloaded = "$DATE"
EOF

echo "Schema update complete!"
echo ""
echo "Updated schemas:"
echo "  - JJ (pinned v0.34.0): $(wc -c < "$SCHEMA_DIR/jj/v0.34.0.json" | tr -d ' ') bytes"
echo "  - JJ (latest): $(wc -c < "$SCHEMA_DIR/jj/latest.json" | tr -d ' ') bytes"
echo "  - Mise: $(wc -c < "$SCHEMA_DIR/mise/latest.json" | tr -d ' ') bytes"
echo "  - Starship: $(wc -c < "$SCHEMA_DIR/starship/latest.json" | tr -d ' ') bytes"
echo "  - Claude: $(wc -c < "$SCHEMA_DIR/claude/latest.json" | tr -d ' ') bytes"

