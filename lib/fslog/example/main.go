package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/neongreen/mono/lib/fslog"
)

func main() {
	// Create a temporary directory for demonstration
	tmpDir, err := os.MkdirTemp("", "fslog-example-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Operating in: %s\n\n", tmpDir)

	// Create a new filesystem
	fs, err := fslog.New(tmpDir, "")
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	// Transaction 1: Create initial configuration
	fmt.Println("=== Transaction 1: Create initial config ===")
	tx1 := fs.Begin(context.Background())
	if err := tx1.WriteFile("config.txt", []byte("version=1\nname=example"), 0644); err != nil {
		log.Fatal(err)
	}
	if err := tx1.Commit(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Created config.txt")

	// Transaction 2: Update configuration
	fmt.Println("\n=== Transaction 2: Update config ===")
	tx2 := fs.Begin(context.Background())
	if err := tx2.WriteFile("config.txt", []byte("version=2\nname=example\nfeature=enabled"), 0644); err != nil {
		log.Fatal(err)
	}
	if err := tx2.Commit(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Updated config.txt")

	// Transaction 3: Add more files
	fmt.Println("\n=== Transaction 3: Add more files ===")
	tx3 := fs.Begin(context.Background())
	if err := tx3.WriteFile("data.txt", []byte("some data"), 0644); err != nil {
		log.Fatal(err)
	}
	if err := tx3.Mkdir("logs", 0755); err != nil {
		log.Fatal(err)
	}
	if err := tx3.WriteFile("logs/app.log", []byte("log entry 1\n"), 0644); err != nil {
		log.Fatal(err)
	}
	if err := tx3.Commit(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Created data.txt, logs/, and logs/app.log")

	// Show operation history
	fmt.Println("\n=== Operation History ===")
	ops, err := fs.History()
	if err != nil {
		log.Fatal(err)
	}

	for _, op := range ops {
		fmt.Printf("Operation %d: %s at %s\n", op.ID, op.Diff(), op.Timestamp.Format("15:04:05"))
	}

	// Demonstrate rollback
	fmt.Println("\n=== Rolling back to state after operation 1 ===")
	if err := fs.RollbackTo(ops[0].ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Rolled back")

	// Show current state
	fmt.Println("\n=== Current filesystem state ===")
	files := []string{"config.txt", "data.txt", "logs/app.log"}
	for _, file := range files {
		path := tmpDir + "/" + file
		if data, err := os.ReadFile(path); err == nil {
			fmt.Printf("  %s: %q\n", file, string(data))
		} else {
			fmt.Printf("  %s: (does not exist)\n", file)
		}
	}

	// Demonstrate transaction rollback
	fmt.Println("\n=== Attempting operation that we'll rollback ===")
	tx4 := fs.Begin(context.Background())
	if err := tx4.WriteFile("temp.txt", []byte("temporary"), 0644); err != nil {
		log.Fatal(err)
	}
	tx4.Rollback()
	fmt.Println("✓ Transaction rolled back, temp.txt not created")

	// Final history
	fmt.Println("\n=== Final Operation History ===")
	ops, err = fs.History()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total operations in log: %d\n", len(ops))
	for _, op := range ops {
		fmt.Printf("  %d: %s\n", op.ID, op.Diff())
	}

	fmt.Println("\n✓ Example completed successfully")
}
