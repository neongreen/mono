#!/bin/bash

# Script to generate SVG files from D2 diagram files
# This allows comparison between diagram-dsl and D2 outputs

set -e

D2_DIR="examples/d2"
D2_OUTPUT_DIR="examples/d2-output"

echo "Generating D2 SVG examples..."
echo

# Create output directory if it doesn't exist
mkdir -p "$D2_OUTPUT_DIR"

# Generate SVGs from all D2 files
for d2_file in "$D2_DIR"/*.d2; do
  filename=$(basename "$d2_file" .d2)
  output_file="$D2_OUTPUT_DIR/${filename}.svg"
  
  echo "  Generating ${filename}.svg from D2..."
  d2 "$d2_file" "$output_file" --theme=200 2>&1 | grep -v "^success:" || true
  
  if [ -f "$output_file" ]; then
    echo "  ✓ Generated ${filename}.svg"
  else
    echo "  ✗ Failed to generate ${filename}.svg"
  fi
done

echo
echo "D2 examples generated successfully!"
echo "Output directory: $D2_OUTPUT_DIR"
