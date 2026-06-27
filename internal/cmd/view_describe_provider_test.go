package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

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

// TestRenderDescribeProvider covers the projection + descriptor rendering across
// the default view, field selection, the derived directory, the sorted labels
// block, and omission of unpopulated fields. name is the identity and is always
// present and first.
func TestRenderDescribeProvider(t *testing.T) {
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

			// Act.
			err := renderDescribeProvider(&buf, &provider, tt.fields)

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

// TestRenderDescribeProviderFullLayout pins the exact aligned descriptor of a
// synced provider — name first, the derived directory, the key-sorted labels
// section, the audit timestamps, and the sync timestamps — through the shared
// renderer (column aligned to the widest leaf label, lastSyncAttemptAt).
func TestRenderDescribeProviderFullLayout(t *testing.T) {
	// Arrange.
	var buf bytes.Buffer
	provider := fullViewProvider()
	want := "name:               claude\n" +
		"directory:          ~/.claude\n" +
		"labels:\n" +
		"  team:             backend\n" +
		"createdAt:          2026-06-21T07:30:00Z\n" +
		"lastUpdatedAt:      2026-06-22T08:00:00Z\n" +
		"lastSyncedAt:       2026-06-25T09:15:00Z\n" +
		"lastSyncAttemptAt:  2026-06-26T06:00:00Z\n"

	// Act.
	err := renderDescribeProvider(&buf, &provider, allDescribeProviderFields())

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, want, buf.String())
}

// TestRenderDescribeProviderWriteError surfaces a writer failure as an io error.
func TestRenderDescribeProviderWriteError(t *testing.T) {
	// Arrange.
	provider := fullViewProvider()

	// Act.
	err := renderDescribeProvider(&failingWriter{}, &provider, allDescribeProviderFields())

	// Assert.
	var ucErr *usecase.Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, usecase.TypeIO, ucErr.Type)
}

// TestRenderNoProvider asserts the none-set line is the constant message followed
// by a newline.
func TestRenderNoProvider(t *testing.T) {
	// Arrange.
	var buf bytes.Buffer

	// Act.
	err := renderNoProvider(&buf)

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, noProviderMessage+"\n", buf.String())
}

// TestRenderNoProviderWriteError surfaces a writer failure as an io error.
func TestRenderNoProviderWriteError(t *testing.T) {
	// Act.
	err := renderNoProvider(&failingWriter{})

	// Assert.
	var ucErr *usecase.Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, usecase.TypeIO, ucErr.Type)
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
