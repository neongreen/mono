package refactor

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/dissect/pkg/commands"
	"github.com/neongreen/mono/dissect/pkg/goutils"
	"github.com/neongreen/mono/dissect/pkg/qualify"
	"github.com/neongreen/mono/dissect/pkg/references"
	"github.com/neongreen/mono/dissect/pkg/symbols"
	"github.com/neongreen/mono/dissect/pkg/typeinfo"
	"github.com/neongreen/mono/dissect/pkg/utils"
	"go/ast"
	"golang.org/x/tools/go/packages"
)

// detectTargetPackage returns the package name of existing Go files in targetDir.
// Returns empty string if the directory doesn't exist or contains no Go files.
// This is used to detect if a file is being moved into an existing package directory.
func detectTargetPackage(targetDir string) (string, error) {
	// Check if directory exists
	dirInfo, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist yet - will be created during move
			return "", nil
		}
		return "", fmt.Errorf("error checking target directory: %w", err)
	}

	if !dirInfo.IsDir() {
		return "", fmt.Errorf("target path is not a directory: %s", targetDir)
	}

	// Find Go files in the directory
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return "", fmt.Errorf("error reading target directory: %w", err)
	}

	// Look for a Go file (non-test) to determine package
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		// Found a Go file - get its package declaration
		filePath := filepath.Join(targetDir, name)
		pkgName, err := goutils.GetPackageDeclaration(filePath)
		if err != nil {
			slog.Warn("Failed to get package from existing file", "file", filePath, "error", err)
			continue
		}

		slog.Debug("Detected target package from existing file", "file", name, "package", pkgName)
		return pkgName, nil
	}

	// No Go files found - directory is empty or contains only test files
	return "", nil
}

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

	// BEFORE moving: capture exported symbols and find references for later fixing
	// Load the source package to get type information
	sourceDir := filepath.Dir(sourceFile)
	var refsByFile map[string][]references.Reference
	var exportedSymbols []symbols.ExportedSymbol
	sourcePkg, err := typeinfo.LoadPackage(sourceDir)
	if err != nil {
		slog.Warn("Failed to load source package for symbol extraction", "error", err)
		// Continue without symbol fixing - the move will still work for import updates
	} else {
		// Find exported symbols in the file being moved
		// Note: We only track exported symbols because unexported symbols cannot be
		// accessed from another package even after qualifying them
		exportedSymbols, err = symbols.FindExportedSymbols(sourceFile, sourcePkg)
		if err != nil {
			slog.Warn("Failed to find symbols", "error", err)
		} else {
			slog.Info("Found exported symbols to track", "count", len(exportedSymbols))

			// Also find ALL symbols to warn about unexported ones
			allSymbols, err := symbols.FindAllSymbols(sourceFile, sourcePkg)
			if err == nil && len(allSymbols) > len(exportedSymbols) {
				unexportedCount := len(allSymbols) - len(exportedSymbols)
				slog.Warn("File contains unexported symbols that will become inaccessible",
					"count", unexportedCount,
					"hint", "Consider making these symbols exported (capitalize first letter) before moving")
			}

			// Find references to these symbols NOW (before moving breaks the package)
			if len(exportedSymbols) > 0 {
				symbolNames := make([]string, len(exportedSymbols))
				for i, sym := range exportedSymbols {
					symbolNames[i] = sym.Name
				}
				slog.Debug("Looking for references to symbols", "symbols", symbolNames)

				refs, err := references.FindReferences(symbolNames, []*packages.Package{sourcePkg})
				if err != nil {
					slog.Warn("Failed to find references to symbols before moving", "error", err)
				} else {
					slog.Info("Found references before moving", "count", len(refs))
					for _, ref := range refs {
						slog.Debug("Reference found", "symbol", ref.Ident.Name, "file", filepath.Base(ref.File), "qualified", ref.Qualified)
					}

					// Group references by file for later processing
					refsByFile = make(map[string][]references.Reference)
					for _, ref := range refs {
						absRefPath, _ := filepath.Abs(ref.File)
						// Only track references in files OTHER than the one being moved
						absSourceFile, _ := filepath.Abs(sourceFile)
						if absRefPath != absSourceFile {
							refsByFile[absRefPath] = append(refsByFile[absRefPath], ref)
						}
					}
				}
			}
		}
	}

	// Check for unexported dependencies BEFORE moving (if changing packages)
	// Simple check: if the directory changes, assume package changes
	targetDir := filepath.Dir(targetFile)

	absSourceDir, _ := filepath.Abs(sourceDir)
	absTargetDir, _ := filepath.Abs(targetDir)

	packageWillChange := absSourceDir != absTargetDir

	if packageWillChange && sourcePkg != nil {
		// Analyze dependencies
		unexportedDeps, err := analyzeMoveDependencies(sourceFile, sourcePkg)
		if err != nil {
			slog.Warn("Failed to analyze dependencies", "error", err)
		} else if len(unexportedDeps) > 0 {
			// Format relative paths for error message
			relSourceFile, _ := filepath.Rel(moduleRoot, sourceFile)
			relTargetFile, _ := filepath.Rel(moduleRoot, targetFile)
			return formatDependencyError(sourceFile, targetFile, unexportedDeps, relSourceFile, relTargetFile)
		}
	}

	// BEFORE moving: detect if target directory has an existing package
	// and update source file's package declaration if needed
	targetPkgName, err := detectTargetPackage(targetDir)
	if err != nil {
		slog.Warn("Failed to detect target package", "error", err)
	} else if targetPkgName != "" && targetPkgName != originalPackage {
		// Target directory has an existing package different from source
		slog.Info("Updating package declaration before move",
			"file", sourceFile, "from", originalPackage, "to", targetPkgName)
		if err := goutils.UpdatePackageDeclaration(sourceFile, targetPkgName); err != nil {
			return fmt.Errorf("error updating package declaration before move: %w", err)
		}
	}

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

		// Fix unqualified references in the old package using pre-found references
		// We found these BEFORE moving to avoid loading a broken package
		if len(refsByFile) > 0 {
			slog.Info("Fixing unqualified references in source package", "filesWithRefs", len(refsByFile))

			// Get the new package name for qualification
			newPackageName := filepath.Base(filepath.Dir(targetFile))

			// Fix references in each file
			fixedCount := 0
			for filePath, fileRefs := range refsByFile {
				err := qualify.QualifyReferences(filePath, fileRefs, newPackageName, newImportPath, moduleRoot)
				if err != nil {
					slog.Warn("Failed to qualify references in file", "file", filePath, "error", err)
				} else {
					fixedCount++
					slog.Debug("Fixed references in file", "file", filePath, "count", len(fileRefs))
				}
			}

			if fixedCount > 0 {
				slog.Info("Reference fixing complete", "filesFixed", fixedCount)
			}
		}

		// ADDITIONAL FIX: Handle files in the source package directory (like test files)
		// that don't import the moved package but use its symbols directly.
		// These files won't be in refsByFile because they're in the same package,
		// but they need to be updated to import and qualify references after the move.
		sourcePkgDir := filepath.Dir(sourceFile)
		targetPackageName := filepath.Base(filepath.Dir(targetFile))

		slog.Debug("Processing files in source package directory for unqualified symbol usage", "dir", sourcePkgDir)

		sourceDirFixCount := 0
		for _, goFile := range allGoFiles {
			// Only process files in the source package directory
			absGoFile, _ := filepath.Abs(goFile)
			absSourceFile, _ := filepath.Abs(sourceFile)
			absTargetFile, _ := filepath.Abs(targetFile)
			absSourceDir, _ := filepath.Abs(sourcePkgDir)

			// Skip the moved file itself
			if absGoFile == absSourceFile || absGoFile == absTargetFile {
				continue
			}

			// Only process files in the source directory
			goFileDir := filepath.Dir(absGoFile)
			if goFileDir != absSourceDir {
				continue
			}

			// Skip if this file was already processed via refsByFile
			if _, alreadyProcessed := refsByFile[goFile]; alreadyProcessed {
				continue
			}

			// Check if this file uses any of the exported symbols
			// by creating dummy references for all exported symbols
			var dummyRefs []references.Reference
			for _, sym := range exportedSymbols {
				dummyRefs = append(dummyRefs, references.Reference{
					File:      goFile,
					Ident:     &ast.Ident{Name: sym.Name},
					Qualified: false,
				})
			}

			// Try to qualify references - this will only modify the file if symbols are actually used
			err := qualify.QualifyReferences(goFile, dummyRefs, targetPackageName, newImportPath, moduleRoot)
			if err != nil {
				slog.Debug("No unqualified symbols found in source dir file", "file", goFile)
			} else {
				sourceDirFixCount++
				slog.Debug("Fixed unqualified references in source dir file", "file", goFile)
			}
		}

		if sourceDirFixCount > 0 {
			slog.Info("Fixed references in source package directory files", "filesFixed", sourceDirFixCount)
		}
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
