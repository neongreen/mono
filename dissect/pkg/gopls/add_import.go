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
//   - filePath: Relative or absolute file path to the Go file where the import should be added.
func AddImport(filePath string, importPath string, moduleRoot string) error {
	slog.Debug("Adding import via gopls", "import", importPath, "file", filePath, "moduleRoot", moduleRoot)

	absFilePath := filePath
	if !filepath.IsAbs(filePath) {
		absFilePath = filepath.Join(moduleRoot, filePath)
	}

	// Construct the gopls command
	cmd := exec.Command("gopls", "execute", "-write", "gopls.add_import", fmt.Sprintf(`{"ImportPath": "%s", "URI": "file://%s"}`, importPath, absFilePath))
	cmd.Dir = moduleRoot // Execute gopls in the Go module root

	slog.Debug("Calling gopls", "command", cmd.String(), "directory", cmd.Dir)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing gopls: %w\n%s", err, stderr.String())
	}

	return nil
}
