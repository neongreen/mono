#!/bin/bash
set -e

# Extract Supabase credentials and configure the book
# This script reads the local Supabase status and injects config into the JavaScript

cd "$(dirname "$0")/.."

# Check if Supabase is running
if ! supabase status > /dev/null 2>&1; then
    echo "Error: Supabase is not running. Start it with 'supabase start'"
    exit 1
fi

# Get Supabase configuration
echo "Extracting Supabase configuration..."

# Extract API_URL and ANON_KEY from supabase status
# Note: The output format is: VARIABLE_NAME="value"
API_URL=$(supabase status -o env 2>/dev/null | grep "^API_URL=" | cut -d'=' -f2 | tr -d '"')
ANON_KEY=$(supabase status -o env 2>/dev/null | grep "^ANON_KEY=" | cut -d'=' -f2 | tr -d '"')

if [ -z "$API_URL" ] || [ -z "$ANON_KEY" ]; then
    echo "Error: Could not extract Supabase credentials"
    echo "API_URL: $API_URL"
    echo "ANON_KEY: ${ANON_KEY:0:20}..."
    exit 1
fi

echo "Supabase URL: $API_URL"
echo "Anon key: ${ANON_KEY:0:20}..."

# Create a configuration file that will be loaded by the book
CONFIG_FILE="example-book/supabase-config.js"
cat > "$CONFIG_FILE" <<EOF
// Auto-generated Supabase configuration
// Do not edit manually - run scripts/configure-supabase.sh to regenerate

window.SUPABASE_CONFIG = {
    url: '$API_URL',
    anonKey: '$ANON_KEY'
};
EOF

echo "Configuration written to $CONFIG_FILE"
echo "Ready to start mdbook!"
