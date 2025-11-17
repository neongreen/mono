//go:build unix

package rotatinglog

import (
	"errors"
	"os"
	"syscall"
)

// lockFile acquires an exclusive lock on the file
func lockFile(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return &lockError{err: err}
	}
	return nil
}

// unlockFile releases the lock on the file
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
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
