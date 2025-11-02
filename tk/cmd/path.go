package main

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var dbPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the current database path",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		path, err := database.GetDBPath()
		if err != nil {
			return err
		}

		if jsonOutput {
			output := map[string]string{"path": path}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal output: %w", err)
			}
			fmt.Println(string(data))
		} else {
			fmt.Println(path)
		}
		return nil
	},
}
