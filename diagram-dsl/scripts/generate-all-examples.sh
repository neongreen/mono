#!/bin/bash

# Master script to generate all examples (both diagram-dsl and D2)
# This allows easy comparison between the two tools

set -e

cd "$(dirname "$0")/.."

echo "=================================================="
echo "Generating all diagram examples"
echo "=================================================="
echo

echo "Step 1: Generating diagram-dsl examples..."
echo "--------------------------------------------------"
npm run examples
echo

echo "Step 2: Generating D2 examples..."
echo "--------------------------------------------------"
./scripts/generate-d2-examples.sh
echo

echo "=================================================="
echo "All examples generated successfully!"
echo "=================================================="
echo
echo "diagram-dsl outputs: examples/*.svg"
echo "D2 outputs:          examples/d2-output/*.svg"
echo
