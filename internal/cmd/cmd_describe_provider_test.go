package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// syncedProvider is a schema-valid settings.yaml stream carrying a synced
// provider, used to assert the full descriptor renders across every field.
const syncedProvider = `apiVersion: sauron.raitonbl.com/v1
kind: Provider
metadata:
  name: claude
  createdAt: "2026-06-21T07:30:00Z"
  lastUpdatedAt: "2026-06-22T08:00:00Z"
  labels:
    team: backend
spec:
  lastSyncedAt: "2026-06-25T09:15:00Z"
  lastSyncAttemptAt: "2026-06-26T06:00:00Z"
`

// runDescribeProvider assembles and runs the subcommand, returning stdout and the
// resulting error.
func runDescribeProvider(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := DescribeProvider()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// TestDescribeGroupAttachesProvider asserts the provider subcommand is attached to
// the describe group.
func TestDescribeGroupAttachesProvider(t *testing.T) {
	// Arrange + Act.
	cmd := Describe()

	// Assert.
	var attached bool
	for _, sub := range cmd.Commands() {
		if sub.Name() == subcmdProvider {
			attached = true
		}
	}
	assert.True(t, attached, "the provider subcommand is attached")
}

// TestDescribeProviderFlagSurface asserts the --fields flag and the argument
// validator are installed.
func TestDescribeProviderFlagSurface(t *testing.T) {
	// Arrange + Act.
	cmd := DescribeProvider()

	// Assert.
	assert.NotNil(t, cmd.Flags().Lookup(fieldsName), "flag fields registered")
	assert.NotNil(t, cmd.Args, "an argument validator is installed")
}

// TestDescribeProviderRejectsBadInput asserts an unexpected argument or an unknown
// flag is rejected before the graph is built and maps to the usage exit code.
func TestDescribeProviderRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: caseUnexpectedArg, args: []string{argExtra}},
		{name: caseUnknown, args: []string{flagUnknown}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			_, err := runDescribeProvider(t, tt.args...)

			// Assert.
			require.Error(t, err)
			assert.Equal(t, exitUsage, exitCode(err))
		})
	}
}

// TestDescribeProviderEndToEnd drives the assembled subcommand through the real fx
// graph against a seeded settings.yaml, covering the full synced detail, field
// selection, the none-set line (exit 0, not an error), and the usage error. name
// is the identity and is always present and first.
func TestDescribeProviderEndToEnd(t *testing.T) {
	tests := []struct {
		name       string
		seed       string
		args       []string
		wantOut    []string
		wantAbsent []string
		wantErr    bool
		wantUsage  bool
	}{
		{
			name:    "full detail shows every populated field including the sync timestamps",
			seed:    syncedProvider,
			wantOut: []string{labelName, nameClaude, labelDirectory, dirClaude, labelLabels, labelTeam, valBackend, labelCreated, createdStamp, labelUpdated, updatedStamp, labelLastSynced, syncedStamp, labelLastSyncAttmpt, attemptStamp},
		},
		{
			name:       "fields selects and orders, name forced first",
			seed:       syncedProvider,
			args:       []string{flagFields, "directory,name"},
			wantOut:    []string{labelName, labelDirectory},
			wantAbsent: []string{labelLabels, labelCreated, labelLastSynced},
		},
		{
			name:    "no provider set reports the none-set line and exits successfully",
			seed:    "",
			wantOut: []string{noProviderMessage},
		},
		{
			name:      "invalid field is a usage error",
			seed:      syncedProvider,
			args:      []string{flagFields, nameBogus},
			wantErr:   true,
			wantUsage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			seedRegistries(t, tt.seed)

			// Act.
			out, err := runDescribeProvider(t, tt.args...)

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantUsage {
					assert.Equal(t, exitUsage, exitCode(err))
				}
				return
			}
			require.NoError(t, err)
			for _, want := range tt.wantOut {
				assert.Contains(t, out, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, out, absent)
			}
		})
	}
}
