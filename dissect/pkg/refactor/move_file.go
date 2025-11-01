package refactor

import (
	"fmt"
	"github.com/neongreen/mono/dissect/pkg/commands"
	"github.com/neongreen/mono/dissect/pkg/goutils"
	"github.com/neongreen/mono/dissect/pkg/utils"
	"go/ast"
	"log/slog"
	"path/filepath"
)

// MoveFileWithImportUpdates performs a refactoring file move from source to target,
// updating all import statements in the codebase that reference the old package path.
func MoveFileWithImportUpdates(sourceFile string, targetFile string, moduleRoot string) error {
	slog.Info("Starting refactoring file move", "from", sourceFile, "to", targetFile)

	// Get the original package name from the source file
	originalPackage, err := goutils.GetPackageDeclaration(sourceFile)
	if err != nil {
		return fmt.Errorf("error getting source package declaration: %w", err)
	}
	slog.Debug("Source package", "package", originalPackage)

	// Get the old import path before moving the file
	oldImportPath, err := commands.GetFullImportPath(sourceFile)
	if err != nil {
		return fmt.Errorf("error getting source import path: %w", err)
	}
	slog.Debug("Source import path", "path", oldImportPath)

	// Find all Go files BEFORE moving (while codebase is still valid)
	// This is critical because after moving, the codebase may be in a broken state
	// with stale imports, which prevents go list from working
	allGoFiles, err := commands.FindGoFiles(moduleRoot)
	if err != nil {
		return fmt.Errorf("error finding Go files: %w", err)
	}
	slog.Debug("Found Go files", "count", len(allGoFiles))

	// Move the physical file
	slog.Debug("Moving file", "from", sourceFile, "to", targetFile)
	if err := utils.MoveFile(sourceFile, targetFile); err != nil {
		return fmt.Errorf("error moving file: %w", err)
	}

	// Get the new import path after moving
	newImportPath, err := commands.GetFullImportPath(targetFile)
	if err != nil {
		return fmt.Errorf("error getting target import path: %w", err)
	}
	slog.Debug("Target import path", "path", newImportPath)

	// Determine if we need to update the package name
	// The package name should only change if we're moving to a different package directory
	packageChanged := oldImportPath != newImportPath

	if packageChanged {
		// The import path changed, so we need to determine the new package name
		// Get the package name from the target directory
		targetDir := filepath.Dir(targetFile)
		newPackageName := filepath.Base(targetDir)

		slog.Debug("Updating package declaration", "file", targetFile, "oldPackage", originalPackage, "newPackage", newPackageName)
		if err := goutils.UpdatePackageDeclaration(targetFile, newPackageName); err != nil {
			return fmt.Errorf("error updating package declaration: %w", err)
		}

		slog.Info("Updating imports in all files", "oldPath", oldImportPath, "newPath", newImportPath)

		// Update imports in all files that reference the old package
		updatedCount := 0
		for _, goFile := range allGoFiles {
			// Skip the moved file itself (at both old and new locations)
			absGoFile, _ := filepath.Abs(goFile)
			absSourceFile, _ := filepath.Abs(sourceFile)
			absTargetFile, _ := filepath.Abs(targetFile)
			if absGoFile == absSourceFile || absGoFile == absTargetFile {
				continue
			}

			updated, err := updateImportInFile(goFile, oldImportPath, newImportPath)
			if err != nil {
				slog.Warn("Error updating imports in file", "file", goFile, "error", err)
				continue
			}
			if updated {
				updatedCount++
				slog.Debug("Updated imports", "file", goFile)
			}
		}

		slog.Info("Import update complete", "filesUpdated", updatedCount)
	} else {
		slog.Info("Import path unchanged, no import or package updates needed")
	}

	return nil
}

// updateImportInFile updates the import statement in a file from oldPath to newPath using AST operations.
// It also updates all package qualifiers (selectors) that reference the old package.
// Returns true if the file was modified, false otherwise.
func updateImportInFile(filePath string, oldImportPath string, newImportPath string) (bool, error) {
	fset, node, err := goutils.ReadGoFile(filePath)
	if err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	modified := false

	// Determine the old and new package names from the import paths
	// The package name is typically the last component of the import path
	oldPackageName := getPackageNameFromPath(oldImportPath)
	newPackageName := getPackageNameFromPath(newImportPath)

	var actualOldPackageName string // Track what name is actually used in the import

	// Iterate through all import declarations to update import paths
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		// Check if this is an import declaration
		if genDecl.Tok.String() != "import" {
			continue
		}

		// Check all import specs
		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}

			// Get the import path (remove quotes)
			importPath := importSpec.Path.Value
			if len(importPath) >= 2 && importPath[0] == '"' && importPath[len(importPath)-1] == '"' {
				importPath = importPath[1 : len(importPath)-1]
			}

			// Check if this import matches the old path
			if importPath == oldImportPath {
				// Update to new path (add quotes back)
				importSpec.Path.Value = fmt.Sprintf(`"%s"`, newImportPath)

				// Determine what package name is actually being used
				hasAlias := importSpec.Name != nil
				if hasAlias {
					// There's an explicit alias - selectors use the alias name, which doesn't change
					actualOldPackageName = "" // Empty means don't update selectors
					slog.Debug("Updated import with alias in file", "file", filePath, "from", oldImportPath, "to", newImportPath, "alias", importSpec.Name.Name)
				} else {
					// Use the default package name (last component of path)
					actualOldPackageName = oldPackageName
					slog.Debug("Updated import in file", "file", filePath, "from", oldImportPath, "to", newImportPath, "oldPkg", oldPackageName, "newPkg", newPackageName)
				}

				modified = true
			}
		}
	}

	// If we updated an import and the package names differ, we need to update all selectors
	if modified && actualOldPackageName != "" && actualOldPackageName != newPackageName {
		slog.Debug("Updating selectors", "file", filePath, "from", actualOldPackageName, "to", newPackageName)

		// Walk the AST to find and update all selector expressions
		ast.Inspect(node, func(n ast.Node) bool {
			if sel, ok := n.(*ast.SelectorExpr); ok {
				// Check if this is a selector on the old package name
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == actualOldPackageName {
						// Update to use the new package name
						ident.Name = newPackageName
						slog.Debug("Updated selector", "file", filePath, "from", actualOldPackageName, "to", newPackageName, "selector", sel.Sel.Name)
					}
				}
			}
			return true
		})
	}

	// If modified, write the file back
	if modified {
		if err := goutils.WriteGoFile(filePath, fset, node); err != nil {
			return false, fmt.Errorf("error writing file: %w", err)
		}

		// Run goimports to clean up
		if err := commands.RunGoimportsOnFile(filePath); err != nil {
			return false, fmt.Errorf("error running goimports: %w", err)
		}
	}

	return modified, nil
}

// getPackageNameFromPath extracts the package name from an import path.
// For example, "github.com/user/repo/pkg/helper" returns "helper"
func getPackageNameFromPath(importPath string) string {
	parts := filepath.SplitList(importPath)
	if len(parts) > 0 {
		// Return the last non-empty part
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" {
				return filepath.Base(parts[i])
			}
		}
	}
	// Fallback to using filepath.Base if SplitList doesn't work as expected
	return filepath.Base(importPath)
}
