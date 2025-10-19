# Multi-stage build for mdbook-comments demo
# This creates a working demo of mdbook with the comments plugin and Neon backend

FROM rust:1.90 AS builder

# Install mdbook
RUN cargo install --locked --version =0.4.52 mdbook

# Copy mdbook-comments source and build preprocessor
WORKDIR /build
COPY Cargo.toml Cargo.lock ./
COPY src ./src

# Build the preprocessor
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

# Install necessary runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy mdbook and the preprocessor binary
COPY --from=builder /usr/local/cargo/bin/mdbook /usr/local/bin/
COPY --from=builder /build/target/release/mdbook-comments /usr/local/bin/

# Set up the example book
WORKDIR /book
COPY example-book/src ./src
# CSS and JS assets are now embedded in the preprocessor binary

# Create book.toml configured for Neon
RUN echo '[book]' > book.toml && \
    echo 'authors = ["mdbook-comments contributors"]' >> book.toml && \
    echo 'language = "en"' >> book.toml && \
    echo 'multilingual = false' >> book.toml && \
    echo 'src = "src"' >> book.toml && \
    echo 'title = "Example Book with Comments (Neon Demo)"' >> book.toml && \
    echo '' >> book.toml && \
    echo '[preprocessor.comments]' >> book.toml && \
    echo '# Note: API URL is not used directly with Neon - the JS handles it' >> book.toml && \
    echo 'api-url = "http://localhost:3000/api"' >> book.toml && \
    echo 'auth-type = "cookie"' >> book.toml && \
    echo 'similarity-threshold = 0.85' >> book.toml && \
    echo 'orphaned-comments-location = "end-of-chapter"' >> book.toml && \
    echo '' >> book.toml && \
    echo '[preprocessor.comments.elements]' >> book.toml && \
    echo 'paragraphs = true' >> book.toml && \
    echo 'lists = true' >> book.toml && \
    echo 'blockquotes = true' >> book.toml && \
    echo 'code-blocks = true' >> book.toml && \
    echo 'tables = true' >> book.toml && \
    echo 'headings = false' >> book.toml && \
    echo '' >> book.toml && \
    echo '[preprocessor.comments.ui]' >> book.toml && \
    echo 'link-text = "comment"' >> book.toml && \
    echo 'show-comment-count = true' >> book.toml && \
    echo '' >> book.toml && \
    echo '[output.html]' >> book.toml

# Create a startup script that configures Neon credentials from env vars
COPY <<'STARTSCRIPT' /usr/local/bin/start-mdbook.sh
#!/bin/bash
set -e

# Check required environment variables
if [ -z "$NEON_API_URL" ]; then
    echo "Error: NEON_API_URL environment variable is required"
    echo "Example: https://ep-xxx-xxx.us-east-2.aws.neon.tech/v2/query"
    exit 1
fi

if [ -z "$NEON_API_KEY" ]; then
    echo "Error: NEON_API_KEY environment variable is required"
    echo "Example: neon_api_key_xxxxxxxxxxxxxxxxxxxxx"
    exit 1
fi

# Configure Neon credentials in the JavaScript file
sed -i "s|apiUrl: 'YOUR_NEON_API_URL'|apiUrl: '${NEON_API_URL}'|g" /book/js/comments-neon.js
sed -i "s|apiKey: 'YOUR_NEON_API_KEY'|apiKey: '${NEON_API_KEY}'|g" /book/js/comments-neon.js

if [ -n "$NEON_DATABASE" ]; then
    sed -i "s|database: 'neondb'|database: '${NEON_DATABASE}'|g" /book/js/comments-neon.js
fi

echo "Starting mdbook serve on http://localhost:3000"
echo "Neon API URL: ${NEON_API_URL}"
echo ""

# Start mdbook serve
exec mdbook serve --hostname 0.0.0.0 --port 3000
STARTSCRIPT

RUN chmod +x /usr/local/bin/start-mdbook.sh

# Expose port 3000
EXPOSE 3000

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/start-mdbook.sh"]
