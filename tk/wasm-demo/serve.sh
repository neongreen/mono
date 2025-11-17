#!/bin/bash

# Get the directory where this script is located
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR"

echo "Starting HTTP server on port 8080..."
echo "Open http://localhost:8080 in your browser"
echo ""
echo "Press Ctrl+C to stop the server"

# Try to use Python's http.server
if command -v python3 &> /dev/null; then
    python3 -m http.server 8080
elif command -v python &> /dev/null; then
    python -m http.server 8080
else
    echo "Error: Python not found. Please install Python or use another HTTP server."
    exit 1
fi
