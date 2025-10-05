package gopls

import (
	"bytes"
	"dissect/pkg/goutils"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

// ExtractToNewFile executes the gopls refactor.extract.toNewFile command.
// It extracts a function from the given file and creates a new file for it.
func ExtractToNewFile(filePath string, funcName string, moduleRoot string) (newFilePath string, err error) {
	slog.Debug("Extracting function to new file via gopls",
		"filePath", filePath, "funcName", funcName, "moduleRoot", moduleRoot)

	// Get absolute path for gopls
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path for %s: %w", filePath, err)
	}

	// Find the position of the old name in the file
	fset, fn, err := goutils.FindFunc(filePath, funcName)
	if err != nil {
		return "", fmt.Errorf("error finding function %s: %w", funcName, err)
	}
	position := fset.Position(fn.Name.Pos())

	// Determine the new file path that gopls will create
	newFileName := GuessGoplsExtractedFileName(funcName)
	newFilePath = filepath.Join(filepath.Dir(filePath), newFileName)

	// Check if the new file already exists
	// (TODO: gopls will actually add a suffix if the file exists, but we don't handle that here yet)
	if _, err := os.Stat(newFilePath); err == nil {
		return "", fmt.Errorf("file %s already exists, can't handle this yet", newFilePath)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("error checking if file %s exists: %w", newFilePath, err)
	}

	// Construct the gopls command
	// gopls codeaction -kind=refactor.extract.toNewFile -exec -w file:///path/to/original.go:line:column
	// Use the position of the function's name

	// Construct the URI without the "file://" prefix, let gopls handle it
	fileURI := fmt.Sprintf("%s:%d:%d", absFilePath, position.Line, position.Column)
	slog.Debug("Calling gopls", "fileURI", fileURI, "moduleRoot", moduleRoot)
	cmd := exec.Command("gopls", "codeaction", "-kind=refactor.extract.toNewFile", "-exec", "-write", fileURI)
	cmd.Dir = moduleRoot // Execute gopls in the Go module root

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error executing gopls: %w\n%s", err, stderr.String())
	}

	// Check if the new file was created
	if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("gopls did not create the new file %s", newFilePath)
	} else if err != nil {
		return "", fmt.Errorf("error checking if new file %s exists: %w", newFilePath, err)
	}

	return newFilePath, nil
}
