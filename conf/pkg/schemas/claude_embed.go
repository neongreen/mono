package schemas

import _ "embed"

// ClaudeSchema contains the embedded Claude Code configuration schema
//
//go:embed claude.json
var ClaudeSchema string
