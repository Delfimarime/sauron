package storage

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/delfimarime/sauron/internal/config"
)

// TestNewStore asserts the store retains the injected filesystem.
func TestNewStore(t *testing.T) {
	// Arrange.
	fs := afero.NewMemMapFs()

	// Act.
	store := NewStore(fs)

	// Assert.
	require.NotNil(t, store)
	assert.Same(t, fs, store.fs)
}

// TestNewFxOptions resolves a Store through the container to exercise the wiring.
func TestNewFxOptions(t *testing.T) {
	// Arrange + Act.
	var store *Store
	app := fx.New(
		fx.Supply(config.Configuration{HomeDirectory: t.TempDir()}),
		NewFxOptions(),
		fx.Populate(&store),
	)

	// Assert.
	require.NoError(t, app.Err())
	require.NotNil(t, store.fs)
}

// TestNewFilesystem asserts paths resolve under the configured home, without touching the real filesystem.
func TestNewFilesystem(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// probe is the path resolved through the returned fs.
		probe string
	}{
		// A flat path resolves directly under home.
		{name: "roots a file under home", probe: "registries.yaml"},
		// A nested path resolves under home too.
		{name: "roots a nested path under home", probe: "sub/track.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: t.TempDir is used only as a home path string; nothing is written.
			home := t.TempDir()
			fs := newFilesystem(config.Configuration{HomeDirectory: home})

			// Act: RealPath resolves the path without any I/O.
			base, ok := fs.(*afero.BasePathFs)
			require.True(t, ok)
			real, err := base.RealPath(tt.probe)

			// Assert: the resolved path sits under the configured home.
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(home, tt.probe), real)
		})
	}
}
