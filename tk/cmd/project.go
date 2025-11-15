package cmd

import (
	project_pkg "github.com/neongreen/mono/tk/cmd/project"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Parent command; JSON only required for read-only data commands
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `Create and manage projects.`,
}

func init() {
	projectCmd.AddCommand(project_pkg.CreateCmd)
	projectCmd.AddCommand(project_pkg.LsCmd)
	projectCmd.AddCommand(project_pkg.RmCmd)
	projectCmd.AddCommand(project_pkg.RenameCmd)
}
