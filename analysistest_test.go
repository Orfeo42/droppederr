package droppederr

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Run("an if-block that never references err is DROPPED, one that does is silent", func(t *testing.T) {
		analysistest.Run(t, analysistest.TestData(), New(Config{}), "dropped")
	})

	t.Run("a file matched by exclude-paths reports nothing", func(t *testing.T) {
		analysistest.Run(t, analysistest.TestData(), New(Config{ExcludePaths: "src/excluded/"}), "excluded")
	})
}
