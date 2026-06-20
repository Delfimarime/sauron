package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

func TestOSFactory_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []extension.Option
		wantErr bool
	}{
		{
			name: "bare uri is accepted",
			opts: []extension.Option{extension.WithURI("/srv/registry")},
		},
		{
			name:    "reference is rejected",
			opts:    []extension.Option{extension.WithURI("/srv/registry"), extension.WithRef("main")},
			wantErr: true,
		},
		{
			name:    "credentials are rejected",
			opts:    []extension.Option{extension.WithURI("/srv/registry"), extension.WithBasicAuth("u", "p")},
			wantErr: true,
		},
		{
			name:    "transport security is rejected",
			opts:    []extension.Option{extension.WithURI("/srv/registry"), extension.WithSkipTLSVerify(true)},
			wantErr: true,
		},
		{
			name:    "ca certificate is rejected",
			opts:    []extension.Option{extension.WithURI("/srv/registry"), extension.WithCACert("/ca.pem")},
			wantErr: true,
		},
		{
			name:    "ssh key is rejected",
			opts:    []extension.Option{extension.WithURI("/srv/registry"), extension.WithSSHKey("/id_ed25519")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			err := newOSFactory().Validate(tt.opts...)

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, api.ErrUsage)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestOSFactory_Open(t *testing.T) {
	t.Parallel()

	// Arrange: a real isolated directory holding a .skills collection.
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".skills"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".skills", "a.yaml"), []byte("x"), 0o644))

	// Act.
	fs, err := newOSFactory().Open(context.Background(), extension.WithURI(root))

	// Assert.
	require.NoError(t, err)
	files, listErr := fs.List(context.Background(), ".skills", source.WithLimit(1))
	require.NoError(t, listErr)
	require.Len(t, files, 1)
	assert.Equal(t, "a.yaml", files[0].Name())
}

func TestOSFactory_Open_Errors(t *testing.T) {
	t.Parallel()

	t.Run("missing uri is a runtime error", func(t *testing.T) {
		t.Parallel()

		// Act.
		_, err := newOSFactory().Open(context.Background(), extension.WithURI(filepath.Join(t.TempDir(), "absent")))

		// Assert.
		require.Error(t, err)
		assert.ErrorIs(t, err, api.ErrRuntime)
	})

	t.Run("file uri is rejected", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		file := filepath.Join(t.TempDir(), "registry")
		require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))

		// Act.
		_, err := newOSFactory().Open(context.Background(), extension.WithURI(file))

		// Assert.
		require.Error(t, err)
		assert.ErrorIs(t, err, api.ErrRuntime)
	})

	t.Run("invalid options surface before access", func(t *testing.T) {
		t.Parallel()

		// Act.
		_, err := newOSFactory().Open(context.Background(),
			extension.WithURI(t.TempDir()), extension.WithRef("main"))

		// Assert.
		require.Error(t, err)
		assert.ErrorIs(t, err, api.ErrUsage)
	})
}
