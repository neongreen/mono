package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/neongreen/mono/lib/ghrelease"
)

// listMonoReleases lists all releases for a project from neongreen/mono
func listMonoReleases(project string) {
	fmt.Printf("Fetching releases for %s from neongreen/mono...\n", project)
	fmt.Println()

	releases, err := ghrelease.ListReleases("neongreen", "mono")
	if err != nil {
		fmt.Printf("Error: Failed to fetch releases: %v\n", err)
		os.Exit(1)
	}

	// Filter releases for the specified project
	var projectReleases []ghrelease.Release
	prefix := project + "--"
	for _, release := range releases {
		if strings.HasPrefix(release.TagName, prefix) {
			projectReleases = append(projectReleases, release)
		}
	}

	fmt.Printf("Fetching open PRs for %s...\n", project)
	openPRs, err := listOpenPRs(project)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch open PRs: %v\n", err)
		fmt.Println()
	}

	if len(projectReleases) == 0 && len(openPRs) == 0 {
		fmt.Printf("No releases or open PRs found for %s\n", project)
		fmt.Println("\nAvailable projects in mono:")
		fmt.Println("  printpdf, dissect, want, prrun, markdown-format, ingest, conf, tk")
		os.Exit(1)
	}

	if len(projectReleases) > 0 {
		fmt.Printf("\nAvailable releases for %s:\n", project)
		fmt.Println()
		for _, release := range projectReleases {

			version := strings.TrimPrefix(release.TagName, prefix)

			status := ""
			if release.Prerelease {
				status = " (prerelease)"
			}

			fmt.Printf("  %s%s\n", version, status)
		}
	}

	if len(openPRs) > 0 {
		fmt.Printf("\nOpen PRs for %s:\n", project)
		fmt.Println()
		for _, pr := range openPRs {
			fmt.Printf("  pr-%d: %s\n", pr.Number, pr.Title)
		}
	}

	fmt.Println()
	fmt.Println("To install a specific version:")
	fmt.Printf("  want mono %s@<version>\n", project)
	fmt.Println()
	fmt.Println("Examples:")
	if len(projectReleases) > 0 {
		version := strings.TrimPrefix(projectReleases[0].TagName, prefix)
		fmt.Printf("  want mono %s@%s\n", project, version)
	}
	if len(openPRs) > 0 {
		fmt.Printf("  want mono %s@pr-%d\n", project, openPRs[0].Number)
	}
}
