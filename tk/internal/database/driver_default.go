//go:build !js && !wasm

package database

import (
	_ "modernc.org/sqlite" // Pure Go SQLite for non-WASM builds
)

const driverName = "sqlite"
