package cmd

// CompoundHandlerFunc is the type for compound command handlers
type CompoundHandlerFunc func(args []string, dryRun bool, planJson bool)

// SetHandlers is called by main package to wire up the actual implementations
func SetHandlers(
	listMonoReleases func(string),
	installMonoRelease func(string, string, bool, bool),
	getCompoundHandler func(string) (CompoundHandlerFunc, bool),
	handleGitHubAsset func(string, bool, bool),
	installToolViaMise func(string, bool, bool),
) {
	ListMonoReleasesFunc = listMonoReleases
	InstallMonoReleaseFunc = installMonoRelease
	GetCompoundHandlerFunc = getCompoundHandler
	HandleGitHubAssetFunc = handleGitHubAsset
	InstallToolViaMiseFunc = installToolViaMise
}
