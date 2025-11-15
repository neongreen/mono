package main

import (
	"github.com/neongreen/mono/linters/dircheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(dircheck.Analyzer)
}
