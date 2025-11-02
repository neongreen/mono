//go:build unix

package invlog

import (
	"os"
	"syscall"
)

// lockFile acquires an exclusive lock on the file for safe concurrent writes
func lockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// unlockFile releases the file lock
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
