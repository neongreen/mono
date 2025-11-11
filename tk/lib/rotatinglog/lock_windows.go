//go:build windows

package rotatinglog

import (
	"errors"
	"os"
	"syscall"
)

// lockFile acquires an exclusive lock on the file (Windows implementation)
func lockFile(f *os.File) error {
	// Get file handle
	handle := syscall.Handle(f.Fd())

	// Lock the entire file (offset 0, length 0xFFFFFFFF)
	overlapped := &syscall.Overlapped{}
	err := syscall.LockFileEx(handle, syscall.LOCKFILE_EXCLUSIVE_LOCK|syscall.LOCKFILE_FAIL_IMMEDIATELY, 0, 0xFFFFFFFF, 0xFFFFFFFF, overlapped)
	if err != nil {
		return &lockError{err: err}
	}
	return nil
}

// unlockFile releases the lock on the file
func unlockFile(f *os.File) error {
	handle := syscall.Handle(f.Fd())
	overlapped := &syscall.Overlapped{}
	return syscall.UnlockFileEx(handle, 0, 0xFFFFFFFF, 0xFFFFFFFF, overlapped)
}

// lockError is returned when a file lock cannot be acquired
type lockError struct {
	err error
}

func (e *lockError) Error() string {
	return "file is locked: " + e.err.Error()
}

// isLockError checks if an error is due to lock failure
func isLockError(err error) bool {
	var le *lockError
	return errors.As(err, &le)
}
