package commands

import (
	"fmt"
)

// RunGoModTidy runs 'go mod tidy' in the specified working directory.
func RunGoModTidy(workingDir string) error {
	_, _, err := RunGoCommand("mod", []string{"tidy"}, workingDir, nil)
	if err != nil {
		return fmt.Errorf("error running 'go mod tidy': %w", err)
	}
	return nil
}
