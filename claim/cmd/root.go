package cmd

import (
	"github.com/neongreen/mono/claim/internal/logger"
	"github.com/spf13/cobra"
)

var debugFlag bool

var RootCmd = &cobra.Command{
	Use:   "claim",
	Short: "Lightweight claim-checking tool using Claude",
	Long: `claim scans files for @claim and @lens blocks and uses Claude to verify
that claims are properly supported by their evidence.

It helps catch false positives by requiring explicit proof for every claim.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.SetDebug(debugFlag)
	},
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
}
