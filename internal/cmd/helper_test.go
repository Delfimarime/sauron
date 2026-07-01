package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/delfimarime/sauron/internal/infrastructure/repository"
	"github.com/delfimarime/sauron/internal/usecase"
)

// shared `registry` command-test literals, named to satisfy goconst across the
// package's set/describe/unset/catalogue command tests.
const (
	subcmdRegistry    = "registry"
	subcmdProvider    = "provider"
	nameClaude        = "claude"
	nameBogus         = "bogus"
	acmeName          = "acme"
	settingsFile      = "settings.yaml"
	caseUnexpectedArg = "rejects an unexpected argument"
	// versionOne is the fixture version shared across artifact test setups.
	versionOne = "1.0.0"
)

// seedRegistries pins SAURON_HOME to a fresh temp dir and writes stream as the
// settings.yaml state, so the registry-reading commands resolve it without
// touching anything durable. An empty stream leaves no registry configured.
func seedRegistries(t *testing.T, stream string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("SAURON_HOME", home)
	if stream == "" {
		return
	}
	require.NoError(t, os.WriteFile(filepath.Join(home, settingsFile), []byte(stream), 0o644))
}

// artifactSummary mirrors the Sauron HTTP Registry API's ArtifactSummary: the
// condensed artifact view the http transport's marketplace client decodes from a
// collection listing.
type artifactSummary struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    int64  `json:"size"`
}

// startHTTPRegistry stands up an in-process httptest.Server implementing the
// minimal Sauron HTTP Registry API the http transport consumes: GET /skills and
// GET /agents answer with the supplied summaries wrapped in an ArtifactList. The
// server is closed when the test ends, keeping the test offline and self-contained.
func startHTTPRegistry(t *testing.T, skills, agents []artifactSummary) string {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/skills", listArtifacts(skills))
	mux.HandleFunc("/agents", listArtifacts(agents))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return srv.URL
}

// listArtifacts answers a collection listing with the given summaries wrapped in
// an ArtifactList body.
func listArtifacts(items []artifactSummary) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Items []artifactSummary `json:"items"`
		}{Items: items})
	}
}

// closedHTTPRegistry returns the URL of an httptest.Server that has already been
// closed: the URL is well-formed but refuses connections, so opening it fails as
// a runtime (unreachable) error rather than a usage error.
func closedHTTPRegistry(t *testing.T) string {
	t.Helper()

	srv := httptest.NewServer(http.NewServeMux())
	url := srv.URL
	srv.Close()

	return url
}

// TestNewApp asserts the transversal fx graph wires and validates cleanly — the
// stubs satisfy the container without panicking — and that the caller's opts are
// appended. It builds the app (NewApp does not start it) and validates it.
func TestNewApp(t *testing.T) {
	tests := []struct {
		name string
		// extra opts the caller appends; nil means none.
		opts func() []fx.Option
		// wantErr is true when the graph is expected to fail validation.
		wantErr bool
	}{
		{
			// The transversal graph alone must validate: every provided
			// constructor resolves and no required dependency is missing.
			name:    "transversal graph validates",
			opts:    func() []fx.Option { return nil },
			wantErr: false,
		},
		{
			// The command-owned opts (repository + usecase) compose onto the
			// transversal graph, and a caller may still decorate the storage
			// filesystem with an in-memory one.
			name: "command opts are appended",
			opts: func() []fx.Option {
				return []fx.Option{
					repository.NewFxOptions(),
					usecase.NewFxOptions(),
					fx.Decorate(func(afero.Fs) afero.Fs { return afero.NewMemMapFs() }),
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			t.Setenv("SAURON_HOME", t.TempDir())
			ctx := context.Background()

			// Act.
			app := NewApp(ctx, tt.opts()...)

			// Assert.
			require.NotNil(t, app)
			err := app.Err()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

// TestNewAppLifecycle starts and stops the app to exercise the pond pool's
// lifecycle hook (OnStop stops and waits for the pool).
func TestNewAppLifecycle(t *testing.T) {
	// Arrange.
	t.Setenv("SAURON_HOME", t.TempDir())
	app := NewApp(context.Background())
	require.NoError(t, app.Err())

	startCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Act + Assert.
	require.NoError(t, app.Start(startCtx))
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	assert.NoError(t, app.Stop(stopCtx))
}

// TestNewAppCommandGraph asserts the set-registry use case resolves when its
// command-owned opts (repository + usecase) are appended to the transversal
// graph — the wiring the command itself performs.
func TestNewAppCommandGraph(t *testing.T) {
	// Arrange.
	t.Setenv("SAURON_HOME", t.TempDir())
	var uc usecase.UseCase[usecase.SetRegistryRequest, usecase.SetRegistryResponse]
	app := NewApp(context.Background(),
		repository.NewFxOptions(),
		usecase.NewFxOptions(),
		fx.Decorate(func(afero.Fs) afero.Fs { return afero.NewMemMapFs() }),
		fx.Populate(&uc),
	)
	require.NoError(t, app.Err())

	startCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Act.
	require.NoError(t, app.Start(startCtx))
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		assert.NoError(t, app.Stop(stopCtx))
	}()

	// Assert.
	assert.NotNil(t, uc)
}

// TestProvidePool runs the pool through fx, submits a task, then stops the app to exercise the OnStop hook.
func TestProvidePool(t *testing.T) {
	// Arrange.
	var pool pond.Pool
	app := fx.New(
		fx.Provide(
			func() context.Context {
				return context.Background()
			},
		),
		fx.Provide(newPondPool), fx.Populate(&pool),
	)
	require.NoError(t, app.Err())
	require.NotNil(t, pool)

	startCtx, startCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startCancel()
	require.NoError(t, app.Start(startCtx))

	ran := make(chan struct{}, 1)
	pool.Submit(func() { ran <- struct{}{} })

	// Act + Assert: the submitted task runs on the pool.
	select {
	case <-ran:
	case <-time.After(5 * time.Second):
		t.Fatal("pool task did not run")
	}

	// Act + Assert: stopping the app stops and waits for the pool.
	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	assert.NoError(t, app.Stop(stopCtx))
	assert.True(t, pool.Stopped())
}

// TestNewCommand exercises the shared command scaffold every builder (leaf or
// group) is built from: use/short are positional (every command needs both,
// no exceptions), everything else — long, args policy, flags, the handler, or
// subcommands — is an option, individually omittable, so one constructor
// serves both a leaf command and a pure group.
func TestNewCommand(t *testing.T) {
	t.Run("wires the scaffold with no options", func(t *testing.T) {
		// Act.
		cmd := newCommand("widget", "short")

		// Assert: no positional-args restriction unless withArgs is given (a group's
		// shape), silencing is always on, and there is no RunE.
		assert.Equal(t, "widget", cmd.Use)
		assert.Equal(t, "short", cmd.Short)
		assert.Nil(t, cmd.Args)
		assert.True(t, cmd.SilenceUsage)
		assert.True(t, cmd.SilenceErrors)
		assert.Nil(t, cmd.RunE)
	})

	t.Run("withLong sets the long description", func(t *testing.T) {
		// Act.
		cmd := newCommand("widget", "short", withLong("long"))

		// Assert.
		assert.Equal(t, "long", cmd.Long)
	})

	t.Run("withSubcommands attaches children — a pure group, no RunE", func(t *testing.T) {
		// Arrange.
		child := newCommand("child", "short")

		// Act.
		cmd := newCommand("widget", "short", withSubcommands(child))

		// Assert.
		assert.Nil(t, cmd.RunE)
		assert.Len(t, cmd.Commands(), 1)
		assert.Equal(t, "child", cmd.Commands()[0].Use)
	})

	t.Run("withArgs, withFlags, and withRunE wire a leaf command", func(t *testing.T) {
		// Arrange.
		var boundFlag string
		var gotCtx context.Context
		var gotArgs []string
		var gotOut io.Writer
		cmd := newCommand("widget", "short",
			withArgs(cobra.MaximumNArgs(1)),
			withFlags(func(c *cobra.Command) {
				c.Flags().StringVar(&boundFlag, "extra", "", "an extra flag bind wires")
			}),
			withRunE(func(ctx context.Context, args []string, stdout io.Writer) error {
				gotCtx, gotArgs, gotOut = ctx, args, stdout
				return nil
			}),
		)
		assert.NotNil(t, cmd.Flags().Lookup("extra"), "withFlags wired the flag")

		// Act.
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetArgs([]string{"one"})
		require.NoError(t, cmd.Execute())

		// Assert: run received the positional args, a live context, and stdout.
		assert.Equal(t, []string{"one"}, gotArgs)
		require.NotNil(t, gotCtx)
		assert.Equal(t, &out, gotOut)
	})

	t.Run("withArgs rejects an argument the policy disallows", func(t *testing.T) {
		// Arrange.
		cmd := newCommand("widget", "short",
			withArgs(cobra.NoArgs),
			withRunE(func(context.Context, []string, io.Writer) error { return nil }),
		)
		cmd.SetArgs([]string{argExtra})

		// Act.
		err := cmd.Execute()

		// Assert.
		require.Error(t, err)
	})
}

// fakeUseCase is a trivial type runUseCase can resolve from a supplied fx.Option,
// letting the helper be exercised without depending on a real use case's graph.
type fakeUseCase struct{}

// TestRunUseCase exercises the shared build->start->cancel->stop helper: the
// success path resolves the use case and surfaces exec's error verbatim, and a
// build failure surfaces the wrapped "build application" message.
func TestRunUseCase(t *testing.T) {
	t.Run("resolves use case and returns exec error", func(t *testing.T) {
		// Arrange.
		t.Setenv("SAURON_HOME", t.TempDir())
		wantErr := errors.New("boom")
		var (
			resolved   bool
			liveAtExec bool
		)
		provide := fx.Provide(func() *fakeUseCase { return &fakeUseCase{} })

		// Act.
		_, err := runUseCase(context.Background(),
			func(runCtx context.Context, uc *fakeUseCase) (*fakeUseCase, error) {
				resolved = uc != nil
				liveAtExec = runCtx.Err() == nil
				return nil, wantErr
			},
			provide,
		)

		// Assert: exec's error is surfaced verbatim and it ran with a live context.
		assert.ErrorIs(t, err, wantErr)
		assert.True(t, resolved, "use case resolved from supplied opt")
		assert.True(t, liveAtExec, "exec receives a still-live run context")
	})

	t.Run("wraps a build failure", func(t *testing.T) {
		// Arrange: no provider for *fakeUseCase, so the fx graph fails to build.
		t.Setenv("SAURON_HOME", t.TempDir())
		var ran bool

		// Act: runUseCase appends repository+usecase, but *fakeUseCase is unprovided.
		_, err := runUseCase(context.Background(),
			func(context.Context, *fakeUseCase) (*fakeUseCase, error) {
				ran = true
				return nil, nil
			},
		)

		// Assert: the build failure is wrapped and exec never ran.
		require.Error(t, err)
		assert.Contains(t, err.Error(), "build application")
		assert.False(t, ran, "exec must not run when the graph fails to build")
	})
}
