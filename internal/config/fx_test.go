package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// TestResolveHome covers home resolution: the $SAURON_HOME override, the
// platform default, and the failure when no home can be derived.
func TestResolveHome(t *testing.T) {
	tests := []struct {
		name string
		// envHome, when non-empty, is exported as $SAURON_HOME.
		envHome string
		// clearHome forces os.UserHomeDir to fail by clearing its inputs.
		clearHome bool
		wantErr   bool
		// want, when set, is the exact resolved home.
		want string
		// wantUserDefault asserts the default ~/.sauron form instead of an exact value.
		wantUserDefault bool
	}{
		{
			// $SAURON_HOME wins verbatim when set.
			name:    "explicit env override",
			envHome: "/custom/sauron",
			want:    "/custom/sauron",
		},
		{
			// Unset env falls back to <user-home>/.sauron.
			name:            "platform default",
			wantUserDefault: true,
		},
		{
			// No env and no derivable user home is an error.
			name:      "no home derivable",
			clearHome: true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			t.Setenv(envHome, tt.envHome)
			if tt.clearHome {
				t.Setenv("HOME", "")
				t.Setenv("USERPROFILE", "")
				t.Setenv("HOMEDRIVE", "")
				t.Setenv("HOMEPATH", "")
			}

			// Act.
			got, err := GetHomeDirectory()

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantUserDefault {
				assert.True(t, filepath.IsAbs(got))
				assert.Equal(t, defaultHomeDir, filepath.Base(got))
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNew resolves both homes on success and surfaces the failures of each
// resolution step.
func TestNew(t *testing.T) {
	t.Run("resolves both homes", func(t *testing.T) {
		// Arrange.
		home := t.TempDir()
		t.Setenv(envHome, home)

		// Act.
		cfg, err := New()

		// Assert.
		require.NoError(t, err)
		assert.Equal(t, home, cfg.HomeDirectory)
		assert.True(t, filepath.IsAbs(cfg.UserHomeDirectory))
	})

	t.Run("fails when no sauron home is derivable", func(t *testing.T) {
		// Arrange: no override and no derivable user home.
		t.Setenv(envHome, "")
		clearUserHome(t)

		// Act.
		_, err := New()

		// Assert.
		require.Error(t, err)
	})

	t.Run("fails when the user home is not derivable", func(t *testing.T) {
		// Arrange: SAURON_HOME resolves, but os.UserHomeDir cannot.
		t.Setenv(envHome, t.TempDir())
		clearUserHome(t)

		// Act.
		_, err := New()

		// Assert.
		require.Error(t, err)
	})
}

// clearUserHome clears the inputs os.UserHomeDir reads, so it fails.
func clearUserHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
}

// TestNewFxOptions verifies the fx option provides a Configuration carrying the
// resolved home and that *viper.Viper is never required to satisfy it.
func TestNewFxOptions(t *testing.T) {
	// Arrange.
	home := t.TempDir()
	t.Setenv(envHome, home)

	var cfg Configuration

	// Act.
	app := fx.New(NewFxOptions(), fx.Populate(&cfg))

	// Assert.
	require.NoError(t, app.Err())
	assert.Equal(t, home, cfg.HomeDirectory)
}
