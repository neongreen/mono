package gopls

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/neongreen/mono/dissect/pkg/goutils"
)

func Rename(filePath string, oldName string, newName string, moduleRoot string) error {
	slog.Debug("Renaming function via gopls",
		"filePath", filePath, "oldName", oldName, "newName", newName, "moduleRoot", moduleRoot)

	// Get absolute path for gopls
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("error getting absolute path for %s: %w", filePath, err)
	}

	// Find the position of the old name in the file
	fset, fn, err := goutils.FindFunc(filePath, oldName)
	if err != nil {
		return fmt.Errorf("error finding function %s: %w", oldName, err)
	}
	position := fset.Position(fn.Name.Pos())

	// Construct the gopls command
	cmd := exec.Command("gopls", "rename", "-write", fmt.Sprintf("%s:%d:%d", absFilePath, position.Line, position.Column), newName)
	cmd.Dir = moduleRoot

	slog.Debug("Calling gopls rename", "command", cmd.String(), "directory", cmd.Dir)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing gopls rename: %w\n%s", err, stderr.String())
	}

	return nil
}
