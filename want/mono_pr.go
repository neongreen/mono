package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/lib/cli"
	"github.com/neongreen/mono/lib/ghclient"
)

// listOpenPRs fetches open PRs that modify the given project
func listOpenPRs(project string) ([]PRInfo, error) {
	ctx := context.Background()
	client := ghclient.NewClient(ctx)

	prs, _, err := client.PullRequests.List(ctx, "neongreen", "mono", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}

	var projectPRs []PRInfo
	for _, pr := range prs {
		if pr.Number == nil || pr.Title == nil || pr.Head == nil || pr.Head.Ref == nil {
			continue
		}

		files, _, err := client.PullRequests.ListFiles(ctx, "neongreen", "mono", *pr.Number, nil)
		if err != nil {
			continue
		}

		modifiesProject := false
		projectPrefix := project + "/"
		for _, file := range files {
			if file.Filename == nil {
				continue
			}
			if strings.HasPrefix(*file.Filename, projectPrefix) {
				modifiesProject = true
				break
			}
		}

		if modifiesProject {
			projectPRs = append(projectPRs, PRInfo{
				Number: *pr.Number,
				Title:  *pr.Title,
				Branch: *pr.Head.Ref,
			})
		}
	}

	return projectPRs, nil
}

// buildMonoFromPR builds a project from a PR branch
func buildMonoFromPR(project string, prNumber int, dryRun bool, planJson bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".local", "bin")
	destPath := filepath.Join(destDir, project)

	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("mono %s@pr-%d", project, prNumber),
		Steps: []PlanStep{
			{
				Type:        "download",
				Description: fmt.Sprintf("Fetch PR #%d information from GitHub", prNumber),
				Command:     fmt.Sprintf("GET https://api.github.com/repos/neongreen/mono/pulls/%d", prNumber),
				Automatic:   true,
			},
			{
				Type:        "download",
				Description: "Clone neongreen/mono repository (PR branch)",
				Command:     "git clone --depth=1 --branch <pr-branch> https://github.com/neongreen/mono.git <tmpdir>",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: fmt.Sprintf("Build %s from source", project),
				Command:     fmt.Sprintf("mise run %s:build", project),
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: fmt.Sprintf("Copy binary to %s", destPath),
				Command:     fmt.Sprintf("cp <tmpdir>/%s/_build/%s %s", project, project, destPath),
				Automatic:   true,
			},
			{
				Type:        "configure",
				Description: "Make binary executable",
				Command:     fmt.Sprintf("chmod +x %s", destPath),
				Automatic:   true,
			},
		},
	}

	if planJson {
		jsonStr, err := plan.ToJSON()
		if err != nil {
			fmt.Printf("Error: Failed to generate JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(jsonStr)
		return
	}

	if dryRun {
		fmt.Printf("Building %s from PR #%d...\n", project, prNumber)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	fmt.Printf("Building %s from PR #%d...\n", project, prNumber)
	fmt.Println()
	plan.PrintPlan()
	fmt.Println()

	ctx := context.Background()
	client := ghclient.NewClient(ctx)
	pr, _, err := client.PullRequests.Get(ctx, "neongreen", "mono", prNumber)
	if err != nil {
		fmt.Printf("Error: Failed to fetch PR #%d: %v\n", prNumber, err)
		os.Exit(1)
	}

	if pr.Head == nil || pr.Head.Ref == nil {
		fmt.Printf("Error: PR #%d has no head branch\n", prNumber)
		os.Exit(1)
	}

	branch := *pr.Head.Ref
	fmt.Printf("PR #%d: %s\n", prNumber, *pr.Title)
	fmt.Printf("Branch: %s\n", cli.Key(branch))
	fmt.Println()

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("want-mono-%s-pr-%d-*", project, prNumber))
	if err != nil {
		fmt.Printf("Error: Failed to create temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Cloning neongreen/mono (branch: %s)...\n", cli.Key(branch))
	cmd := exec.Command("git", "clone", "--depth=1", "--branch", branch,
		"https://github.com/neongreen/mono.git", tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to clone repository: %v\n", err)
		os.Exit(1)
	}

	if err := buildWithMiseAndCopy(project, tmpDir, destPath); err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("%s Built and installed %s from PR #%d to: %s\n", cli.Success("✓"), cli.Key(project), prNumber, cli.Path(destPath))
	fmt.Println()

	printPathInfo(project, destPath)
}
