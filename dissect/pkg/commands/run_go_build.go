package commands

import (
	"fmt"
)

// RunGoBuild runs 'go build -o /dev/null ./...' in the specified working directory.
func RunGoBuild(workingDir string) error {
	_, _, err := RunGoCommand("build", []string{"-o", "/dev/null", "./..."}, workingDir, nil)
	if err != nil {
		return fmt.Errorf("error running 'go build': %w", err)
	}
	return nil
}
