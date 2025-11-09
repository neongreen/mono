package cobralint_test

import (
	"testing"

	"github.com/neongreen/mono/linters/cobralint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, cobralint.Analyzer, "a")
}
