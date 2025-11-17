package refactor

import (
	"fmt"
	"go/ast"
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
func MoveFileWithImportUpdates(sourceFile string, targetFile string, moduleRoot string, goimportsPath string) error {
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
		// Check 1: Analyze if the file BEING MOVED depends on unexported symbols from the source package
		unexportedDeps, err := analyzeMoveDependencies(sourceFile, sourcePkg)
		if err != nil {
			slog.Warn("Failed to analyze dependencies", "error", err)
		} else if len(unexportedDeps) > 0 {
			// Format relative paths for error message
			relSourceFile, _ := filepath.Rel(moduleRoot, sourceFile)
			relTargetFile, _ := filepath.Rel(moduleRoot, targetFile)
			return formatDependencyError(sourceFile, targetFile, unexportedDeps, relSourceFile, relTargetFile)
		}

		// Check 2: Check if OTHER files depend on unexported symbols FROM the file being moved
		// Find ALL symbols (exported and unexported) in the file being moved
		allSymbols, err := symbols.FindAllSymbols(sourceFile, sourcePkg)
		if err != nil {
			slog.Warn("Failed to find symbols for unexported check", "error", err)
		} else if len(allSymbols) > 0 {
			// Get all symbol names
			symbolNames := make([]string, len(allSymbols))
			for i, sym := range allSymbols {
				symbolNames[i] = sym.Name
			}

			// Find ALL references to these symbols
			allRefs, err := references.FindReferences(symbolNames, []*packages.Package{sourcePkg})
			if err != nil {
				slog.Warn("Failed to find references for unexported check", "error", err)
			} else {
				// Check if any unexported symbols are referenced from files NOT being moved
				ops := []MoveOp{{From: sourceFile, To: targetFile}}
				unexportedRefs := detectUnexportedExternalRefs(allRefs, ops, []*packages.Package{sourcePkg}, sourcePkg.PkgPath)
				if len(unexportedRefs) > 0 {
					return buildUnexportedSymbolError(unexportedRefs)
				}
			}
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

			updated, err := updateImportInFile(goFile, oldImportPath, newImportPath, goimportsPath)
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
func updateImportInFile(filePath string, oldImportPath string, newImportPath string, goimportsPath string) (bool, error) {
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
		if err := commands.RunGoimportsOnFile(goimportsPath, filePath); err != nil {
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

// MoveOp represents a single file move operation
type MoveOp struct {
	From string // Absolute source path
	To   string // Absolute destination path
}

// MoveBatchFiles performs atomic batch file moves with refactoring support.
// All files are moved together and import updates happen once for all files.
func MoveBatchFiles(moveOps []MoveOp, moduleRoot string, goimportsPath string) error {
	if len(moveOps) == 0 {
		return fmt.Errorf("no files to move")
	}

	slog.Info("Starting batch file move", "count", len(moveOps))

	// Phase 1: Validation
	if err := validateBatchMoves(moveOps); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Phase 2: Collect exported symbols and load source packages BEFORE moving
	// We need this to update references in the old package after the move
	exportedSymbolsByFile := make(map[string][]string)   // source file -> exported symbol names
	sourcePackages := make(map[string]*packages.Package) // source dir -> loaded package

	for _, op := range moveOps {
		// Collect exported symbols
		syms, err := symbols.ExtractExportedSymbols(op.From)
		if err != nil {
			slog.Warn("Failed to extract exported symbols", "file", op.From, "error", err)
			continue
		}
		if len(syms) > 0 {
			exportedSymbolsByFile[op.From] = syms
			slog.Debug("Collected exported symbols", "file", op.From, "count", len(syms))
		}

		// Load source package (once per directory)
		sourceDir := filepath.Dir(op.From)
		if _, ok := sourcePackages[sourceDir]; !ok {
			pkg, err := typeinfo.LoadPackage(sourceDir)
			if err != nil {
				slog.Warn("Failed to load source package", "dir", sourceDir, "error", err)
			} else {
				sourcePackages[sourceDir] = pkg
				slog.Debug("Loaded source package", "dir", sourceDir)
			}
		}
	}

	// Phase 3: Determine target packages BEFORE moving
	// This prevents detecting the moved files themselves
	targetPackages := make(map[string]string) // target file -> package name
	for _, op := range moveOps {
		targetDir := filepath.Dir(op.To)

		// Determine target package name
		targetPackage, err := detectTargetPackage(targetDir)
		if err != nil {
			return fmt.Errorf("failed to detect target package for %s: %w", targetDir, err)
		}

		// If no existing package, infer from directory name
		if targetPackage == "" {
			targetPackage = filepath.Base(targetDir)
			slog.Debug("Inferred package name from directory", "package", targetPackage, "dir", targetDir)
		}

		targetPackages[op.To] = targetPackage
	}

	// Phase 3.5: Check for unexported symbols BEFORE moving
	// Load all packages in module to check for cross-package references
	slog.Info("Checking for unexported symbol references", "moduleRoot", moduleRoot)
	allPkgs, err := typeinfo.LoadPackages([]string{"./..."}, moduleRoot)
	if err != nil {
		return fmt.Errorf("failed to load packages for unexported symbol check: %w", err)
	}

	// Group moves by source directory and check each group
	movesBySourceDir := make(map[string][]MoveOp)
	for _, op := range moveOps {
		sourceDir := filepath.Dir(op.From)
		movesBySourceDir[sourceDir] = append(movesBySourceDir[sourceDir], op)
	}

	for sourceDir, ops := range movesBySourceDir {
		// Collect ALL symbols (exported and unexported) from files in this group
		allSymbols := []string{}
		for _, op := range ops {
			syms, err := symbols.ExtractAllSymbols(op.From)
			if err != nil {
				slog.Warn("Failed to extract symbols for unexported check", "file", op.From, "error", err)
				continue
			}
			allSymbols = append(allSymbols, syms...)
		}

		if len(allSymbols) == 0 {
			continue // No symbols to check
		}

		// Get the source package
		pkg, ok := sourcePackages[sourceDir]
		if !ok {
			slog.Warn("Source package not loaded, skipping unexported symbol check", "dir", sourceDir)
			continue
		}

		// Find all references to these symbols across the module
		refs, err := references.FindReferences(allSymbols, allPkgs)
		if err != nil {
			slog.Warn("Failed to find references for unexported check", "symbols", allSymbols, "error", err)
			continue
		}

		// Check for unexported symbols referenced from other packages
		unexportedRefs := detectUnexportedExternalRefs(refs, ops, allPkgs, pkg.PkgPath)
		if len(unexportedRefs) > 0 {
			return buildUnexportedSymbolError(unexportedRefs)
		}
	}

	// Phase 4: Physical moves
	movedFiles := make(map[string]string) // from -> to mapping for rollback
	for _, op := range moveOps {
		if err := performPhysicalMove(op.From, op.To); err != nil {
			// Rollback: undo all moves
			slog.Error("Physical move failed, rolling back", "error", err, "from", op.From, "to", op.To)
			rollbackMoves(movedFiles)
			return fmt.Errorf("failed to move %s to %s: %w", op.From, op.To, err)
		}
		movedFiles[op.From] = op.To
		slog.Debug("Physically moved file", "from", op.From, "to", op.To)
	}

	// Phase 5: Update package declarations
	for _, op := range moveOps {
		targetPackage := targetPackages[op.To]

		// Update package declaration
		if err := goutils.UpdatePackageDeclaration(op.To, targetPackage); err != nil {
			slog.Error("Failed to update package declaration, rolling back", "error", err, "file", op.To)
			rollbackMoves(movedFiles)
			return fmt.Errorf("failed to update package in %s: %w", op.To, err)
		}
		slog.Debug("Updated package declaration", "file", op.To, "package", targetPackage)
	}

	// Phase 6: Update references to moved symbols in the old package
	// Use the pre-loaded source packages and all packages
	if err := updateReferencesToMovedSymbols(moveOps, exportedSymbolsByFile, sourcePackages, allPkgs, moduleRoot); err != nil {
		slog.Error("Failed to update references to moved symbols", "error", err)
		// Don't rollback - files are moved and may be partially fixed
		return fmt.Errorf("failed to update symbol references: %w", err)
	}

	// Phase 7: Update imports across codebase (once for all moves)
	// Collect all import path mappings
	importMappings := make(map[string]string) // oldPath -> newPath
	for _, op := range moveOps {
		oldPath, err := commands.GetFullImportPath(op.From)
		if err != nil {
			// File was moved, try to get import path from moved location
			oldPath, err = getImportPathFromMovedFile(op.From, op.To, moduleRoot)
			if err != nil {
				slog.Warn("Could not determine old import path", "from", op.From, "error", err)
				continue
			}
		}

		newPath, err := commands.GetFullImportPath(op.To)
		if err != nil {
			slog.Warn("Could not determine new import path", "to", op.To, "error", err)
			continue
		}

		if oldPath != newPath {
			importMappings[oldPath] = newPath
			slog.Debug("Import path mapping", "old", oldPath, "new", newPath)
		}
	}

	// Update all imports in the module
	if len(importMappings) > 0 {
		if err := updateImportsForBatch(moduleRoot, importMappings, goimportsPath); err != nil {
			slog.Error("Failed to update imports", "error", err)
			// Don't rollback here - files are already moved and packages updated
			// Imports can be fixed manually if needed
			return fmt.Errorf("failed to update imports: %w", err)
		}
	}

	slog.Info("Batch file move completed successfully", "count", len(moveOps))
	return nil
}

// validateBatchMoves checks that all move operations are valid
func validateBatchMoves(moveOps []MoveOp) error {
	// Check for duplicate sources
	seen := make(map[string]bool)
	for _, op := range moveOps {
		if seen[op.From] {
			return fmt.Errorf("duplicate source file: %s", op.From)
		}
		seen[op.From] = true

		// Check source exists
		if _, err := os.Stat(op.From); os.IsNotExist(err) {
			return fmt.Errorf("source file does not exist: %s", op.From)
		}

		// Check source and target are not the same
		if op.From == op.To {
			return fmt.Errorf("source and target are the same: %s", op.From)
		}
	}

	// Check for target conflicts
	targets := make(map[string]string) // target -> source
	for _, op := range moveOps {
		if existing, exists := targets[op.To]; exists {
			return fmt.Errorf("multiple files would be moved to %s (from %s and %s)", op.To, existing, op.From)
		}
		targets[op.To] = op.From

		// Check if target already exists (unless it's being moved by this batch)
		if _, err := os.Stat(op.To); err == nil {
			// Target exists - check if it's being moved by this batch
			isBeingMoved := false
			for _, otherOp := range moveOps {
				if otherOp.From == op.To {
					isBeingMoved = true
					break
				}
			}
			if !isBeingMoved {
				return fmt.Errorf("target file already exists: %s", op.To)
			}
		}
	}

	return nil
}

// performPhysicalMove moves a file from source to target, creating directories as needed
func performPhysicalMove(from, to string) error {
	// Create target directory if needed
	targetDir := filepath.Dir(to)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Move the file
	if err := utils.MoveFile(from, to); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// rollbackMoves undoes all file moves
func rollbackMoves(movedFiles map[string]string) {
	slog.Warn("Rolling back file moves", "count", len(movedFiles))
	for from, to := range movedFiles {
		if err := utils.MoveFile(to, from); err != nil {
			slog.Error("Failed to rollback move", "from", to, "to", from, "error", err)
		} else {
			slog.Debug("Rolled back move", "from", to, "to", from)
		}
	}
}

// getImportPathFromMovedFile reconstructs the old import path from a moved file
func getImportPathFromMovedFile(oldPath, newPath, moduleRoot string) (string, error) {
	// Get relative path from module root to old location
	relPath, err := filepath.Rel(moduleRoot, filepath.Dir(oldPath))
	if err != nil {
		return "", err
	}

	// Get module name
	moduleName, err := commands.GetModuleName(moduleRoot)
	if err != nil {
		return "", err
	}

	return filepath.Join(moduleName, filepath.ToSlash(relPath)), nil
}

// updateImportsForBatch updates all imports in the module based on path mappings
func updateImportsForBatch(moduleRoot string, mappings map[string]string, goimportsPath string) error {
	slog.Info("Updating imports across codebase", "mappings", len(mappings))

	// Find all Go files
	allGoFiles, err := commands.FindGoFiles(moduleRoot)
	if err != nil {
		return fmt.Errorf("failed to find Go files: %w", err)
	}

	// Update imports in each file
	for _, filePath := range allGoFiles {
		modified := false
		for oldPath, newPath := range mappings {
			fileModified, err := updateImportInFile(filePath, oldPath, newPath, goimportsPath)
			if err != nil {
				return fmt.Errorf("failed to update imports in %s: %w", filePath, err)
			}
			if fileModified {
				modified = true
			}
		}

		// Run goimports if file was modified
		if modified {
			if err := commands.RunGoimportsOnFile(goimportsPath, filePath); err != nil {
				slog.Warn("Failed to run goimports", "file", filePath, "error", err)
			}
		}
	}

	return nil
}

// updateReferencesToMovedSymbols finds all references to moved symbols in the original package
// and updates them to use qualified references with the new package.
func updateReferencesToMovedSymbols(moveOps []MoveOp, exportedSymbolsByFile map[string][]string, sourcePackages map[string]*packages.Package, allPackages []*packages.Package, moduleRoot string) error {
	// Group moves by source directory (original package)
	movesBySourceDir := make(map[string][]MoveOp) // source dir -> move ops
	for _, op := range moveOps {
		sourceDir := filepath.Dir(op.From)
		movesBySourceDir[sourceDir] = append(movesBySourceDir[sourceDir], op)
	}

	// Process each source package
	for sourceDir, ops := range movesBySourceDir {
		// Check if all files are moving within the same directory (package)
		// If so, no reference qualification is needed
		allSameDir := true
		for _, op := range ops {
			if filepath.Dir(op.To) != sourceDir {
				allSameDir = false
				break
			}
		}
		if allSameDir {
			slog.Debug("All moves within same package, skipping reference updates", "dir", sourceDir)
			continue
		}

		// Collect all exported symbols from this group
		var allSymbols []string
		for _, op := range ops {
			if syms, ok := exportedSymbolsByFile[op.From]; ok {
				allSymbols = append(allSymbols, syms...)
			}
		}

		if len(allSymbols) == 0 {
			continue // No exported symbols to process
		}

		// Check that the source package was loaded
		if _, ok := sourcePackages[sourceDir]; !ok {
			slog.Warn("Source package not loaded, skipping reference updates", "dir", sourceDir)
			continue
		}

		// Find all references to the moved symbols across the entire module
		slog.Info("Looking for references to moved symbols", "symbols", allSymbols, "sourceDir", sourceDir)
		refs, err := references.FindReferences(allSymbols, allPackages)
		if err != nil {
			slog.Warn("Failed to find references", "symbols", allSymbols, "error", err)
			continue
		}

		slog.Info("Found references to moved symbols", "count", len(refs))
		if len(refs) == 0 {
			slog.Info("No references to update for moved symbols")
			continue // No references to update
		}

		// Determine the target package and import path from the first move op
		// (all ops in this group should have the same target package)
		firstOp := ops[0]
		targetPackageName := filepath.Base(filepath.Dir(firstOp.To))

		// Get the full import path for the target
		targetImportPath, err := commands.GetFullImportPath(firstOp.To)
		if err != nil {
			slog.Warn("Failed to get target import path", "file", firstOp.To, "error", err)
			continue
		}

		// Group references by file
		refsByFile := make(map[string][]references.Reference)
		for _, ref := range refs {
			refsByFile[ref.File] = append(refsByFile[ref.File], ref)
		}

		// Update each file that has references
		for filename, fileRefs := range refsByFile {
			// Skip the moved files themselves (check both old and new paths)
			isMovedFile := false
			for _, op := range ops {
				if filename == op.From || filename == op.To {
					isMovedFile = true
					break
				}
			}
			if isMovedFile {
				slog.Debug("Skipping moved file itself", "file", filename)
				continue
			}

			slog.Debug("Qualifying references in file", "file", filename, "count", len(fileRefs))

			// Qualify the references
			if err := qualify.QualifyReferences(filename, fileRefs, targetPackageName, targetImportPath, moduleRoot); err != nil {
				return fmt.Errorf("failed to qualify references in %s: %w", filename, err)
			}
		}
	}

	return nil
}

// UnexportedRef represents an unexported symbol being referenced from another package
type UnexportedRef struct {
	SymbolName     string
	FileName       string   // File containing the symbol
	ReferencedFrom []string // Files referencing it
}

// detectUnexportedExternalRefs detects unexported symbols that are referenced from other packages.
// These would cause build failures after the move.
func detectUnexportedExternalRefs(refs []references.Reference, ops []MoveOp, allPkgs []*packages.Package, sourcePkgPath string) []UnexportedRef {
	// Build a map of files being moved
	movedFiles := make(map[string]bool)
	for _, op := range ops {
		movedFiles[op.From] = true
	}

	// Track unexported symbols referenced from other packages
	unexportedMap := make(map[string]*UnexportedRef)

	for _, ref := range refs {
		// Check if the symbol is unexported (starts with lowercase or is a field)
		isUnexported := false
		symbolName := ref.Ident.Name

		// Check if it's a selector expression (field/method access)
		if ref.Selector != "" {
			// It's a field or method access - check the field/method name (after the last dot)
			parts := strings.Split(ref.Selector, ".")
			if len(parts) > 1 {
				fieldName := parts[len(parts)-1]
				if len(fieldName) > 0 && fieldName[0] >= 'a' && fieldName[0] <= 'z' {
					isUnexported = true
					symbolName = ref.Selector // Use the full selector path (e.g., "DB.data")
				}
			}
		} else {
			// It's a simple identifier (not a field access)
			if len(symbolName) > 0 && symbolName[0] >= 'a' && symbolName[0] <= 'z' {
				isUnexported = true
			}
		}

		if !isUnexported {
			continue // Only care about unexported symbols
		}

		// Skip if the reference file is also being moved (they'll stay in the same package)
		if movedFiles[ref.File] {
			continue
		}

		// If the reference file is NOT being moved, it will stay in its current package.
		// Since we're moving files to a different package, any reference from a non-moved file
		// to an unexported symbol will become a cross-package reference after the move.

		// Find which file defines this symbol (one of the moved files)
		var definingFile string
		for _, op := range ops {
			// For now, we'll assume the symbol is in one of the moved files
			// In a more sophisticated implementation, we'd parse the moved files to check
			definingFile = op.From
			break
		}

		// Track this unexported symbol reference
		key := symbolName + ":" + definingFile
		if existing, ok := unexportedMap[key]; ok {
			existing.ReferencedFrom = append(existing.ReferencedFrom, ref.File)
		} else {
			unexportedMap[key] = &UnexportedRef{
				SymbolName:     symbolName,
				FileName:       definingFile,
				ReferencedFrom: []string{ref.File},
			}
		}
	}

	// Convert map to slice
	var result []UnexportedRef
	for _, ref := range unexportedMap {
		result = append(result, *ref)
	}

	return result
}

// buildUnexportedSymbolError creates a user-friendly error message with copy-pasteable commands
// to fix unexported symbol issues.
func buildUnexportedSymbolError(unexportedRefs []UnexportedRef) error {
	var msg strings.Builder
	msg.WriteString("Error: Cannot move files - unexported symbols are referenced from other packages\n\n")
	msg.WriteString("The following symbols in files being moved are unexported but referenced externally:\n")

	// Get current working directory to make paths relative
	cwd, _ := os.Getwd()

	for _, ref := range unexportedRefs {
		msg.WriteString(fmt.Sprintf("  • %s (%s) - referenced in %s\n",
			ref.SymbolName,
			filepath.Base(ref.FileName),
			strings.Join(mapBasenames(ref.ReferencedFrom), ", ")))
	}

	msg.WriteString("\nTo fix, first export these symbols by running:\n\n")

	for _, ref := range unexportedRefs {
		// Generate the dissect move command with relative paths
		// Format: dissect move file.go:oldName file.go:NewName
		exportedName := exportSymbolName(ref.SymbolName)
		relPath := ref.FileName
		if cwd != "" {
			if rel, err := filepath.Rel(cwd, ref.FileName); err == nil {
				relPath = rel
			}
		}
		msg.WriteString(fmt.Sprintf("  dissect move %s:%s %s:%s\n",
			relPath,
			ref.SymbolName,
			relPath,
			exportedName))
	}

	msg.WriteString("\nThen retry the batch move.")

	return fmt.Errorf("%s", msg.String())
}

// exportSymbolName capitalizes the first letter of a symbol name (or the last part if it's a field)
func exportSymbolName(symbolName string) string {
	parts := strings.Split(symbolName, ".")
	lastPart := parts[len(parts)-1]

	if len(lastPart) == 0 {
		return symbolName
	}

	// Capitalize the first letter
	exported := strings.ToUpper(string(lastPart[0])) + lastPart[1:]

	if len(parts) > 1 {
		// It's a field - reconstruct the full path
		parts[len(parts)-1] = exported
		return strings.Join(parts, ".")
	}

	return exported
}

// mapBasenames converts a slice of file paths to their basenames
func mapBasenames(paths []string) []string {
	result := make([]string, len(paths))
	for i, path := range paths {
		result[i] = filepath.Base(path)
	}
	return result
}
