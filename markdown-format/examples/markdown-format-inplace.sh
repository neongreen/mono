#!/bin/bash
# Wrapper script to make markdown-format work in-place for treefmt
# This script takes file paths as arguments and formats them in place

# Adjust this path to where your markdown-format binary is located
MARKDOWN_FORMAT="./markdown-format/markdown-format"

for file in "$@"; do
    if [ -f "$file" ]; then
        # Create a temporary file
        temp_file=$(mktemp)
        # Format the file and save to temp
        "$MARKDOWN_FORMAT" "$file" > "$temp_file"
        # Replace original file with formatted content
        mv "$temp_file" "$file"
    fi
done
