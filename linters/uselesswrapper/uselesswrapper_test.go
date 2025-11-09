package uselesswrapper_test

import (
	"testing"

	"github.com/neongreen/mono/linters/uselesswrapper"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, uselesswrapper.Analyzer, "a")
}
