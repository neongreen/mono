//go:build tools
// +build tools

package tools

// This file declares dependencies that are used for development and testing
// but not in the actual application code. By importing them here with blank
// imports, we ensure they are tracked in go.mod and available for go install.

import (
	_ "gotest.tools/gotestsum"
)
