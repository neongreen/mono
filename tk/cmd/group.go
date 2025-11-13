package cmd

import (
	group_pkg "github.com/neongreen/mono/tk/cmd/group"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups (unordered sets)",
	Long:  `Create and manage group containers (unordered collections).`,
}

func init() {
	groupCmd.AddCommand(group_pkg.CreateCmd)
	groupCmd.AddCommand(group_pkg.ListCmd)
	groupCmd.AddCommand(group_pkg.ShowCmd)
	groupCmd.AddCommand(group_pkg.AddCmd)
	groupCmd.AddCommand(group_pkg.RemoveCmd)
	groupCmd.AddCommand(group_pkg.RenameCmd)
	groupCmd.AddCommand(group_pkg.RmCmd)
}
