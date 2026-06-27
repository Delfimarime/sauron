package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shared literals the unset tests assert against.
const (
	docNameAcme  = "source: git@github.com:acme/artifacts.git"
	unknownFlag  = "--nope"
	summaryUnset = "registry unset; installed artifacts preserved"
	oneRegistry  = authRegistries
)

// runUnsetRegistry assembles and runs the subcommand, returning stdout and the
// resulting error.
func runUnsetRegistry(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := UnsetRegistry()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// readSettings reads the seeded settings.yaml back from SAURON_HOME.
func readSettings(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(os.Getenv("SAURON_HOME"), settingsFile))
	if os.IsNotExist(err) {
		return ""
	}
	require.NoError(t, err)
	return string(data)
}

// TestUnsetGroup asserts the unset group has no run behaviour and attaches the
// registry subcommand.
func TestUnsetGroup(t *testing.T) {
	// Arrange + Act.
	cmd := Unset()

	// Assert.
	assert.Equal(t, "unset", cmd.Name())
	assert.Nil(t, cmd.RunE, "the group has no run behaviour")

	var registry *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == subcmdRegistry {
			registry = sub
		}
	}
	require.NotNil(t, registry, "the registry subcommand is attached")
}

// TestUnsetRegistryFlagSurface asserts the --dry-run flag and the argument
// validator are installed.
func TestUnsetRegistryFlagSurface(t *testing.T) {
	// Arrange + Act.
	cmd := UnsetRegistry()

	// Assert.
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"), "the dry-run flag is registered")
	assert.NotNil(t, cmd.Args, "an argument validator is installed")
}

// TestUnsetRegistryRejectsBadInput asserts an unexpected argument or an unknown
// flag is rejected and maps to the usage exit code.
func TestUnsetRegistryRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: caseUnexpectedArg, args: []string{argExtra}},
		{name: "rejects an unknown unset flag", args: []string{unknownFlag}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			_, err := runUnsetRegistry(t, tt.args...)

			// Assert.
			require.Error(t, err)
			assert.Equal(t, exitUsage, exitCode(err))
		})
	}
}

// TestUnsetRegistryEndToEnd drives the assembled subcommand through the real fx
// graph against a seeded settings.yaml, covering removal, the no-registry
// success, and the dry-run preview.
func TestUnsetRegistryEndToEnd(t *testing.T) {
	tests := []struct {
		name       string
		seed       string
		args       []string
		wantOut    []string
		wantAbsent []string // substrings the state must no longer contain
		wantState  []string // substrings the state must still contain
	}{
		{
			name:       "removes the registry and reports artifacts preserved",
			seed:       oneRegistry,
			wantOut:    []string{summaryUnset},
			wantAbsent: []string{docNameAcme},
		},
		{
			name:    "no registry reports nothing was unset",
			seed:    "",
			wantOut: []string{"nothing was unset"},
		},
		{
			name:      "dry-run previews without changing state",
			seed:      oneRegistry,
			args:      []string{"--dry-run"},
			wantOut:   []string{"would be unset"},
			wantState: []string{docNameAcme},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			seedRegistries(t, tt.seed)

			// Act.
			out, err := runUnsetRegistry(t, tt.args...)

			// Assert.
			require.NoError(t, err)
			for _, want := range tt.wantOut {
				assert.Contains(t, out, want)
			}
			state := readSettings(t)
			for _, want := range tt.wantState {
				assert.Contains(t, state, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, state, absent)
			}
		})
	}
}
