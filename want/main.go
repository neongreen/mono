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

	cmd.SetHandlers(
		listMonoReleases,
		installMonoRelease,
		getCompoundHandlerWrapped,
		handleGitHubAsset,
		installToolViaMise,
	)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
