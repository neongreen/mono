#!/bin/bash
set -e

# Configure the book to use JSON Server instead of Supabase

cd "$(dirname "$0")/.."

# Create a configuration file that will be loaded by the book
CONFIG_FILE="example-book/json-server-config.js"
cat > "$CONFIG_FILE" <<EOF
// Auto-generated JSON Server configuration
// Do not edit manually - run scripts/configure-json-server.sh to regenerate

window.JSON_SERVER_CONFIG = {
    url: 'http://localhost:54322'
};
EOF

echo "Configuration written to $CONFIG_FILE"
echo "Ready to start mdbook!"
