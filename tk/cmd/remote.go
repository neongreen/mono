package cmd

import (
	remote_pkg "github.com/neongreen/mono/tk/cmd/remote"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remotes",
}

func init() {
	remoteCmd.AddCommand(remote_pkg.AddCmd)
	remoteCmd.AddCommand(remote_pkg.LsCmd)
	remoteCmd.AddCommand(remote_pkg.RmCmd)
}
