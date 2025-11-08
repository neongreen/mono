package project

import (
	"github.com/spf13/cobra"
)

var AliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage project aliases",
	Long:  `Add or remove project aliases.`,
}

func init() {
	AliasCmd.AddCommand(AliasAddCmd)
	AliasCmd.AddCommand(AliasRemoveCmd)
}
