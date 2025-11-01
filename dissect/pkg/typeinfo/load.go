package typeinfo

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

// LoadPackage loads a single package with full type information.
// The dir parameter should be the directory containing the package.
func LoadPackage(dir string) (*packages.Package, error) {
	pkgs, err := LoadPackages([]string{"."}, dir)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found in directory %s", dir)
	}
	if len(pkgs) > 1 {
		return nil, fmt.Errorf("expected 1 package, found %d in directory %s", len(pkgs), dir)
	}
	return pkgs[0], nil
}

// LoadPackages loads multiple packages with full type information.
// The patterns parameter specifies which packages to load (e.g., ".", "./...", "path/to/pkg").
// The dir parameter is the working directory for resolving the patterns.
func LoadPackages(patterns []string, dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir: dir,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("error loading packages: %w", err)
	}

	// Check for errors in loaded packages
	var errs []error
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("errors in loaded packages: %v", errs)
	}

	return pkgs, nil
}

