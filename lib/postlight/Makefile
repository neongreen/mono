# Makefile for building Postlight Parser WASM module

.PHONY: all clean install build-js build-wasm

# Default target
all: build-wasm

# Install npm dependencies
install:
	cd js && npm install

# Build the bundled JavaScript
build-js: install
	cd js && npm run build

# Build WASM module using javy
# Note: Requires javy to be installed (https://github.com/bytecodealliance/javy)
build-wasm: build-js
	@echo "Building WASM module with javy..."
	@if ! command -v javy >/dev/null 2>&1; then \
		echo "Error: javy is not installed. Install it from https://github.com/bytecodealliance/javy"; \
		echo "Download from: https://github.com/bytecodealliance/javy/releases"; \
		exit 1; \
	fi
	javy build js/dist/bundle.js -o parser.wasm
	@echo "✓ WASM module built: parser.wasm"

# Clean build artifacts
clean:
	rm -rf js/node_modules js/dist parser.wasm

# Help target
help:
	@echo "Postlight Parser WASM Build System"
	@echo ""
	@echo "Targets:"
	@echo "  make all        - Build everything (default)"
	@echo "  make install    - Install npm dependencies"
	@echo "  make build-js   - Bundle JavaScript code"
	@echo "  make build-wasm - Compile to WASM (requires javy)"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "Requirements:"
	@echo "  - Node.js and npm"
	@echo "  - javy (download from: https://github.com/bytecodealliance/javy/releases)"
