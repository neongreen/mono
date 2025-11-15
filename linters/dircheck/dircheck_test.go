package dircheck_test

import (
	"testing"

	"github.com/neongreen/mono/linters/dircheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, dircheck.Analyzer, "a")
}
