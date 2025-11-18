# Claude Trace Architecture

## Build, Test, and Run Commands

**All commands must be run from the mono repository root.**

```bash
# Build
go build ./claude-trace

# Test
go test ./claude-trace/...

# Run
go run ./claude-trace [args...]

# Install (builds and places in $GOPATH/bin)
go install ./claude-trace
```

**Important:** Use `go` commands directly. Do not use `mise` for building or running claude-trace.

## Processing Pipeline

The claude-trace tool follows a three-stage pipeline:

1. **Discovery/Loading** (`pkg/storage/`)
   - Find trace files in various locations
   - Load raw file content
   - Handle different file formats (.jsonl, .log, .txt, etc.)

2. **Parsing** (`pkg/parser/`)
   - Parse raw trace content into internal representation
   - Handle JSONL format (one JSON object per line)
   - Extract conversation items: user messages, assistant responses, tool uses, etc.
   - Normalize timestamps, extract metadata
   - **Location:** Find trace parsing logic here, NOT in render package

3. **Rendering** (`pkg/render/`)
   - Take parsed internal representation
   - Render to JSON (structured, not raw)
   - Render to Markdown (human-readable conversation format)
   - Each renderer operates on the same IR

## Why This Architecture?

- **Separation of concerns:** Parsing is independent of output format
- **Reusability:** Same parsed data can be rendered to multiple formats
- **Testability:** Each stage can be tested independently
- **Maintainability:** Changes to output format don't affect parsing logic

## Internal Representation

The parser produces a structured representation (`ParsedTrace`) containing:
- Trace metadata (session ID, timestamps, etc.)
- Ordered list of conversation items
- Each item is typed (UserMessage, AssistantMessage, ToolUse, etc.)
- All fields are properly typed, no raw JSON

## Output Formats

### JSON Output
Structured data reflecting the internal representation, not raw trace content.

### Markdown Output
Human-readable conversation with:
- Clear message boundaries
- User/Assistant labels
- Tool uses formatted nicely
- Collapsible thinking sections
- Timestamps in readable format

