# Agent Guidelines for jj-run

## Build, Test, and Run Commands

**All commands must be run from the mono repository root.**

```bash
# Build
go build ./jj-run

# Test
go test ./jj-run/...

# Run
go run ./jj-run [args...]

# Install (builds and places in $GOPATH/bin)
go install ./jj-run
```

**Important:** Use `go` commands directly. Do not use `mise` for building or running jj-run.
