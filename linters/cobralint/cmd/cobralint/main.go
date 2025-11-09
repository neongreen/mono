package main

import (
	"github.com/neongreen/mono/linters/cobralint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(cobralint.Analyzer)
}
