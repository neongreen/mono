#!/bin/bash
# Show tk projects and help on Claude session start

echo "=== TK Projects ==="
tk project ls 2>/dev/null || echo "Could not list tk projects"

echo ""
echo "=== TK Help ==="
tk help 2>/dev/null || echo "Could not show tk help"
