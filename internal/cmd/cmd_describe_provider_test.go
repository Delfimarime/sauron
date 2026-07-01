package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
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
			assert.Equal(t, exitUsage, ExitCode(err))
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
					assert.Equal(t, exitUsage, ExitCode(err))
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

// describe-provider-view-test literals, named to satisfy goconst across the
// package.
const (
	labelName           = "name:"
	labelDirectory      = "directory:"
	labelLabels         = "labels:"
	labelLastSynced     = "lastSyncedAt:"
	labelLastSyncAttmpt = "lastSyncAttemptAt:"
	dirClaude           = "~/.claude"
	syncedStamp         = "2026-06-25T09:15:00Z"
	attemptStamp        = "2026-06-26T06:00:00Z"
	labelTeam           = "team:"
	valBackend          = "backend"
	nameZencoder        = "zencoder"
	dirZencoder         = "~/.zencoder"
)

// allDescribeProviderFields is the full, ordered field selection a default
// describe yields.
func allDescribeProviderFields() []string {
	return []string{
		describeProviderFieldName, describeProviderFieldDirectory, describeProviderFieldLabels,
		describeFieldCreated, describeFieldUpdated,
		describeProviderFieldLastSynced, describeProviderFieldLastSyncAttempt,
	}
}

// fullViewProvider is a provider populated across every describable field.
func fullViewProvider() types.Provider {
	return types.Provider{
		Metadata: types.Metadata{
			Name:          nameClaude,
			Labels:        map[string]string{"team": valBackend},
			CreatedAt:     createdStamp,
			LastUpdatedAt: updatedStamp,
		},
		Spec: types.ProviderSpec{
			LastSyncedAt:      syncedStamp,
			LastSyncAttemptAt: attemptStamp,
		},
	}
}

// TestProjectProvider covers the projection + descriptor rendering across
// the default view, field selection, the derived directory, the sorted labels
// block, and omission of unpopulated fields. name is the identity and is always
// present and first. projectProvider is composed with the shared descriptor
// renderer directly — this is what describeProvider's handler does inline,
// without a separate render function to call.
func TestProjectProvider(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// provider is the record to project.
		provider types.Provider
		// fields is the resolved, ordered field selection.
		fields []string
		// wantContains are substrings the output must contain, in order.
		wantContains []string
		// wantAbsent are substrings the output must never contain.
		wantAbsent []string
	}{
		{
			name: "common case shows name, directory, created, updated and omits sync",
			provider: types.Provider{
				Metadata: types.Metadata{
					Name:          nameClaude,
					CreatedAt:     createdStamp,
					LastUpdatedAt: updatedStamp,
				},
			},
			fields:       allDescribeProviderFields(),
			wantContains: []string{labelName, nameClaude, labelDirectory, dirClaude, labelCreated, createdStamp, labelUpdated, updatedStamp},
			wantAbsent:   []string{labelLabels, labelLastSynced, labelLastSyncAttmpt},
		},
		{
			name:     "default shows every populated field including the sync timestamps",
			provider: fullViewProvider(),
			fields:   allDescribeProviderFields(),
			wantContains: []string{
				labelName, nameClaude,
				labelDirectory, dirClaude,
				labelLabels, labelTeam, valBackend,
				labelCreated, createdStamp,
				labelUpdated, updatedStamp,
				labelLastSynced, syncedStamp,
				labelLastSyncAttmpt, attemptStamp,
			},
		},
		{
			name:         "projects the resolved fields in order",
			provider:     fullViewProvider(),
			fields:       []string{describeProviderFieldName, describeProviderFieldDirectory, describeProviderFieldLastSynced},
			wantContains: []string{labelName, labelDirectory, labelLastSynced},
			wantAbsent:   []string{labelLabels, labelCreated, labelUpdated, labelLastSyncAttmpt},
		},
		{
			name: "labels render their keys sorted",
			provider: types.Provider{
				Metadata: types.Metadata{
					Name:   nameClaude,
					Labels: map[string]string{"zone": "eu", "app": "web", "team": "backend"},
				},
			},
			fields:       []string{describeProviderFieldName, describeProviderFieldLabels},
			wantContains: []string{labelLabels, "app:", labelTeam, "zone:"},
		},
		{
			name:         "zencoder derives its own directory",
			provider:     types.Provider{Metadata: types.Metadata{Name: types.ProviderZencoder}},
			fields:       []string{describeProviderFieldName, describeProviderFieldDirectory},
			wantContains: []string{labelName, nameZencoder, labelDirectory, dirZencoder},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer
			provider := tt.provider
			view := descriptor{Fields: projectProvider(provider, tt.fields)}

			// Act.
			err := view.render(&buf)

			// Assert.
			require.NoError(t, err)
			out := buf.String()
			lastIndex := -1
			for _, want := range tt.wantContains {
				idx := strings.Index(out, want)
				require.GreaterOrEqualf(t, idx, 0, "output %q missing %q", out, want)
				assert.Greaterf(t, idx, lastIndex, "%q is out of order in %q", want, out)
				lastIndex = idx
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContainsf(t, out, absent, "output unexpectedly contains %q", absent)
			}
		})
	}
}

// TestProjectProviderFullLayout pins the exact aligned descriptor of a
// synced provider — name first, the derived directory, the key-sorted labels
// section, the audit timestamps, and the sync timestamps — through the shared
// renderer (column aligned to the widest leaf label, lastSyncAttemptAt).
func TestProjectProviderFullLayout(t *testing.T) {
	// Arrange.
	var buf bytes.Buffer
	provider := fullViewProvider()
	view := descriptor{Fields: projectProvider(provider, allDescribeProviderFields())}
	want := "name:               claude\n" +
		"directory:          ~/.claude\n" +
		"labels:\n" +
		"  team:             backend\n" +
		"createdAt:          2026-06-21T07:30:00Z\n" +
		"lastUpdatedAt:      2026-06-22T08:00:00Z\n" +
		"lastSyncedAt:       2026-06-25T09:15:00Z\n" +
		"lastSyncAttemptAt:  2026-06-26T06:00:00Z\n"

	// Act.
	err := view.render(&buf)

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, want, buf.String())
}

// TestDescribeProviderWriteError drives the real command with a failing
// stdout, covering both the descriptor write (a synced provider) and the
// none-set line write (no provider configured) — both surface as a
// classified io error.
func TestDescribeProviderWriteError(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// seed is the settings.yaml content; empty means no provider set.
		seed string
	}{
		{name: "descriptor write fails", seed: syncedProvider},
		{name: "no-provider line write fails", seed: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			seedRegistries(t, tt.seed)
			cmd := DescribeProvider()
			cmd.SetOut(&failingWriter{})
			cmd.SetContext(context.Background())
			cmd.SetArgs(nil)

			// Act.
			err := cmd.Execute()

			// Assert.
			var ucErr *usecase.Error
			require.ErrorAs(t, err, &ucErr)
			assert.Equal(t, usecase.TypeIO, ucErr.Type)
		})
	}
}

// TestSelectDescribeProviderFields covers the default, identity-first ordering,
// dedupe, and unknown-field paths of the view's field selector. An unknown field
// is a usage error raised at the command boundary.
func TestSelectDescribeProviderFields(t *testing.T) {
	t.Run("empty request yields every field in order", func(t *testing.T) {
		got, err := selectDescribeProviderFields(nil)
		require.NoError(t, err)
		assert.Equal(t, allDescribeProviderFields(), got)
	})

	t.Run("selection forces name present and first, deduped", func(t *testing.T) {
		got, err := selectDescribeProviderFields([]string{describeProviderFieldDirectory, describeProviderFieldLabels, describeProviderFieldDirectory})
		require.NoError(t, err)
		assert.Equal(t, []string{describeProviderFieldName, describeProviderFieldDirectory, describeProviderFieldLabels}, got)
	})

	t.Run("unknown field is a usage error", func(t *testing.T) {
		got, err := selectDescribeProviderFields([]string{nameBogus})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, errInvalidFlag)
	})
}
