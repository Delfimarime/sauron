//go:build !unit

package e2e

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime/docker"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime/host"
)

const (
	DisabledSandboxFeatureTag = "@no-sandbox"
	directoryKey              = "@integration.directory"
	runtimeFactoryKey         = "@integration.runtime.factory"
	runtimeInstanceKey        = "@integration.runtime.instance"
)

// TestFeatures is the godog entrypoint. The suite runs under `go test` (no
// main), so it is invisible to the root module's `go test ./...` across the
// module boundary and is invoked only by the gate-integration task, which
// builds the binary and points SAURON_BIN at it.
func TestFeatures(t *testing.T) {
	stat, err := os.Stat("../../dist/app")
	if err != nil {
		t.Fatalf("unable to stat dist/app.\ncaused by:%s", err.Error())
	}
	if stat.IsDir() {
		t.Fatal("dist/app must be a file not directory")
	}
	suite := godog.TestSuite{
		Name: "sauron",
		ScenarioInitializer: CreateInitFunc(
			t.TempDir(), "dist/app", //TODO add values here
		),
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

func CreateInitFunc(homeDirectory, binaryURI string, opts ...func(*godog.ScenarioContext, runtime.Runtime)) func(*godog.ScenarioContext) {
	hostRuntimeFactory := &host.Factory{}
	dockerRuntimeFactory := &docker.Factory{}

	return func(sc *godog.ScenarioContext) {
		if len(opts) == 0 {
			return
		}
		isStartedUp := false

		sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
			var f runtime.Factory

			for _, tag := range scenario.Tags {
				if tag.Name == DisabledSandboxFeatureTag {
					f = hostRuntimeFactory
					break
				}
			}
			if f == nil {
				f = dockerRuntimeFactory
			}

			directory := fmt.Sprintf("%s/%s", homeDirectory, strings.ToLower(scenario.Id))
			c := context.WithValue(ctx, runtimeFactoryKey, f)

			if err := os.Mkdir(directory, 0600); err != nil {
				return c, fmt.Errorf("unable to create scenario directory")
			}
			c = context.WithValue(c, directoryKey, directory)
			return c, nil
		})

		sc.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
			if !isStartedUp && strings.HasPrefix(st.Text, "When") {
				f, hasValue := ctx.Value(runtimeFactoryKey).(runtime.Factory)
				if !hasValue {
					return ctx, fmt.Errorf("An unexpected error when attempting to retrieve the runtime factory")
				}
				r, err := f.New(binaryURI, ctx.Value(directoryKey).(string))
				if err != nil {
					return ctx, fmt.Errorf("unable to create runtime.\ncaused by:%w", err)
				}
				if err := r.Start(ctx); err != nil {
					return ctx, err
				}
				isStartedUp = true
				return context.WithValue(ctx, runtimeInstanceKey, r), nil
			}
			return ctx, nil
		})

		sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
			r, hasValue := ctx.Value(runtimeInstanceKey).(runtime.Runtime)
			if hasValue {
				if err := r.Stop(ctx); err != nil {
					return ctx, fmt.Errorf("An unexpected error occured while attempting to stop runtime. caused by: %w", err)
				}
			}
			return ctx, nil
		})

		for _, opt := range opts {
			opt(sc, nil)
		}

	}
}
