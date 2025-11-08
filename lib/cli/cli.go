// Package cli provides consistent color formatting for CLI output across all tools in the monorepo.
package cli

import (
	"fmt"

	"github.com/fatih/color"
)

// Color formatters for consistent semantic styling.
var (
	// Success formats text in green for successful operations and positive states
	successColor = color.New(color.FgGreen)

	// Warning formats text in yellow for warnings and in-progress states
	warningColor = color.New(color.FgYellow)

	// Error formats text in red for errors and failures
	errorColor = color.New(color.FgRed)

	// Key formats text in cyan for keys, paths, and primary identifiers
	keyColor = color.New(color.FgCyan, color.Bold)

	// Path formats text in cyan for file paths
	pathColor = color.New(color.FgCyan)

	// Value formats text in green for set/configured values
	valueColor = color.New(color.FgGreen)

	// Type formats text in yellow for type information
	typeColor = color.New(color.FgYellow)

	// Secondary formats text in blue for secondary information
	secondaryColor = color.New(color.FgBlue)

	// Muted formats text in bright black for disabled/unset values
	mutedColor = color.New(color.FgHiBlack)

	// Header formats text as bold for section headers
	headerColor = color.New(color.Bold)
)

// Success returns a green-colored string for successful operations.
//
//nolint:uselesswrapper // Provides semantic naming for success states
func Success(s string) string {
	return successColor.Sprint(s)
}

// Warning returns a yellow-colored string for warnings.
//
//nolint:uselesswrapper // Provides semantic naming for warning states
func Warning(s string) string {
	return warningColor.Sprint(s)
}

// Error returns a red-colored string for errors.
//
//nolint:uselesswrapper // Provides semantic naming for error states
func Error(s string) string {
	return errorColor.Sprint(s)
}

// Key returns a cyan bold-colored string for configuration keys and identifiers.
//
//nolint:uselesswrapper // Provides semantic naming for keys/identifiers
func Key(s string) string {
	return keyColor.Sprint(s)
}

// Path returns a cyan-colored string for file paths.
//
//nolint:uselesswrapper // Provides semantic naming for paths
func Path(s string) string {
	return pathColor.Sprint(s)
}

// Value returns a green-colored string for configured values.
//
//nolint:uselesswrapper // Provides semantic naming for values
func Value(v any) string {
	return valueColor.Sprint(v)
}

// Type returns a yellow-colored string for type information.
//
//nolint:uselesswrapper // Provides semantic naming for types
func Type(s string) string {
	return typeColor.Sprint(s)
}

// Secondary returns a blue-colored string for secondary information.
//
//nolint:uselesswrapper // Provides semantic naming for secondary info
func Secondary(s string) string {
	return secondaryColor.Sprint(s)
}

// Muted returns a bright-black-colored string for disabled or unset values.
//
//nolint:uselesswrapper // Provides semantic naming for muted/disabled states
func Muted(s string) string {
	return mutedColor.Sprint(s)
}

// Header returns a bold string for section headers.
//
//nolint:uselesswrapper // Provides semantic naming for headers
func Header(s string) string {
	return headerColor.Sprint(s)
}

// Sprintf variants for formatted strings

// Successf returns a formatted green-colored string.
//
//nolint:uselesswrapper // Provides semantic naming for success states
func Successf(format string, a ...any) string {
	return successColor.Sprintf(format, a...)
}

// Warningf returns a formatted yellow-colored string.
//
//nolint:uselesswrapper // Provides semantic naming for warning states
func Warningf(format string, a ...any) string {
	return warningColor.Sprintf(format, a...)
}

// Errorf returns a formatted red-colored string.
//
//nolint:uselesswrapper // Provides semantic naming for error states
func Errorf(format string, a ...any) string {
	return errorColor.Sprintf(format, a...)
}

// Keyf returns a formatted cyan bold-colored string.
//
//nolint:uselesswrapper // Provides semantic naming for keys/identifiers
func Keyf(format string, a ...any) string {
	return keyColor.Sprintf(format, a...)
}

// Pathf returns a formatted cyan-colored string.
//
//nolint:uselesswrapper // Provides semantic naming for paths
func Pathf(format string, a ...any) string {
	return pathColor.Sprintf(format, a...)
}

// Valuef returns a formatted green-colored string.
//
//nolint:uselesswrapper // Provides semantic naming for values
func Valuef(format string, a ...any) string {
	return valueColor.Sprintf(format, a...)
}

// Typef returns a formatted yellow-colored string.
//
//nolint:uselesswrapper // Provides semantic naming for types
func Typef(format string, a ...any) string {
	return typeColor.Sprintf(format, a...)
}

// Secondaryf returns a formatted blue-colored string.
//
//nolint:uselesswrapper // Provides semantic naming for secondary info
func Secondaryf(format string, a ...any) string {
	return secondaryColor.Sprintf(format, a...)
}

// Mutedf returns a formatted bright-black-colored string.
//
//nolint:uselesswrapper // Provides semantic naming for muted/disabled states
func Mutedf(format string, a ...any) string {
	return mutedColor.Sprintf(format, a...)
}

// Headerf returns a formatted bold string.
//
//nolint:uselesswrapper // Provides semantic naming for headers
func Headerf(format string, a ...any) string {
	return headerColor.Sprintf(format, a...)
}

// PrintSuccess prints a green-colored message to stdout.
func PrintSuccess(s string) {
	fmt.Println(Success(s))
}

// PrintWarning prints a yellow-colored message to stdout.
func PrintWarning(s string) {
	fmt.Println(Warning(s))
}

// PrintError prints a red-colored message to stdout.
func PrintError(s string) {
	fmt.Println(Error(s))
}

// PrintHeader prints a bold header to stdout.
func PrintHeader(s string) {
	fmt.Println(Header(s))
}
