package dircheck_test

import (
	"testing"

	"github.com/neongreen/mono/linters/dircheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestDebug(t *testing.T) {
	t.Log("Running dircheck analyzer...")
	testdata := analysistest.TestData()
	t.Logf("Testdata dir: %s", testdata)

	// Run without checking results to see what diagnostics are reported
	results := analysistest.Run(t, testdata, dircheck.Analyzer, "a")
	for _, result := range results {
		t.Logf("Result diagnostics: %+v", result.Diagnostics)
	}
}
