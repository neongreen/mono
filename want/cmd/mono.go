package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listReleases bool

var monoCmd = &cobra.Command{
	Use:   "mono <project>[@version]",
	Short: "Install tools from neongreen/mono repository",
	Long: `Install tools from the neongreen/mono repository.

You can specify a version, branch, commit, or PR to install from.

Examples:
  want mono printpdf                     # Build from main branch (default)
  want mono printpdf --list              # List all releases and open PRs
  want mono printpdf@main.1              # Install version main.1
  want mono printpdf@local               # Build from current directory
  want mono dissect@feature-branch       # Build from a specific branch
  want mono want@abc1234                 # Build from a specific commit
  want mono dissect@pr-42                # Build from PR #42`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		project := args[0]

		// If --list flag is set, show releases
		if listReleases {
			return listMonoReleases(project)
		}

		// Parse project[@version]
		var version string
		if strings.Contains(project, "@") {
			parts := strings.SplitN(project, "@", 2)
			project = parts[0]
			version = parts[1]
		} else {
			// Default to main if no version specified
			version = "main"
		}

		return installMonoRelease(project, version, dryRun, planJSON)
	},
}

func init() {
	monoCmd.Flags().BoolVar(&listReleases, "list", false, "List all releases and open PRs")
	RootCmd.AddCommand(monoCmd)
}

// These functions will be properly wired up to the main package implementations
// They are exported so cmd can call them

var (
	// Function references that will be set by main package
	ListMonoReleasesFunc   func(string)
	InstallMonoReleaseFunc func(string, string, bool, bool)
)

func listMonoReleases(project string) error {
	if ListMonoReleasesFunc != nil {
		ListMonoReleasesFunc(project)
		return nil
	}
	return fmt.Errorf("listMonoReleases not implemented")
}

func installMonoRelease(project, version string, dryRun, planJSON bool) error {
	if InstallMonoReleaseFunc != nil {
		InstallMonoReleaseFunc(project, version, dryRun, planJSON)
		return nil
	}
	return fmt.Errorf("installMonoRelease not implemented")
}
