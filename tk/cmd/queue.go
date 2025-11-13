package cmd

import (
	queue_pkg "github.com/neongreen/mono/tk/cmd/queue"
	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Manage queues (FIFO containers)",
	Long:  `Create and manage queue containers (first in, first out).`,
}

func init() {
	queueCmd.AddCommand(queue_pkg.CreateCmd)
	queueCmd.AddCommand(queue_pkg.ListCmd)
	queueCmd.AddCommand(queue_pkg.ShowCmd)
	queueCmd.AddCommand(queue_pkg.PushCmd)
	queueCmd.AddCommand(queue_pkg.PopCmd)
	queueCmd.AddCommand(queue_pkg.RenameCmd)
	queueCmd.AddCommand(queue_pkg.RmCmd)
}
