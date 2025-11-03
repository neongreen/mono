package termutil

import (
	"os"
	"strconv"

	"golang.org/x/term"
)

// GetTerminalWidth returns the current terminal width in columns.
// It tries multiple sources in order:
// 1. /dev/tty (works even when stdout is redirected/piped)
// 2. COLUMNS environment variable
// 3. stdout file descriptor
// 4. Default fallback of 80 columns
func GetTerminalWidth() int {
	// Try /dev/tty first (works even when stdout is redirected)
	if tty, err := os.Open("/dev/tty"); err == nil {
		defer tty.Close()
		if width, _, err := term.GetSize(int(tty.Fd())); err == nil {
			return width
		}
	}

	// Try COLUMNS environment variable
	if colsStr := os.Getenv("COLUMNS"); colsStr != "" {
		if cols, err := strconv.Atoi(colsStr); err == nil && cols > 0 {
			return cols
		}
	}

	// Try stdout (works for normal terminal usage)
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return width
	}

	// Default fallback
	return 80
}
