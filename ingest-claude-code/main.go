package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ingest-claude-code",
	Short: "Ingest Claude Code traces to JSONL",
	Long:  `ingest-claude-code extracts sessions and messages from Claude Code conversation traces and outputs them as JSONL.`,
}

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Output sessions as JSONL",
	Long:  `Extract session metadata from all Claude Code trace files and output one JSONL line per session.`,
	Run:   runSessions,
}

var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Output messages as JSONL",
	Long:  `Extract all messages from Claude Code trace files and output one JSONL line per message.`,
	Run:   runMessages,
}

func main() {
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(messagesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
