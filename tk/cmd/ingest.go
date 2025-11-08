package cmd

import (
	"fmt"
	"os"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/remote"
	"github.com/spf13/cobra"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest [path-or-remote]",
	Short: "Ingest events from segment files",
	Long: `Ingest events from segment files into the local database.

Examples:
  tk ingest /path/to/segment.jsonl.zst    # Ingest from a single file
  tk ingest icloud                         # Ingest from remote's segments
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("path or remote name required")
		}

		pathOrRemote := args[0]

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check if it's a file or a remote name
		if _, err := os.Stat(pathOrRemote); err == nil {
			// It's a file
			result, err := remote.IngestFile(db, pathOrRemote)
			if err != nil {
				return err
			}
			
			// Print warnings if any
			for _, projErr := range result.ProjectionErrors {
				fmt.Fprintf(os.Stderr, "Warning: projection failed for %s\n", projErr)
			}
			
			fmt.Printf("Ingested %d events (%d duplicates skipped)\n", result.EventsIngested, result.Duplicates)
			return nil
		}

		// Try as a remote name
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		remoteConfig, exists := config.Remotes[pathOrRemote]
		if !exists {
			return fmt.Errorf("'%s' is neither a file nor a configured remote", pathOrRemote)
		}

		stateDir, err := config_pkg.GetStateDir()
		if err != nil {
			return err
		}

		result, err := remote.IngestRemote(db, pathOrRemote, remoteConfig, stateDir)
		if err != nil {
			return err
		}

		// Print warnings if any
		for _, segErr := range result.SegmentErrors {
			fmt.Fprintf(os.Stderr, "Warning: failed to read segment %s\n", segErr)
		}
		for _, projErr := range result.ProjectionErrors {
			fmt.Fprintf(os.Stderr, "Warning: projection failed for %s\n", projErr)
		}

		if result.EventsIngested == 0 {
			fmt.Println("No new events to ingest")
		} else {
			fmt.Printf("Ingested %d events from %d segments (%d duplicates skipped)\n",
				result.EventsIngested, result.SegmentsRead, result.Duplicates)
		}
		return nil
	},
}

// IngestRemote is kept for backward compatibility with cmd/sync.go
// New code should use remote.IngestRemote directly
func IngestRemote(db *database.DB, remoteName string, remoteConfig config_pkg.RemoteConfig) error {
	stateDir, err := config_pkg.GetStateDir()
	if err != nil {
		return err
	}

	result, err := remote.IngestRemote(db, remoteName, remoteConfig, stateDir)
	if err != nil {
		return err
	}

	// Print warnings if any
	for _, segErr := range result.SegmentErrors {
		fmt.Fprintf(os.Stderr, "Warning: failed to read segment %s\n", segErr)
	}
	for _, projErr := range result.ProjectionErrors {
		fmt.Fprintf(os.Stderr, "Warning: projection failed for %s\n", projErr)
	}

	if result.EventsIngested == 0 {
		fmt.Println("No new events to ingest")
	} else {
		fmt.Printf("Ingested %d events from %d segments (%d duplicates skipped)\n",
			result.EventsIngested, result.SegmentsRead, result.Duplicates)
	}
	return nil
}
