package cmd

import (
	"github.com/neongreen/mono/lib/version"
)

func init() {
	RootCmd.AddCommand(version.NewVersionCommand("want"))
}
