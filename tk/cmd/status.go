package cmd

import (
	status_pkg "github.com/neongreen/mono/tk/cmd/status"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Manage task status and sync status",
}

func init() {
	statusCmd.AddCommand(status_pkg.SyncCmd)
}
