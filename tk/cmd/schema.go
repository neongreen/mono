package cmd

import (
	schema_pkg "github.com/neongreen/mono/tk/cmd/schema"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage container and type schemas",
	Long:  `Define and manage container kinds and other schema elements.`,
}

func init() {
	schemaCmd.AddCommand(schema_pkg.AddKindCmd)
	schemaCmd.AddCommand(schema_pkg.ListKindsCmd)
}
