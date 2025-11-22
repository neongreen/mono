#!/usr/bin/env bash
# Script to regenerate lion's documentation using lion itself

set -euo pipefail

cd "$(dirname "$0")"

echo "Regenerating lion documentation..."
go run . generate . --output ./docs

echo "Documentation regenerated in ./docs/"
echo ""
echo "Generated files:"
ls -1 docs/
