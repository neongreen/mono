# Extract declarations to new file

The "Extract declarations to new file" feature in `gopls` allows users to move selected top-level declarations (functions, methods, variables, constants, types) from an existing Go file into a newly created Go file. This refactoring operation is exposed as a code action in LSP-compatible editors.

## Implementation Details

The functionality is implemented through a series of interconnected components within the `gopls` codebase:

1.  **Code Action Producer (`refactorExtractToNewFile` in [](../../_gotools/gopls/internal/golang/codeaction.go)):**
    *   This function acts as a producer for the "Extract declarations to new file" code action.
    *   It first checks if the current selection is valid for extraction by calling `canExtractToNewFile`.
    *   If valid, it constructs a `command.NewExtractToNewFileCommand` which encapsulates the details of the refactoring request.
    *   This command is then added to the list of available code actions that `gopls` offers to the client.

2.  **Command Handler (`ExtractToNewFile` in [](../../_gotools/gopls/internal/server/command.go)):**
    *   This function is the server-side handler for the `ExtractToNewFile` command. When the client requests to execute this code action, this handler is invoked.
    *   Its primary responsibility is to orchestrate the refactoring process. It calls the core logic function `golang.ExtractToNewFile` to obtain the necessary file modifications.
    *   Upon successful execution of the core logic, it uses `applyChanges` to send the generated `DocumentChange` operations back to the LSP client.

3.  **Core Refactoring Logic (`golang.ExtractToNewFile` in [](../../_gotools/gopls/internal/golang/extracttofile.go)):**
    *   This is the heart of the "Extract to new file" feature, responsible for performing the actual code manipulation.
    *   It takes the selected range and the current file's context as input.
    *   **Selection Validation and Expansion:** It uses `selectedToplevelDecls` to identify and validate the top-level declarations within the selected range. This function also expands the selection to include entire declarations and their associated comments, ensuring a syntactically correct extraction.
    *   **Import Management:** It analyzes the imports in the original file and the usage of imported packages within the selected declarations. `findImportEdits` determines which imports need to be moved to the new file and which can be removed from the original file.
    *   **New File Content Generation:** It constructs the complete content for the new Go file. This includes:
        *   Preserving copyright and build constraint comments from the original file.
        *   Adding the appropriate `package` declaration (same as the original file).
        *   Adding the necessary `import` statements identified by `findImportEdits`.
        *   Inserting the source code of the extracted declarations.
    *   **New File Naming:** It uses `chooseNewFile` to generate a unique and descriptive filename for the new file, typically based on the name of the first extracted symbol.
    *   **Change Generation:** Finally, it generates a slice of `protocol.DocumentChange` objects. These changes describe:
        *   The deletion of the extracted declarations and their associated imports from the original file.
        *   The creation of the new file.
        *   The insertion of the generated content into the new file.

4.  **Selection Validation Helper (`canExtractToNewFile` in [](../../_gotools/gopls/internal/golang/extracttofile.go)):**
    *   A utility function used by the code action producer to quickly determine if a given selection is eligible for extraction. It primarily relies on `selectedToplevelDecls` for its logic.

5.  **Change Application (`applyChanges` in [](../../_gotools/gopls/internal/server/command.go)):**
    *   This helper function is responsible for communicating the generated `DocumentChange` operations to the LSP client.
    *   It packages the changes into a `protocol.WorkspaceEdit` and sends it via `cli.ApplyEdit`. The LSP client then applies these edits to the user's open files and creates the new file as instructed.

In summary, the "Extract declarations to new file" feature in `gopls` provides a robust refactoring capability by intelligently analyzing code structure, managing dependencies (imports), generating new file content, and seamlessly applying these changes to the user's workspace through the LSP.

When extracting a function, the position provided to `gopls` should be the location of the function's *name*, not the first character of the `func` keyword.
