//go:build js || wasm

package rotatinglog

import "os"

// lockFile is a no-op on js/wasm
func lockFile(f *os.File) error {
	return nil
}

// unlockFile is a no-op on js/wasm
func unlockFile(f *os.File) {
}

// isLockError returns false on js/wasm (no file locking)
func isLockError(err error) bool {
	return false
}
