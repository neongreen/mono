package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/tk/internal/collision"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/segment"
	"github.com/neongreen/mono/tk/internal/sync"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [remote-name]",
	Short: "Push local segments to remote",
	Long: `Push local segment files to a remote.

Examples:
  tk push icloud
  tk push icloud --space personal
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space, _ := cmd.Flags().GetString("space")

		// Get remote name
		var remoteName string
		if len(args) > 0 {
			remoteName = args[0]
		}

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// If no remote name provided, try to find default
		if remoteName == "" {
			if len(config.Remotes) == 0 {
				return fmt.Errorf("no remotes configured")
			}
			if len(config.Remotes) == 1 {
				for name := range config.Remotes {
					remoteName = name
				}
			} else {
				return fmt.Errorf("multiple remotes configured; please specify which one")
			}
		}

		// Get remote config
		remote, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found", remoteName)
		}

		if !remote.Push {
			return fmt.Errorf("push is disabled for remote '%s'", remoteName)
		}

		// Get state directory
		stateDir, err := config_pkg.GetStateDir()
		if err != nil {
			return err
		}

		// Load local index mirror
		localIndexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
		localIndex, err := loadIndexFile(localIndexPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load local index: %w", err)
		}
		if localIndex == nil {
			fmt.Println("No local segments to push")
			return nil
		}

		// Check if remote index exists
		remoteIndexPath := filepath.Join(remote.Path, space, "index.json")
		remoteIndex, err := loadIndexFile(remoteIndexPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load remote index: %w", err)
		}

		// If remote index doesn't exist, create it
		if remoteIndex == nil {
			remoteIndex = &sync.IndexFile{
				Schema:   "tk.index.v1",
				Space:    space,
				Segments: []sync.SegmentInfo{},
			}
		}

		// Find segments that are in local but not in remote
		remoteSegmentPaths := make(map[string]bool)
		for _, seg := range remoteIndex.Segments {
			remoteSegmentPaths[seg.Rel] = true
		}

		var segmentsToPush []sync.SegmentInfo
		for _, seg := range localIndex.Segments {
			if !remoteSegmentPaths[seg.Rel] {
				segmentsToPush = append(segmentsToPush, seg)
			}
		}

		if len(segmentsToPush) == 0 {
			fmt.Println("No new segments to push")
			return nil
		}

		// Copy segment files to remote (they should already exist locally in the remote path)
		// Since we're using folder remotes, the segments are already written to the remote path
		// We just need to update the index

		// Add new segments to remote index
		remoteIndex.Segments = append(remoteIndex.Segments, segmentsToPush...)

		// Save remote index
		if err := saveIndexFile(remoteIndexPath, remoteIndex); err != nil {
			return fmt.Errorf("failed to save remote index: %w", err)
		}

		fmt.Printf("Pushed %d segments, index updated\n", len(segmentsToPush))
		return nil
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull [remote-name]",
	Short: "Pull segments from remote",
	Long: `Pull segment files from a remote.

Examples:
  tk pull icloud
  tk pull icloud --space personal
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space, _ := cmd.Flags().GetString("space")

		// Get remote name
		var remoteName string
		if len(args) > 0 {
			remoteName = args[0]
		}

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// If no remote name provided, try to find default
		if remoteName == "" {
			if len(config.Remotes) == 0 {
				return fmt.Errorf("no remotes configured")
			}
			if len(config.Remotes) == 1 {
				for name := range config.Remotes {
					remoteName = name
				}
			} else {
				return fmt.Errorf("multiple remotes configured; please specify which one")
			}
		}

		// Get remote config
		remote, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found", remoteName)
		}

		if !remote.Pull {
			return fmt.Errorf("pull is disabled for remote '%s'", remoteName)
		}

		// Load remote index
		remoteIndexPath := filepath.Join(remote.Path, space, "index.json")
		remoteIndex, err := loadIndexFile(remoteIndexPath)
		if err != nil {
			if os.IsNotExist(err) {
				// Try to reconstruct index by scanning segments
				fmt.Println("Remote index not found, scanning segments...")
				remoteIndex, err = reconstructIndex(remote.Path, space)
				if err != nil {
					return fmt.Errorf("failed to reconstruct index: %w", err)
				}
				// Save reconstructed index
				if err := saveIndexFile(remoteIndexPath, remoteIndex); err != nil {
					fmt.Printf("Warning: failed to save reconstructed index: %v\n", err)
				}
			} else {
				return fmt.Errorf("failed to load remote index: %w", err)
			}
		}

		if remoteIndex == nil || len(remoteIndex.Segments) == 0 {
			fmt.Println("No segments found on remote")
			return nil
		}

		fmt.Printf("Found %d segments on remote\n", len(remoteIndex.Segments))

		// Verify segments exist - if any are missing, regenerate index from actual files
		missingSegments := 0
		for _, seg := range remoteIndex.Segments {
			segPath := filepath.Join(remote.Path, seg.Rel)
			if _, err := os.Stat(segPath); err != nil {
				missingSegments++
			}
		}

		// If any segments are missing, regenerate index from actual files
		if missingSegments > 0 {
			fmt.Printf("Index out of sync (%d missing segments), regenerating from disk...\n", missingSegments)
			remoteIndex, err = reconstructIndex(remote.Path, space)
			if err != nil {
				return fmt.Errorf("failed to reconstruct index: %w", err)
			}
			// Save regenerated index to remote
			if err := saveIndexFile(remoteIndexPath, remoteIndex); err != nil {
				fmt.Printf("Warning: failed to save regenerated index: %v\n", err)
			}
		}

		// For folder remotes, segments are already accessible locally
		// Count segments that actually exist
		segmentCount := 0
		for _, seg := range remoteIndex.Segments {
			segPath := filepath.Join(remote.Path, seg.Rel)
			if _, err := os.Stat(segPath); err == nil {
				segmentCount++
			}
		}

		// Save local index mirror
		stateDir, err := config_pkg.GetStateDir()
		if err != nil {
			return err
		}

		localIndexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
		if err := saveIndexFile(localIndexPath, remoteIndex); err != nil {
			return fmt.Errorf("failed to save local index mirror: %w", err)
		}

		fmt.Printf("Pulled %d segments\n", segmentCount)
		return nil
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync [remote-name]",
	Short: "Sync with remote (pull → ingest → export → push)",
	Long: `Sync local events with a remote.

This command performs the full sync flow:
1. Pull segments from remote
2. Ingest events from remote segments
3. Export new local events to segments
4. Push new segments to remote

Examples:
  tk sync icloud
  tk sync icloud --space personal
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space, _ := cmd.Flags().GetString("space")

		// Get remote name
		var remoteName string
		if len(args) > 0 {
			remoteName = args[0]
		}

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return err
		}

		// If no remote name provided, try to find default
		if remoteName == "" {
			if len(config.Remotes) == 0 {
				return fmt.Errorf("no remotes configured")
			}
			if len(config.Remotes) == 1 {
				for name := range config.Remotes {
					remoteName = name
				}
			} else {
				return fmt.Errorf("multiple remotes configured; please specify which one")
			}
		}

		// Get remote config
		remote, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found", remoteName)
		}

		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check for node collisions first
		fmt.Println("Checking for node collisions...")
		if err := collision.CheckNodeCollision(db, remoteName, remote); err != nil {
			return fmt.Errorf("node collision detected: %w", err)
		}

		// 1. Pull
		fmt.Println("Pulling from remote...")
		if remote.Pull {
			pullCmd.Flags().Set("space", space)
			if err := pullCmd.RunE(pullCmd, []string{remoteName}); err != nil {
				return fmt.Errorf("pull failed: %w", err)
			}
		}

		// 2. Ingest
		fmt.Println("Ingesting events...")
		if err := IngestRemote(db, remoteName, remote); err != nil {
			return fmt.Errorf("ingest failed: %w", err)
		}

		// 3. Export
		fmt.Println("Exporting local events...")
		exportCmd.Flags().Set("space", space)
		if err := exportCmd.RunE(exportCmd, []string{remoteName}); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}

		// 4. Push
		fmt.Println("Pushing to remote...")
		if remote.Push {
			pushCmd.Flags().Set("space", space)
			if err := pushCmd.RunE(pushCmd, []string{remoteName}); err != nil {
				return fmt.Errorf("push failed: %w", err)
			}
		}

		fmt.Println("Sync complete")
		return nil
	},
}

// loadIndexFile loads an index file
func loadIndexFile(path string) (*sync.IndexFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var index sync.IndexFile
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index: %w", err)
	}

	return &index, nil
}

// saveIndexFile saves an index file
func saveIndexFile(path string, index *sync.IndexFile) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// reconstructIndex scans a segments directory and reconstructs the index
func reconstructIndex(remotePath, space string) (*sync.IndexFile, error) {
	segmentsDir := filepath.Join(remotePath, space, "segments")
	if _, err := os.Stat(segmentsDir); os.IsNotExist(err) {
		return &sync.IndexFile{
			Schema:   "tk.index.v1",
			Space:    space,
			Segments: []sync.SegmentInfo{},
		}, nil
	}

	var segments []sync.SegmentInfo
	err := filepath.Walk(segmentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".zst" {
			// Calculate relative path
			relPath, err := filepath.Rel(remotePath, path)
			if err != nil {
				return err
			}

			// Calculate SHA256
			sha, err := segment.CalculateSHA256(path)
			if err != nil {
				return fmt.Errorf("failed to calculate SHA256 for %s: %w", path, err)
			}

			segments = append(segments, sync.SegmentInfo{
				Rel:    relPath,
				SHA256: sha,
				Size:   info.Size(),
				MTime:  info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &sync.IndexFile{
		Schema:   "tk.index.v1",
		Space:    space,
		Segments: segments,
	}, nil
}

func init() {
	pushCmd.Flags().String("space", "personal", "Space to push")
	pullCmd.Flags().String("space", "personal", "Space to pull")
	syncCmd.Flags().String("space", "personal", "Space to sync")
}
