package main

import (
	"testing"
)

func TestCobraCommandsExist(t *testing.T) {
	// Test that root command has the expected subcommands
	if rootCmd == nil {
		t.Fatal("Root command should not be nil")
	}
	
	// Check subcommands exist
	jjSubCmd, _, err := rootCmd.Find([]string{"jj"})
	if err != nil {
		t.Fatalf("jj subcommand should exist: %v", err)
	}
	if jjSubCmd.Use != "jj [config.path] [value]" {
		t.Errorf("jj command has wrong usage: %s", jjSubCmd.Use)
	}
	
	miseSubCmd, _, err := rootCmd.Find([]string{"mise"})
	if err != nil {
		t.Fatalf("mise subcommand should exist: %v", err)
	}
	if miseSubCmd.Use != "mise [config.path] [value]" {
		t.Errorf("mise command has wrong usage: %s", miseSubCmd.Use)
	}
	
	completionSubCmd, _, err := rootCmd.Find([]string{"completion"})
	if err != nil {
		t.Fatalf("completion subcommand should exist: %v", err)
	}
	if completionSubCmd.Use != "completion [bash|zsh|fish]" {
		t.Errorf("completion command has wrong usage: %s", completionSubCmd.Use)
	}
}

func TestCommandStructure(t *testing.T) {
	if rootCmd.Use != "conf" {
		t.Errorf("Root command should be 'conf', got: %s", rootCmd.Use)
	}
	
	if len(rootCmd.Commands()) < 3 {
		t.Errorf("Root command should have at least 3 subcommands, got: %d", len(rootCmd.Commands()))
	}
}
