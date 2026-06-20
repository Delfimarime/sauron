package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

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
			// A caller opt that overrides the filesystem with an in-memory one
			// must compose without breaking the graph.
			name: "caller opt is appended",
			opts: func() []fx.Option {
				return []fx.Option{
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

// TestBindFlags verifies each shared flag group registers its flags with the
// documented defaults and binds them to its struct.
func TestBindFlags(t *testing.T) {
	// Arrange.
	cmd := &cobra.Command{Use: "x"}
	var listing listingFlags
	var paging pagingFlags
	var dry dryRunFlags
	var timeout timeoutFlags
	var kind kindFlags

	// Act.
	bindListingFlags(cmd, &listing)
	bindPagingFlags(cmd, &paging)
	bindDryRunFlags(cmd, &dry)
	bindTimeoutFlags(cmd, &timeout)
	bindKindFlags(cmd, &kind)

	// Assert: defaults bound onto the structs.
	assert.Equal(t, "", listing.Search)
	assert.Equal(t, "", listing.Sort)
	assert.Equal(t, "asc", listing.Order)
	assert.Empty(t, listing.Fields)
	assert.Equal(t, 0, paging.Offset)
	assert.Equal(t, 0, paging.Limit)
	assert.False(t, dry.DryRun)
	assert.Equal(t, 30*time.Second, timeout.Timeout)
	assert.Equal(t, kindHTTP, kind.Kind)

	// Assert: flags are registered on the command.
	for _, name := range []string{"search", "sort", "order", "fields", "offset", "limit", "dry-run", "timeout", "kind"} {
		assert.NotNilf(t, cmd.Flags().Lookup(name), "flag %q registered", name)
	}
}
