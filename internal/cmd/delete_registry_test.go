package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shared literals the delete tests assert against, hoisted so a repeated value is
// not spelled out across cases.
const (
	internalName  = "internal"
	docNameAcme   = "name: acme"
	unknownFlag   = "--nope"
	summaryRemove = `registry "acme" removed`
)

// runDeleteRegistry assembles and runs the subcommand, returning stdout and the
// resulting error.
func runDeleteRegistry(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := DeleteRegistry()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// readRegistries reads the seeded registries.yaml back from SAURON_HOME.
func readRegistries(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(os.Getenv("SAURON_HOME"), "registries.yaml"))
	if os.IsNotExist(err) {
		return ""
	}
	require.NoError(t, err)
	return string(data)
}

// TestNewDeleteRegistryRequestMapsArgs asserts the positional name and the
// --dry-run flag land on the use case request.
func TestNewDeleteRegistryRequestMapsArgs(t *testing.T) {
	// Arrange.
	var stdout bytes.Buffer
	flags := deleteRegistryFlags{dryRunFlags: dryRunFlags{DryRun: true}}

	// Act.
	request := newDeleteRegistryRequest(context.Background(), &flags, []string{acmeName}, &stdout)

	// Assert.
	require.NotNil(t, request)
	assert.Equal(t, acmeName, request.Name)
	assert.True(t, request.DryRun)
	assert.Same(t, &stdout, request.Out())
}

// TestDeleteGroup asserts the delete group has no run behaviour and attaches the
// registry subcommand.
func TestDeleteGroup(t *testing.T) {
	// Arrange + Act.
	cmd := Delete()

	// Assert.
	assert.Equal(t, "delete", cmd.Name())
	assert.Nil(t, cmd.RunE, "the group has no run behaviour")

	var registry *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == subcmdRegistry {
			registry = sub
		}
	}
	require.NotNil(t, registry, "the registry subcommand is attached")
}

// TestDeleteRegistryFlagSurface asserts the --dry-run flag and the argument
// validator are installed.
func TestDeleteRegistryFlagSurface(t *testing.T) {
	// Arrange + Act.
	cmd := DeleteRegistry()

	// Assert.
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"), "the dry-run flag is registered")
	assert.NotNil(t, cmd.Args, "an argument validator is installed")
}

// TestDeleteRegistryRejectsBadInput asserts a missing name or an unknown flag is
// rejected and maps to the usage exit code (FR-006).
func TestDeleteRegistryRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "rejects a missing name", args: nil},
		{name: "rejects an extra positional", args: []string{"a", "b"}},
		{name: "rejects an unknown delete flag", args: []string{acmeName, unknownFlag}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			_, err := runDeleteRegistry(t, tt.args...)

			// Assert.
			require.Error(t, err)
			assert.Equal(t, exitUsage, exitCode(err))
		})
	}
}

// TestDeleteRegistryEndToEnd drives the assembled subcommand through the real fx
// graph against a seeded registries.yaml, covering removal, the not-found success,
// and the dry-run preview.
func TestDeleteRegistryEndToEnd(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOut    []string
		wantState  []string // substrings the state must still contain
		wantAbsent []string // substrings the state must no longer contain
	}{
		{
			name:       "removes the named registry and reports the summary",
			args:       []string{acmeName},
			wantOut:    []string{summaryRemove + "; 0 artifacts removed"},
			wantState:  []string{internalName},
			wantAbsent: []string{docNameAcme},
		},
		{
			name:      "unknown registry reports nothing was deleted and leaves state",
			args:      []string{argGhost},
			wantOut:   []string{"nothing was deleted"},
			wantState: []string{docNameAcme, "name: " + internalName},
		},
		{
			name:      "dry-run previews without changing state",
			args:      []string{acmeName, "--dry-run"},
			wantOut:   []string{"would be removed"},
			wantState: []string{docNameAcme, "name: " + internalName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			seedRegistries(t, twoRegistries)

			// Act.
			out, err := runDeleteRegistry(t, tt.args...)

			// Assert.
			require.NoError(t, err)
			for _, want := range tt.wantOut {
				assert.Contains(t, out, want)
			}
			state := readRegistries(t)
			for _, want := range tt.wantState {
				assert.Contains(t, state, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, state, absent)
			}
		})
	}
}

// TestDeleteRegistryEmptiesState removes the only registry, leaving an empty stream.
func TestDeleteRegistryEmptiesState(t *testing.T) {
	// Arrange: a single registry.
	single := strings.SplitN(twoRegistries, "---\n", 2)[0]
	seedRegistries(t, single)

	// Act.
	out, err := runDeleteRegistry(t, acmeName)

	// Assert: removed, summary printed, and acme is gone from state.
	require.NoError(t, err)
	assert.Contains(t, out, summaryRemove)
	assert.NotContains(t, readRegistries(t), docNameAcme)
}
