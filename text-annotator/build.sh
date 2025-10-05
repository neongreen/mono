#!/bin/bash
set -e

echo "Building TextAnnotator..."
swift build -c release

echo ""
echo "Build complete!"
echo "Run with: .build/release/TextAnnotator"
