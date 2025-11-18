# Agent Guidelines for ingest

## Build, Test, and Run Commands

**All commands must be run from the repository root (`/home/user/mono`).**

```bash
# Build
go build ./ingest

# Test
go test ./ingest/...

# Run
go run ./ingest [args...]

# Install (builds and places in $GOPATH/bin)
go install ./ingest
```

**Important:** Use `go` commands directly. Do not use `mise` for building or running ingest.
