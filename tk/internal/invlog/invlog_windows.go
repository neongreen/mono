//go:build windows

package invlog

import (
	"os"
)

// lockFile is a no-op on Windows
// File locking on Windows is not implemented to keep the code simple
// Concurrent writes may result in interleaved log entries
func lockFile(f *os.File) error {
	return nil
}

// unlockFile is a no-op on Windows
func unlockFile(f *os.File) error {
	return nil
}
