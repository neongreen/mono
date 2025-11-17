//go:build js && wasm

package database

import (
	_ "github.com/ncruces/go-sqlite3/driver" // WASM-compatible SQLite
	_ "github.com/ncruces/go-sqlite3/embed"  // Embed SQLite WASM binary
)

const driverName = "sqlite3"
