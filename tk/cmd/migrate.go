package cmd

import (
	migrate_pkg "github.com/neongreen/mono/tk/cmd/migrate"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration and deprecation tools",
	Long:  `Tools for managing database migrations and tracking deprecated code.`,
}

func init() {
	migrateCmd.AddCommand(migrate_pkg.ScanDeprecatedCmd)
	migrateCmd.AddCommand(migrate_pkg.FixRelocateBugCmd)
}
