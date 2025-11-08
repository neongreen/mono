package cmd

import (
	events_pkg "github.com/neongreen/mono/tk/cmd/events"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Debug commands for inspecting events",
}

func init() {
	eventsCmd.AddCommand(events_pkg.ListCmd)
	eventsCmd.AddCommand(events_pkg.ShowCmd)
	eventsCmd.AddCommand(events_pkg.StatsCmd)
}
