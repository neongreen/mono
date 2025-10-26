package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv [old-id] [new-spec] [...]",
	Short: "Move or renumber tasks",
	Long: `Move tasks between prefixes or renumber them.

Examples:
  # Move task tak-11 to cnc-1
  tk mv tak-11 cnc:1

  # Move task tak-11 to cnc, keeping the same number
  tk mv tak-11 cnc --keep-number

  # Move task tak-11 to cnc, auto-assign next available number
  tk mv tak-11 cnc --auto

  # Move multiple tasks
  tk mv tak-11 cnc:1 tk-28 want:1 tk-29 want:2

  # Dry run to see what would happen
  tk mv tak-11 cnc:1 -n`,
	Args: cobra.MinimumNArgs(2),
	RunE: runMvCmd,
}

type moveSpec struct {
	oldID       string
	newPrefix   string
	newNumber   int64
	autoNumber  bool
	keepNumber  bool
	addAlias    bool
	onCollision string // "fail", "auto", "swap"
}

func init() {
	mvCmd.Flags().BoolP("dry-run", "n", false, "Show what would happen without making changes")
	mvCmd.Flags().Bool("alias", true, "Create alias for old ID (default: true)")
	mvCmd.Flags().Bool("no-alias", false, "Don't create alias for old ID")
	mvCmd.Flags().Bool("auto", false, "Auto-assign next available number on collision")
	mvCmd.Flags().Bool("keep-number", false, "Keep the same number in the new prefix")
	mvCmd.Flags().String("on-collision", "fail", "What to do on collision: fail, auto, swap")
}

func runMvCmd(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	aliasFlag, _ := cmd.Flags().GetBool("alias")
	noAliasFlag, _ := cmd.Flags().GetBool("no-alias")
	autoFlag, _ := cmd.Flags().GetBool("auto")
	keepNumberFlag, _ := cmd.Flags().GetBool("keep-number")
	onCollision, _ := cmd.Flags().GetString("on-collision")

	// Parse move specifications from args
	specs, err := parseMoveSpecs(args)
	if err != nil {
		return err
	}

	// Apply global flags to specs
	for i := range specs {
		if noAliasFlag {
			specs[i].addAlias = false
		} else {
			specs[i].addAlias = aliasFlag
		}
		if autoFlag {
			specs[i].autoNumber = true
			// When --auto is set, also set on-collision to auto unless explicitly specified
			if !cmd.Flags().Changed("on-collision") {
				specs[i].onCollision = "auto"
			} else {
				specs[i].onCollision = onCollision
			}
		} else {
			specs[i].onCollision = onCollision
		}
		if keepNumberFlag {
			specs[i].keepNumber = true
		}
	}

	// Validate collision strategy
	if onCollision != "fail" && onCollision != "auto" && onCollision != "swap" {
		return fmt.Errorf("invalid on-collision value: %s (must be fail, auto, or swap)", onCollision)
	}

	db, err := openExistingDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Build reducer to resolve task IDs to UUIDs
	events, err := db.GetEvents()
	if err != nil {
		return err
	}
	reducer, err := BuildFromEvents(events)
	if err != nil {
		return err
	}

	currentUser, err := getCurrentUser()
	if err != nil {
		return err
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return err
	}

	// Process each move specification
	var plan []movePlanEntry
	reserved := make(map[string]struct{}) // Track reserved IDs in this batch
	for _, spec := range specs {
		// Resolve to UUID first (handles aliases and reprefixed tasks)
		taskUUID, err := db.ResolveTaskIDToUUID(spec.oldID)
		if err != nil {
			return fmt.Errorf("failed to resolve task ID %s: %w", spec.oldID, err)
		}

		// Get current task ID from reducer
		task, ok := reducer.GetTask(taskUUID)
		if !ok {
			return fmt.Errorf("task not found: %s", spec.oldID)
		}
		spec.oldID = task.TaskID

		entry, err := planMove(db, reducer, nodeID, spec, reserved)
		if err != nil {
			return fmt.Errorf("failed to plan move for %s: %w", spec.oldID, err)
		}
		plan = append(plan, entry)

		// Mark the new ID as reserved for subsequent plans
		reserved[entry.newID] = struct{}{}
	}

	// Display plan
	displayMovePlan(plan)

	if dryRun {
		fmt.Println("\nDry run - no changes made")
		return nil
	}

	// Execute moves
	for _, entry := range plan {
		if err := executeMove(db, currentUser, entry); err != nil {
			return fmt.Errorf("failed to execute move: %w", err)
		}
	}

	fmt.Printf("\nSuccessfully moved %d task(s)\n", len(plan))
	return nil
}

type movePlanEntry struct {
	oldID     string
	newID     string
	taskUUID  string
	oldPrefix string
	oldNumber int64
	oldNode   string
	newPrefix string
	newNumber int64
	action    string // "move", "renumber", "reprefix"
	note      string
	addAlias  bool
	conflict  bool
}

func parseMoveSpecs(args []string) ([]moveSpec, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("arguments must come in pairs: <old-id> <new-spec>")
	}

	var specs []moveSpec
	for i := 0; i < len(args); i += 2 {
		oldID := args[i]
		newSpec := args[i+1]

		spec := moveSpec{
			oldID:    oldID,
			addAlias: true, // default
		}

		// Parse new spec: "prefix" or "prefix:number"
		parts := strings.Split(newSpec, ":")
		spec.newPrefix = parts[0]

		if len(parts) == 1 {
			// No number specified - will need to determine later
			spec.autoNumber = true
		} else if len(parts) == 2 {
			num, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number in spec %q: %w", newSpec, err)
			}
			spec.newNumber = num
		} else {
			return nil, fmt.Errorf("invalid new spec format: %q (expected prefix or prefix:number)", newSpec)
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

// findCollisionFreeNumber finds a number for the given prefix that doesn't collide
// with any existing task from any node. It starts with the local node's next number
// and increments until it finds a free slot.
func findCollisionFreeNumber(db *DB, reducer *Reducer, prefix string, reserved map[string]struct{}) (int64, error) {
	// Start with the next number from the local node's counter
	nextNum, err := db.GetNextTaskNumberForPrefix(prefix)
	if err != nil {
		return 0, err
	}

	// Get all existing task IDs to check for collisions
	allTasks := reducer.GetAllTasks()
	usedNumbers := make(map[int64]bool)

	// Check which numbers are already used for this prefix (from any node)
	for _, task := range allTasks {
		parts := strings.Split(task.TaskID, "-")
		if len(parts) >= 2 && parts[0] == prefix {
			if num, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				usedNumbers[num] = true
			}
		}
		// Also check aliases
		for _, alias := range task.Aliases {
			parts := strings.Split(alias, "-")
			if len(parts) >= 2 && parts[0] == prefix {
				if num, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					usedNumbers[num] = true
				}
			}
		}
	}

	// Also check reserved numbers from this batch
	for reservedID := range reserved {
		parts := strings.Split(reservedID, "-")
		if len(parts) >= 2 && parts[0] == prefix {
			if num, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				usedNumbers[num] = true
			}
		}
	}

	// Find the first available number starting from nextNum
	candidate := nextNum
	for {
		if !usedNumbers[candidate] {
			return candidate, nil
		}
		candidate++
		// Safety check to prevent infinite loop
		if candidate > nextNum+10000 {
			return 0, fmt.Errorf("could not find available number for prefix %s after checking 10000 candidates", prefix)
		}
	}
}

func planMove(db *DB, reducer *Reducer, nodeID string, spec moveSpec, reserved map[string]struct{}) (movePlanEntry, error) {
	// Resolve old ID to task
	task, ok := reducer.GetTask(spec.oldID)
	if !ok {
		return movePlanEntry{}, fmt.Errorf("task not found: %s", spec.oldID)
	}

	// Parse old ID parts
	parts := strings.Split(task.TaskID, "-")
	if len(parts) < 3 {
		return movePlanEntry{}, fmt.Errorf("invalid task ID format: %s", task.TaskID)
	}

	oldPrefix := parts[0]
	oldNumber, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return movePlanEntry{}, fmt.Errorf("invalid task number in ID %s: %w", task.TaskID, err)
	}
	oldNode := parts[2]

	// Determine new number - handle --keep-number vs --auto precedence
	newNumber := spec.newNumber
	if spec.keepNumber {
		// --keep-number takes precedence
		newNumber = oldNumber
	} else if spec.autoNumber {
		// Get next available number that doesn't collide with any node
		var err error
		newNumber, err = findCollisionFreeNumber(db, reducer, spec.newPrefix, reserved)
		if err != nil {
			return movePlanEntry{}, fmt.Errorf("failed to find collision-free number for prefix %s: %w", spec.newPrefix, err)
		}
	}

	// Construct new ID
	newID := fmt.Sprintf("%s-%d-%s", spec.newPrefix, newNumber, oldNode)

	// Determine action type
	action := "move"
	if oldPrefix != spec.newPrefix && oldNumber == newNumber {
		action = "reprefix"
	} else if oldPrefix == spec.newPrefix && oldNumber != newNumber {
		action = "renumber"
	}

	note := ""
	conflict := false

	// Check for collision (if new ID already exists in reducer or is reserved in this batch)
	_, existsInReducer := reducer.GetTask(newID)
	_, existsInReserved := reserved[newID]
	if existsInReducer || existsInReserved {
		conflict = true
		if existsInReserved {
			note = fmt.Sprintf("collision with earlier move in batch to %s", newID)
		} else {
			note = fmt.Sprintf("collision with existing task %s", newID)
		}
		if spec.onCollision == "auto" {
			// Auto-assign collision-free number
			newNumber, err = findCollisionFreeNumber(db, reducer, spec.newPrefix, reserved)
			if err != nil {
				return movePlanEntry{}, err
			}
			newID = fmt.Sprintf("%s-%d-%s", spec.newPrefix, newNumber, oldNode)
			note = fmt.Sprintf("auto-assigned to %s due to collision", newID)
			conflict = false
		}
	}

	return movePlanEntry{
		oldID:     task.TaskID,
		newID:     newID,
		taskUUID:  task.TaskUUID,
		oldPrefix: oldPrefix,
		oldNumber: oldNumber,
		oldNode:   oldNode,
		newPrefix: spec.newPrefix,
		newNumber: newNumber,
		action:    action,
		note:      note,
		addAlias:  spec.addAlias,
		conflict:  conflict,
	}, nil
}

func displayMovePlan(plan []movePlanEntry) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Old ID", "→", "New ID", "Action", "Note"})
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = false
	t.Style().Options.DrawBorder = false

	for _, entry := range plan {
		arrow := "→"
		if entry.conflict {
			arrow = "✗"
		}
		note := entry.note
		if entry.addAlias {
			if note != "" {
				note += ", alias added"
			} else {
				note = "alias added"
			}
		}
		t.AppendRow(table.Row{entry.oldID, arrow, entry.newID, entry.action, note})
	}

	fmt.Println("\nMove Plan:")
	t.Render()
}

func executeMove(db *DB, currentUser string, entry movePlanEntry) error {
	if entry.conflict {
		return fmt.Errorf("cannot move %s: collision with %s", entry.oldID, entry.newID)
	}

	// Check if target prefix exists
	exists, err := db.PrefixExists(entry.newPrefix)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("target prefix %q does not exist. Create it first with: tk prefix create %s <description>", entry.newPrefix, entry.newPrefix)
	}

	// Generate event ID
	eventID, err := GenerateEventID(db)
	if err != nil {
		return err
	}

	// Get next Lamport timestamp
	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	// Create task.reprefix event
	payload := TaskReprefixPayload{
		TaskUUID:  entry.taskUUID,
		OldPrefix: entry.oldPrefix,
		NewPrefix: entry.newPrefix,
		OldNumber: entry.oldNumber,
		NewNumber: entry.newNumber,
		OldNode:   entry.oldNode,
		Reason:    "user requested move",
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	now := time.Now()
	event := Event{
		ID:        eventID,
		TS:        lamportTS,
		CreatedAt: now,
		Actor:     currentUser,
		Role:      "human",
		Kind:      "task.reprefix",
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	// Add alias if requested
	if entry.addAlias {
		aliasEventID, err := GenerateEventID(db)
		if err != nil {
			return err
		}

		aliasLamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		aliasPayload := TaskAliasAddedPayload{
			TaskUUID: entry.taskUUID,
			AliasID:  entry.oldID,
		}
		aliasPayloadJSON, err := json.Marshal(aliasPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal alias payload: %w", err)
		}

		aliasEvent := Event{
			ID:        aliasEventID,
			TS:        aliasLamportTS,
			CreatedAt: now,
			Actor:     currentUser,
			Role:      "human",
			Kind:      "task.alias.added",
			Payload:   aliasPayloadJSON,
		}

		if err := db.InsertEvent(aliasEvent); err != nil {
			return err
		}
	}

	return nil
}
