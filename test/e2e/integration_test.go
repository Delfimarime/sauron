package e2e

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/delfimarime/sauron/test/e2e/internal/gherkin"
)

// TestFeatures is the godog entrypoint. The suite runs under `go test` (no
// main), so it is invisible to the root module's `go test ./...` across the
// module boundary and is invoked only by the gate-integration task, which
// builds the binary and points SAURON_BIN at it.
func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "sauron",
		ScenarioInitializer: gherkin.RegisterSteps,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"testdata"},
			Output:   colors.Colored(os.Stdout),
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
