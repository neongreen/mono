package main

import (
	"fmt"
	"os"

	"github.com/neongreen/mono/want/cmd"
)

func main() {
	// Wire up the handlers
	// Wrap getCompoundHandler to match the expected type
	getCompoundHandlerWrapped := func(s string) (cmd.CompoundHandlerFunc, bool) {
		handler, ok := getCompoundHandler(s)
		if !ok {
			return nil, false
		}
		// Convert CompoundHandler to CompoundHandlerFunc
		return cmd.CompoundHandlerFunc(handler), true
	}

	// Wrap handler functions to return errors
	handleJsonCommandWrapped := func(args []string, dryRun bool, planJSON bool) error {
		handleJsonCommand(args, dryRun, planJSON)
		return nil
	}

	handleMarkdownCommandWrapped := func(args []string, dryRun bool, planJSON bool) error {
		handleMarkdownCommand(args, dryRun, planJSON)
		return nil
	}

	handleExcalifontCommandWrapped := func(args []string, dryRun bool, planJSON bool) error {
		handleExcalifontCommand(args, dryRun, planJSON)
		return nil
	}

	cmd.SetHandlers(
		listMonoReleases,
		installMonoRelease,
		getCompoundHandlerWrapped,
		handleGitHubAsset,
		installToolViaMise,
		handleJsonCommandWrapped,
		handleMarkdownCommandWrapped,
		handleExcalifontCommandWrapped,
	)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
