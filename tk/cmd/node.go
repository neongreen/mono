package cmd

import (
	node_pkg "github.com/neongreen/mono/tk/cmd/node"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage node ID",
}

func init() {
	nodeCmd.AddCommand(node_pkg.ShowCmd)
	nodeCmd.AddCommand(node_pkg.RegenCmd)
}
