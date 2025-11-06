package main

import (
	"github.com/neongreen/mono/lib/version"
)

func init() {
	rootCmd.AddCommand(version.NewVersionCommand("jj-run"))
}
