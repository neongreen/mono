package main

import (
	"github.com/neongreen/mono/linters/uselesswrapper"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(uselesswrapper.Analyzer)
}
