//go:build !unit

package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/delfimarime/sauron/test/e2e/internal/gherkin"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime/docker"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime/host"
)

const (
	DisabledSandboxFeatureTag = "@no-sandbox"
	envSauronBin              = "SAURON_BIN"
)

// TestFeatures is the godog entrypoint. The suite runs under `go test` (no main)
// and is the module's default test; the gate-integration task builds the binary,
// points SAURON_BIN at it, and runs `go test ./...`. Strict mode makes undefined
// or pending steps fail, so the suite can never pass without exercising its steps.
func TestFeatures(t *testing.T) {
	binaryURI := os.Getenv(envSauronBin)
	if binaryURI == "" {
		t.Fatalf("%s is not set; the gate-integration task must point it at the built binary", envSauronBin)
	}
	stat, err := os.Stat(binaryURI)
	if err != nil {
		t.Fatalf("stat %s %q: %s", envSauronBin, binaryURI, err.Error())
	}
	if stat.IsDir() {
		t.Fatalf("%s %q must be a file, not a directory", envSauronBin, binaryURI)
	}

	suite := godog.TestSuite{
		Name: "sauron",
		ScenarioInitializer: CreateInitFunc(
			t.TempDir(), binaryURI, gherkin.Init,
		),
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{determineTestdataDirectory(t)},
			Output:   colors.Colored(os.Stdout),
			TestingT: t,
			Strict:   true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

// CreateInitFunc returns a godog ScenarioInitializer that, per scenario, selects
// the runtime by tag (@no-sandbox -> host, otherwise a compose sandbox), attaches
// it to a lazily-started composition handle, and applies each opt to register
// step definitions against that handle. The runtime starts on the first command
// and is stopped after the scenario.
func CreateInitFunc(
	homeDirectory, binaryURI string,
	opts ...func(*godog.ScenarioContext, runtime.Runtime),
) func(*godog.ScenarioContext) {
	hostFactory := &host.Factory{}
	dockerFactory := &docker.Factory{}

	return func(sc *godog.ScenarioContext) {
		cx := &compositionBasedRuntime{}

		sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
			var factory runtime.Factory = dockerFactory
			for _, tag := range scenario.Tags {
				if tag.Name == DisabledSandboxFeatureTag {
					factory = hostFactory
					break
				}
			}
			directory := filepath.Join(homeDirectory, strings.ToLower(scenario.Id))
			r, err := factory.New(binaryURI, directory)
			if err != nil {
				return ctx, fmt.Errorf("create runtime for scenario %q: %w", scenario.Id, err)
			}
			cx.backedBy = r
			return ctx, nil
		})

		sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
			if err := cx.Stop(ctx); err != nil {
				return ctx, fmt.Errorf("stop runtime: %w", err)
			}
			return ctx, nil
		})

		for _, opt := range opts {
			opt(sc, cx)
		}

	}
}
