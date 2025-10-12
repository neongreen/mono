#!/bin/bash

# Script to regenerate all sample PDFs
# This demonstrates the various input sources supported by printpdf

set -e

cd "$(dirname "$0")/.."

echo "Building printpdf..."
go build -o printpdf ./cmd

echo ""
echo "Generating samples..."
echo ""

# Clean old samples
rm -f samples/*.pdf

# Local markdown file
echo "1. Converting local Markdown file (sample.md)..."
./printpdf -converters weasyprint -o samples samples/sample.md
mv samples/output-weasyprint.pdf samples/local-markdown-sample.pdf

# GitHub file
echo "2. Converting GitHub file (golang/go README)..."
./printpdf -converters weasyprint -o samples https://github.com/golang/go/blob/master/README.md
mv samples/output-weasyprint.pdf samples/github-golang-readme.pdf

# Local HTML file with readability
echo "3. Converting local HTML file with Readability (sample.html)..."
./printpdf -converters weasyprint -o samples samples/sample.html
mv samples/output-weasyprint.pdf samples/html-readability-sample.pdf

# Self-documentation
echo "4. Converting printpdf README..."
./printpdf -converters weasyprint -o samples README.md
mv samples/output-weasyprint.pdf samples/printpdf-readme.pdf

echo ""
echo "Done! Generated samples:"
ls -lh samples/*.pdf

echo ""
echo "Note: These samples use WeasyPrint converter."
echo "To try other converters, install Typst or Prince and run:"
echo "  ./printpdf -converters typst,prince,weasyprint -o samples samples/sample.md"
