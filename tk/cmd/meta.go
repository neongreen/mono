package cmd

import (
	meta_pkg "github.com/neongreen/mono/tk/cmd/meta"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Parent command; JSON only required for read-only data commands
var metaCmd = &cobra.Command{
	Use:   "meta",
	Short: "Manage task metadata",
	Long: `Manage task metadata with claims and authority resolution.

Metadata keys can have competing values from different roles (human, agent, qa, etc).
The effective value is resolved using the same authority lattice as status axes.`,
}

func init() {
	metaCmd.AddCommand(meta_pkg.SetCmd)
	metaCmd.AddCommand(meta_pkg.GetCmd)
	metaCmd.AddCommand(meta_pkg.ListCmd)
	metaCmd.AddCommand(meta_pkg.ClaimsCmd)
}
