// +build js,wasm

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall/js"

	"github.com/neongreen/mono/tk/cmd"
)

// executeCommand runs a tk command and captures its output
func executeCommand(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"success": false,
			"stdout":  "",
			"stderr":  "No command provided",
		}
	}

	commandString := args[0].String()
	commandArgs := strings.Fields(commandString)

	if len(commandArgs) == 0 {
		return map[string]interface{}{
			"success": false,
			"stdout":  "",
			"stderr":  "Empty command",
		}
	}

	// Save old Args
	oldArgs := os.Args

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}

	// Set custom output writers for Cobra
	cmd.RootCmd.SetOut(stdoutBuf)
	cmd.RootCmd.SetErr(stderrBuf)

	os.Args = append([]string{"tk"}, commandArgs...)

	// Reset the root command to allow multiple executions
	cmd.RootCmd.SetArgs(commandArgs)

	// Execute the command
	cmdErr := cmd.RootCmd.Execute()

	// Restore
	os.Args = oldArgs
	cmd.RootCmd.SetOut(os.Stdout)
	cmd.RootCmd.SetErr(os.Stderr)

	success := cmdErr == nil

	return map[string]interface{}{
		"success": success,
		"stdout":  stdoutBuf.String(),
		"stderr":  stderrBuf.String(),
	}
}

// initDB initializes the tk database in browser storage
func initDB(this js.Value, args []js.Value) interface{} {
	// Use in-memory database for WASM demo
	os.Setenv("TK_DB_PATH", ":memory:")

	// Run tk init
	return executeCommand(this, []js.Value{js.ValueOf("init")})
}

func main() {
	fmt.Println("tk WASM module loaded")

	// Set up environment for WASM
	os.Setenv("HOME", ".")
	os.Setenv("USER", "wasm-demo")
	// Use in-memory database for WASM demo
	os.Setenv("TK_DB_PATH", ":memory:")

	// Register JavaScript functions
	js.Global().Set("tkExecute", js.FuncOf(executeCommand))
	js.Global().Set("tkInit", js.FuncOf(initDB))

	// Keep the Go program running
	<-make(chan struct{})
}
