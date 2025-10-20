# Web Viewer

This directory contains the React-based web viewer for claude-trace.

## Building

To build the web viewer:

```bash
cd web
npm install
npm run build
```

This creates a single HTML file at `dist/index.html` with all CSS and JavaScript inlined.

## Embedding in Go

After building, copy the generated HTML file to the Go package:

```bash
cp dist/index.html ../pkg/viewer/index.html
```

The HTML file is embedded into the Go binary using `go:embed` in `pkg/viewer/embed.go`.

## Development

To run the viewer in development mode:

```bash
npm run dev
```

This will start a development server at `http://localhost:5173`.

Note: In development mode, you'll need to mock the API endpoint at `/api/trace` or proxy it to a running claude-trace server.
