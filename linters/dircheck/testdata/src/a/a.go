package a

import (
	"os"
	"path/filepath"
)

// Bad: WriteFile without directory check
func writeConfigBad(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644) // want "file write without directory check"
}

// Good: WriteFile with directory check
func writeConfigGood(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Bad: Create without directory check
func createFileBad(path string) (*os.File, error) {
	return os.Create(path) // want "file write without directory check"
}

// Good: Create with directory check
func createFileGood(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return os.Create(path)
}

// Bad: OpenFile without directory check
func openFileBad(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644) // want "file write without directory check"
}

// Good: OpenFile with directory check
func openFileGood(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
}

// Good: Suppressed with nolint
func writeConfigSuppressed(path string, data []byte) error {
	//nolint:dircheck
	return os.WriteFile(path, data, 0o644)
}

// Good: Suppressed with generic nolint
func writeConfigSuppressedGeneric(path string, data []byte) error {
	//nolint
	return os.WriteFile(path, data, 0o644)
}
