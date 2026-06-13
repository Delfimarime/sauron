package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew exercises the root command builder: the version banner content and the
// home-resolution success and failure paths. It follows the Serve()/serve()
// testability spirit — New is constructed with arbitrary identity strings and
// its output is asserted without a real process.
func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		appName    string
		appVersion string
		appHash    string
		// home, when non-empty, is set as $SAURON_HOME for the case.
		home string
		// clearUserHome, when true, forces os.UserHomeDir to fail by clearing
		// HOME so the default-home path errors.
		clearUserHome bool
		wantErr       bool
		// wantBanner is the exact banner expected on success.
		wantBanner string
	}{
		{
			// Success with an explicit $SAURON_HOME: the banner reflects it.
			name:       "explicit home",
			appName:    "sauron",
			appVersion: "1.2.3",
			appHash:    "abc1234",
			home:       "/tmp/sauron-home",
			wantBanner: "sauron v1.2.3\nHash abc1234\nHome: /tmp/sauron-home\n",
		},
		{
			// Failure: no $SAURON_HOME and os.UserHomeDir cannot resolve.
			name:          "home resolution fails",
			appName:       "sauron",
			appVersion:    "0.0.0",
			appHash:       "deadbee",
			clearUserHome: true,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			if tt.home != "" {
				t.Setenv("SAURON_HOME", tt.home)
			} else {
				t.Setenv("SAURON_HOME", "")
			}
			if tt.clearUserHome {
				clearHomeEnv(t)
			}

			// Act.
			root, err := New(tt.appName, tt.appVersion, tt.appHash)

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, root)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, root)

			var out bytes.Buffer
			root.SetOut(&out)
			root.SetArgs(nil)
			require.NoError(t, root.Execute())
			assert.Equal(t, tt.wantBanner, out.String())
		})
	}
}

// TestNewVersionFlag verifies --version emits the same banner as the bare root
// command.
func TestNewVersionFlag(t *testing.T) {
	// Arrange.
	t.Setenv("SAURON_HOME", "/var/lib/sauron")
	root, err := New("sauron", "9.9.9", "cafef00d")
	require.NoError(t, err)

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--version"})

	// Act.
	require.NoError(t, root.Execute())

	// Assert.
	assert.Equal(t, "sauron v9.9.9\nHash cafef00d\nHome: /var/lib/sauron\n", out.String())
}

// clearHomeEnv removes every variable os.UserHomeDir consults so it returns an
// error, forcing config.ResolveHome's failure branch. Values are restored by
// t.Setenv/t.Cleanup.
func clearHomeEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", "")
	// On non-Unix platforms os.UserHomeDir consults other variables; clearing
	// them keeps the test deterministic across targets.
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
}
