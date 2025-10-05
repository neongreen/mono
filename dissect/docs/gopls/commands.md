# Gopls commands

## Internal commands

Gopls provides a bunch of internal lower-level commands.

List: <https://pkg.go.dev/golang.org/x/tools/gopls/internal/protocol/command>

As of Jun 29, 2025:

```go
const (
	AddDependency           Command = "gopls.add_dependency"
	AddImport               Command = "gopls.add_import"
	AddTelemetryCounters    Command = "gopls.add_telemetry_counters"
	AddTest                 Command = "gopls.add_test"
	ApplyFix                Command = "gopls.apply_fix"
	Assembly                Command = "gopls.assembly"
	ChangeSignature         Command = "gopls.change_signature"
	CheckUpgrades           Command = "gopls.check_upgrades"
	ClientOpenURL           Command = "gopls.client_open_url"
	DiagnoseFiles           Command = "gopls.diagnose_files"
	Doc                     Command = "gopls.doc"
	EditGoDirective         Command = "gopls.edit_go_directive"
	ExtractToNewFile        Command = "gopls.extract_to_new_file"
	FetchVulncheckResult    Command = "gopls.fetch_vulncheck_result"
	FreeSymbols             Command = "gopls.free_symbols"
	GCDetails               Command = "gopls.gc_details"
	Generate                Command = "gopls.generate"
	GoGetPackage            Command = "gopls.go_get_package"
	ListImports             Command = "gopls.list_imports"
	ListKnownPackages       Command = "gopls.list_known_packages"
	MaybePromptForTelemetry Command = "gopls.maybe_prompt_for_telemetry"
	MemStats                Command = "gopls.mem_stats"
	ModifyTags              Command = "gopls.modify_tags"
	Modules                 Command = "gopls.modules"
	PackageSymbols          Command = "gopls.package_symbols"
	Packages                Command = "gopls.packages"
	RegenerateCgo           Command = "gopls.regenerate_cgo"
	RemoveDependency        Command = "gopls.remove_dependency"
	ResetGoModDiagnostics   Command = "gopls.reset_go_mod_diagnostics"
	RunGoWorkCommand        Command = "gopls.run_go_work_command"
	RunGovulncheck          Command = "gopls.run_govulncheck"
	RunTests                Command = "gopls.run_tests"
	ScanImports             Command = "gopls.scan_imports"
	StartDebugging          Command = "gopls.start_debugging"
	StartProfile            Command = "gopls.start_profile"
	StopProfile             Command = "gopls.stop_profile"
	Tidy                    Command = "gopls.tidy"
	UpdateGoSum             Command = "gopls.update_go_sum"
	UpgradeDependency       Command = "gopls.upgrade_dependency"
	Vendor                  Command = "gopls.vendor"
	Views                   Command = "gopls.views"
	Vulncheck               Command = "gopls.vulncheck"
	WorkspaceStats          Command = "gopls.workspace_stats"
)
```

## Code actions

Listed at <https://github.com/golang/tools/blob/master/gopls/doc/features/transformation.md>.

```
quickfix, which applies unambiguously safe fixes
source.organizeImports
source.assembly
source.doc
source.freesymbols
source.test (undocumented)
source.addTest
source.toggleCompilerOptDetails
gopls.doc.features, which opens gopls' index of features in a browser
refactor.extract.constant
refactor.extract.function
refactor.extract.method
refactor.extract.toNewFile
refactor.extract.variable
refactor.extract.variable-all
refactor.inline.call
refactor.inline.variable
refactor.rewrite.addTags
refactor.rewrite.changeQuote
refactor.rewrite.fillStruct
refactor.rewrite.fillSwitch
refactor.rewrite.invertIf
refactor.rewrite.joinLines
refactor.rewrite.moveParamLeft
refactor.rewrite.moveParamRight
refactor.rewrite.removeTags
refactor.rewrite.removeUnusedParam
refactor.rewrite.splitLines
```