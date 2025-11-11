package cmd

import (
	log_pkg "github.com/neongreen/mono/tk/cmd/log"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Query and search invocation logs",
	Long:  `Tools for querying and searching tk invocation logs.`,
}

func init() {
	logCmd.AddCommand(log_pkg.QueryCmd)
	logCmd.AddCommand(log_pkg.SearchCmd)
}
