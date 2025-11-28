package gopls

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
)

// AddImport executes the `gopls.add_import` command.
// It adds an import to the given file.
//
// Parameters:
//   - goplsPath: Path to the gopls binary
//   - filePath: Relative or absolute file path to the Go file where the import should be added.
func AddImport(goplsPath string, filePath string, importPath string, moduleRoot string) error {
	slog.Debug("Adding import via gopls", "import", importPath, "file", filePath, "moduleRoot", moduleRoot)

	absFilePath := filePath
	if !filepath.IsAbs(filePath) {
		absFilePath = filepath.Join(moduleRoot, filePath)
	}

	// Construct the gopls command
	cmd := exec.Command(goplsPath, "execute", "-write", "gopls.add_import", fmt.Sprintf(`{"ImportPath": "%s", "URI": "file://%s"}`, importPath, absFilePath)) //nolint:gosec // G204: intentional gopls execution with validated path
	cmd.Dir = moduleRoot                                                                                                                                      // Execute gopls in the Go module root

	slog.Debug("Calling gopls", "command", cmd.String(), "directory", cmd.Dir, "goplsPath", goplsPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if _, err := runWithTextFileBusyRetry(cmd); err != nil {
		return fmt.Errorf("error executing gopls: %w\n%s", err, stderr.String())
	}

	return nil
}
