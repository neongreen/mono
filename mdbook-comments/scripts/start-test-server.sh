#!/bin/bash
# Start json-server with test database for Playwright tests

set -e

# Reset to test database
cat > /tmp/test-db.json << 'EOF'
{
  "comments": []
}
EOF

# Start json-server with test database
json-server /tmp/test-db.json \
  --port 54322 \
  --middlewares json-server-middleware.js \
  --routes routes.json \
  --watch
