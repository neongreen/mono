# Agent Guidelines for markdown-format

## Build, Test, and Run Commands

**All commands must be run from the mono repository root.**

```bash
# Build
go build ./markdown-format

# Test
go test ./markdown-format/...

# Run
go run ./markdown-format [args...]

# Install (builds and places in $GOPATH/bin)
go install ./markdown-format
```

**Important:** Use `go` commands directly. Do not use `mise` for building or running markdown-format.
