#!/bin/bash

# Script to create a new presentation from template

if [ -z "$1" ]; then
  echo "Usage: ./scripts/create-presentation.sh <presentation-name>"
  echo "Example: ./scripts/create-presentation.sh my-new-presentation"
  exit 1
fi

PRESENTATION_NAME=$1
PRESENTATION_DIR="presentations/$PRESENTATION_NAME"

if [ -d "$PRESENTATION_DIR" ]; then
  echo "Error: Presentation '$PRESENTATION_NAME' already exists"
  exit 1
fi

echo "Creating presentation: $PRESENTATION_NAME"

# Create directory structure
mkdir -p "$PRESENTATION_DIR/src"

# Create package.json
cat > "$PRESENTATION_DIR/package.json" << EOF
{
  "name": "$PRESENTATION_NAME",
  "version": "1.0.0",
  "description": "Presentation created with diagram-dsl",
  "type": "module",
  "scripts": {
    "generate": "tsx src/presentation.tsx",
    "dev": "tsx src/presentation.tsx"
  },
  "keywords": [
    "presentation",
    "diagram-dsl"
  ],
  "author": "",
  "license": "MIT",
  "dependencies": {
    "diagram-dsl": "workspace:*",
    "react": "^19.2.0",
    "react-dom": "^19.2.0"
  },
  "devDependencies": {
    "@types/node": "^24.6.2",
    "@types/react": "^19.2.0",
    "tsx": "^4.20.6",
    "typescript": "^5.9.3"
  }
}
EOF

# Create tsconfig.json
cat > "$PRESENTATION_DIR/tsconfig.json" << EOF
{
  "extends": "../../diagram-dsl/tsconfig.json",
  "compilerOptions": {
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "include": ["src/**/*"]
}
EOF

# Create README
cat > "$PRESENTATION_DIR/README.md" << EOF
# $PRESENTATION_NAME

Description of your presentation goes here.

## Generating the Presentation

\`\`\`bash
pnpm install
pnpm generate
\`\`\`

This will create an \`output/\` directory containing:
- Individual SVG files for each slide
- An \`index.html\` viewer for browsing the presentation

## Viewing the Presentation

Open \`output/index.html\` in your web browser. Use:
- Arrow keys or space bar to navigate
- Previous/Next buttons for mouse navigation

## Customizing

Edit \`src/presentation.tsx\` to:
- Add more slides
- Change content
- Modify styling
- Add diagrams and arrows

See [diagram-dsl/PRESENTATION_COMPONENTS.md](../../diagram-dsl/PRESENTATION_COMPONENTS.md) for component documentation.
EOF

echo "Creating presentation template..."
# Note: The actual template file will be created separately to avoid escaping issues
# Just create a placeholder
echo "// Template will be added" > "$PRESENTATION_DIR/src/presentation.tsx"

echo ""
echo "✓ Created presentation: $PRESENTATION_NAME"
echo ""
echo "Next steps:"
echo "  cd $PRESENTATION_DIR"
echo "  # Edit src/presentation.tsx with your content"
echo "  pnpm install"
echo "  pnpm generate"
echo ""
echo "Tip: Copy from presentations/llm-context-management/src/presentation-v2.tsx"
echo "     as a starting point for your presentation."
echo ""
