package main

import (
	"go/ast"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neongreen/mono/dissect/pkg/commands"
	"github.com/neongreen/mono/dissect/pkg/gopls"
	"github.com/neongreen/mono/dissect/pkg/goutils"
	"github.com/neongreen/mono/dissect/pkg/lsp"
	"github.com/neongreen/mono/dissect/pkg/refactor"
	"github.com/neongreen/mono/dissect/pkg/utils"
)

type RefactorStatus int

const (
	NothingToDo RefactorStatus = iota
	Refactored
	Skipped
	Failed
)

// Extract functions from the given file and move them to new files.
// If lspClient is provided, it will be used instead of CLI gopls calls.
func ProcessFile(absPath string, goplsPath string, goimportsPath string, lspClient *lsp.Client) (status RefactorStatus, exclusionReason string, err error) {
	// Use relPath in logs for humans and absPath in debug logs
	relPath := cwdRelPath(absPath)

	slog.Debug("Processing file", "file", relPath)

	moduleRoot, err := commands.FindGoModuleRoot(absPath)
	if err != nil {
		slog.Error("Error finding Go module root", "error", err, "file", absPath)
		return Failed, "", err
	}

	status = NothingToDo

	// Check if we should even consider refactoring this file
	shouldRefactor, exclusionReason, err := goutils.ShouldRefactor(absPath)
	if err != nil {
		slog.Error("Error checking if file should be refactored", "file", relPath, "error", err)
		return Failed, "", err
	}
	if !shouldRefactor {
		slog.Info("Skipping file", "file", cwdRelPath(absPath), "reason", exclusionReason)
		return Skipped, exclusionReason, nil
	}

	changedFiles := make(map[string]struct{})

	for {
		_, node, err := goutils.ReadGoFile(absPath)
		if err != nil {
			return Failed, "", err
		}

		packageName := node.Name.Name

		foundFunctionToMove := false
		for _, decl := range node.Decls {
			if fn, isFn := decl.(*ast.FuncDecl); isFn {
				funcName := fn.Name.Name
				slog.Debug("Found function", "package", packageName, "file", absPath, "function", funcName)

				// Skip main function, it should not be refactored
				if funcName == "main" {
					slog.Info("Skipping main function", "file", relPath)
					continue
				}

				// Determine target for extraction
				targetFilePath, targetFuncName, targetPackageName := refactor.DetermineExtractionTarget(absPath, fn)
				relTargetFilePath := cwdRelPath(targetFilePath)

				// Skip function if it is already in the file it should be
				if targetFilePath == absPath {
					slog.Debug("Skipping function as it is already in the right place",
						"function", funcName, "file", targetFilePath)
					continue
				}

				// Come up with a new unique name based on hash
				tempFuncName := ""
				slog.Debug("Attempting to get path relative to module root", "moduleRoot", moduleRoot, "file", absPath)
				if relPath, err := filepath.Rel(moduleRoot, absPath); err != nil {
					slog.Error("Error getting relative path", "moduleRoot", moduleRoot, "file", absPath, "error", err)
					return Failed, "", err
				} else {
					tempFuncName = "Temp_" + funcName + "_" + utils.HashString(relPath + "_" + funcName)[:8]
				}

				slog.Info("Moving function to a new file",
					"function", funcName, "from", relPath, "to", relTargetFilePath)

				// Gopls decides the file name based on the function name.
				// Since *we* want to control the file name, we rename the function to a temp name first so that we don't get a clash.
				// Once the function is moved, we rename it back to the original name and rename the file.
				renameStart := time.Now()
				var renameErr error
				if lspClient != nil {
					renameErr = lsp.RenameWithClient(lspClient, absPath, funcName, tempFuncName)
				} else {
					renameErr = gopls.Rename(goplsPath, absPath, funcName, tempFuncName, moduleRoot)
				}
				if renameErr != nil {
					slog.Error("Error renaming function", "from", funcName, "to", tempFuncName, "error", renameErr)
					return Failed, "", renameErr
				}
				slog.Debug("[TIMING] gopls.Rename", "duration", time.Since(renameStart), "function", funcName)
				changedFiles[absPath] = struct{}{}

				extractStart := time.Now()
				var goplsFilePath string
				var extractErr error
				if lspClient != nil {
					goplsFilePath, extractErr = lsp.ExtractToNewFileWithClient(lspClient, absPath, tempFuncName)
				} else {
					goplsFilePath, extractErr = gopls.ExtractToNewFile(goplsPath, absPath, tempFuncName, moduleRoot)
				}
				if extractErr != nil {
					slog.Error("Error extracting function with gopls",
						"file", absPath, "function", tempFuncName, "error", extractErr,
					)
					return Failed, "", extractErr
				}
				slog.Debug("[TIMING] gopls.ExtractToNewFile", "duration", time.Since(extractStart), "function", funcName)
				status = Refactored
				slog.Debug("Gopls extracted function to a new file",
					"function", funcName, "source", absPath, "target", goplsFilePath,
				)

				// Create target directory if it doesn't exist
				if _, err := os.Stat(filepath.Dir(targetFilePath)); os.IsNotExist(err) {
					err := os.MkdirAll(filepath.Dir(targetFilePath), 0o755)
					if err != nil {
						slog.Error("Error creating target directory", "dir", filepath.Dir(targetFilePath), "error", err)
						return Failed, "", err
					}
				}

				// Move file
				err = utils.MoveFile(goplsFilePath, targetFilePath)
				if err != nil {
					slog.Error("Error moving file", "from", goplsFilePath, "to", targetFilePath, "error", err)
					return Failed, "", err
				}
				changedFiles[targetFilePath] = struct{}{}
				slog.Debug("Moved new file to", "file", targetFilePath)

				// Update package declaration and imports, if necessary
				if packageName != targetPackageName {
					slog.Info("Updating package declaration",
						"from", packageName, "to", targetPackageName, "file", relTargetFilePath)

					err = goutils.UpdatePackageDeclaration(targetFilePath, targetPackageName)
					if err != nil {
						slog.Error("Error updating package declaration", "file", targetFilePath, "to", targetPackageName, "error", err)
						return Failed, "", err
					}
					slog.Debug("Updated package declaration", "file", targetFilePath, "to", targetPackageName)

					// Add import to the original file
					importPath, err := commands.GetFullImportPath(targetFilePath)
					if err != nil {
						slog.Error("Error getting full import path", "file", targetFilePath, "error", err)
						return Failed, "", err
					}
					addImportStart := time.Now()
					var addImportErr error
					if lspClient != nil {
						addImportErr = lsp.AddImportWithClient(lspClient, absPath, importPath)
					} else {
						addImportErr = gopls.AddImport(goplsPath, absPath, importPath, moduleRoot)
					}
					if addImportErr != nil {
						slog.Error("Error adding import", "import", importPath, "to_file", absPath, "error", addImportErr)
						return Failed, "", addImportErr
					}
					slog.Debug("[TIMING] gopls.AddImport", "duration", time.Since(addImportStart), "import", importPath)
				}

				// Rename the function back. This time we can use search and replace, since the name is unique.

				// targetReference is the expression we will use to refer to the function in other files.
				// It can just be the function name, or it can be qualified with the package name if the package has changed.
				targetReference := targetFuncName
				if targetPackageName != packageName {
					targetReference = targetPackageName + "." + targetFuncName
				}
				slog.Debug("Renaming function back to original name using search and replace",
					"temp_name", tempFuncName, "target_reference", targetReference, "target_func_name", targetFuncName)
				findFilesStart := time.Now()
				allGoFiles, err := commands.FindGoFiles(moduleRoot)
				if err != nil {
					slog.Error("Error finding Go files in module root", "moduleRoot", moduleRoot, "error", err)
					return Failed, "", err
				}
				slog.Debug("[TIMING] commands.FindGoFiles", "duration", time.Since(findFilesStart), "fileCount", len(allGoFiles))
				slog.Debug("Found Go files to rename in", "moduleRoot", moduleRoot, "files", allGoFiles)

				replaceStart := time.Now()
				for _, goFile := range allGoFiles {
					// Use search and replace to rename the function back
					data, err := os.ReadFile(goFile) // Ensure the file exists
					if err != nil {
						slog.Error("Error reading Go file", "file", goFile, "error", err)
						return Failed, "", err
					}
					// In the target file we use function name, in other files we use reference (which can be qualified)
					newData := ""
					if goFile == targetFilePath {
						// In the target file we use function name
						newData = strings.ReplaceAll(string(data), tempFuncName, targetFuncName)
					} else {
						// In other files we use reference (which can be qualified)
						newData = strings.ReplaceAll(string(data), tempFuncName, targetReference)
					}
					if newData != string(data) {
						err = os.WriteFile(goFile, []byte(newData), 0o644)
						if err != nil {
							slog.Error("Error writing Go file", "file", goFile, "error", err)
							return Failed, "", err
						}
						changedFiles[goFile] = struct{}{}
					}
				}
				slog.Debug("[TIMING] search/replace loop", "duration", time.Since(replaceStart), "filesProcessed", len(allGoFiles))

				foundFunctionToMove = true
				break // Break inner loop to re-parse file and find next function
			}
		}
		if !foundFunctionToMove {
			break // No more functions to move
		}
	}

	if status == Refactored {
		// We have to run goimports, since moving files around can break imports.
		goimportsStart := time.Now()
		for file := range changedFiles {
			err := commands.RunGoimportsOnFile(goimportsPath, file)
			if err != nil {
				return Failed, "", err
			}
		}
		slog.Debug("[TIMING] goimports (all files)", "duration", time.Since(goimportsStart), "fileCount", len(changedFiles))
	}

	return status, "", nil
}
