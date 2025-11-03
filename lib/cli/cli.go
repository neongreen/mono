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
func Success(s string) string {
	return successColor.Sprint(s)
}

// Warning returns a yellow-colored string for warnings.
func Warning(s string) string {
	return warningColor.Sprint(s)
}

// Error returns a red-colored string for errors.
func Error(s string) string {
	return errorColor.Sprint(s)
}

// Key returns a cyan bold-colored string for configuration keys and identifiers.
func Key(s string) string {
	return keyColor.Sprint(s)
}

// Path returns a cyan-colored string for file paths.
func Path(s string) string {
	return pathColor.Sprint(s)
}

// Value returns a green-colored string for configured values.
func Value(v interface{}) string {
	return valueColor.Sprint(v)
}

// Type returns a yellow-colored string for type information.
func Type(s string) string {
	return typeColor.Sprint(s)
}

// Secondary returns a blue-colored string for secondary information.
func Secondary(s string) string {
	return secondaryColor.Sprint(s)
}

// Muted returns a bright-black-colored string for disabled or unset values.
func Muted(s string) string {
	return mutedColor.Sprint(s)
}

// Header returns a bold string for section headers.
func Header(s string) string {
	return headerColor.Sprint(s)
}

// Sprintf variants for formatted strings

// Successf returns a formatted green-colored string.
func Successf(format string, a ...interface{}) string {
	return successColor.Sprintf(format, a...)
}

// Warningf returns a formatted yellow-colored string.
func Warningf(format string, a ...interface{}) string {
	return warningColor.Sprintf(format, a...)
}

// Errorf returns a formatted red-colored string.
func Errorf(format string, a ...interface{}) string {
	return errorColor.Sprintf(format, a...)
}

// Keyf returns a formatted cyan bold-colored string.
func Keyf(format string, a ...interface{}) string {
	return keyColor.Sprintf(format, a...)
}

// Pathf returns a formatted cyan-colored string.
func Pathf(format string, a ...interface{}) string {
	return pathColor.Sprintf(format, a...)
}

// Valuef returns a formatted green-colored string.
func Valuef(format string, a ...interface{}) string {
	return valueColor.Sprintf(format, a...)
}

// Typef returns a formatted yellow-colored string.
func Typef(format string, a ...interface{}) string {
	return typeColor.Sprintf(format, a...)
}

// Secondaryf returns a formatted blue-colored string.
func Secondaryf(format string, a ...interface{}) string {
	return secondaryColor.Sprintf(format, a...)
}

// Mutedf returns a formatted bright-black-colored string.
func Mutedf(format string, a ...interface{}) string {
	return mutedColor.Sprintf(format, a...)
}

// Headerf returns a formatted bold string.
func Headerf(format string, a ...interface{}) string {
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
