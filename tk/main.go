package main

import (
	"os"

	"github.com/neongreen/mono/tk/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
