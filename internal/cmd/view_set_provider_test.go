package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// set-provider-view-test literals, named to satisfy goconst across the package.
const (
	skillAcmeGoStyle  = "sauron-acme-go-style"
	agentAcmeReviewer = "sauron-acme-code-reviewer"
)

// TestRenderSetProvider asserts the rendered output for each outcome shape.
func TestRenderSetProvider(t *testing.T) {
	tests := []struct {
		name   string
		result *usecase.SetProviderResponse
		want   string
	}{
		{
			name: "change with both groups",
			result: &usecase.SetProviderResponse{
				Provider: nameZencoder,
				Migrated: 2,
				Skills:   []string{skillAcmeGoStyle},
				Agents:   []string{agentAcmeReviewer},
			},
			want: "skills:\n  ~ " + skillAcmeGoStyle + "\nagents:\n  ~ " + agentAcmeReviewer + "\nprovider set to \"zencoder\"; 2 artifacts migrated\n",
		},
		{
			name: "change with only skills",
			result: &usecase.SetProviderResponse{
				Provider: nameClaude,
				Migrated: 1,
				Skills:   []string{skillAcmeGoStyle},
			},
			want: "skills:\n  ~ " + skillAcmeGoStyle + "\nprovider set to \"claude\"; 1 artifacts migrated\n",
		},
		{
			name:   "first set with nothing installed",
			result: &usecase.SetProviderResponse{Provider: nameClaude},
			want:   "provider set to \"claude\"\n",
		},
		{
			name:   "already active reports no change",
			result: &usecase.SetProviderResponse{Provider: nameClaude, Unchanged: true},
			want:   "provider already set to \"claude\"\n",
		},
		{
			// A2: migration failures are rendered as "! name: reason" lines after the
			// migrated groups and before the summary; they do not affect the exit code.
			name: "change with migration failures",
			result: &usecase.SetProviderResponse{
				Provider: nameZencoder,
				Migrated: 1,
				Skills:   []string{skillAcmeGoStyle},
				Failures: []usecase.MigrateFailure{
					{
						Reason:   "file not found",
						Artifact: types.Artifact{Metadata: types.Metadata{Name: "orphan-skill"}},
					},
				},
			},
			want: "skills:\n  ~ " + skillAcmeGoStyle + "\n  ! orphan-skill: file not found\nprovider set to \"zencoder\"; 1 artifacts migrated\n",
		},
		{
			// A2: multiple failures are each rendered on their own line.
			name: "change with multiple failures",
			result: &usecase.SetProviderResponse{
				Provider: nameClaude,
				Failures: []usecase.MigrateFailure{
					{Reason: "permission denied", Artifact: types.Artifact{Metadata: types.Metadata{Name: "locked-agent"}}},
					{Reason: "disk full", Artifact: types.Artifact{Metadata: types.Metadata{Name: "heavy-skill"}}},
				},
			},
			want: "  ! locked-agent: permission denied\n  ! heavy-skill: disk full\nprovider set to \"claude\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := renderSetProvider(&buf, tt.result)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestRenderSetProviderWriteError surfaces a writer failure as an io error on
// each reachable write for the change path (skills + agents, no failures).
func TestRenderSetProviderWriteError(t *testing.T) {
	result := &usecase.SetProviderResponse{
		Provider: nameZencoder,
		Migrated: 2,
		Skills:   []string{"s"},
		Agents:   []string{"a"},
	}

	for _, after := range []int{0, 1, 2, 3, 4} {
		err := renderSetProvider(&failingWriter{writeAfter: after}, result)
		var ucErr *usecase.Error
		require.ErrorAs(t, err, &ucErr)
		assert.Equal(t, usecase.TypeIO, ucErr.Type)
	}
}

// TestRenderSetProviderFailureWriteError surfaces a writer failure on the
// failure-line write of the change path.
func TestRenderSetProviderFailureWriteError(t *testing.T) {
	// Arrange: one skill group (2 writes) then the failure line (1 write).
	result := &usecase.SetProviderResponse{
		Provider: nameZencoder,
		Migrated: 1,
		Skills:   []string{"s"},
		Failures: []usecase.MigrateFailure{
			{Reason: "oops", Artifact: types.Artifact{Metadata: types.Metadata{Name: "x"}}},
		},
	}

	// Fail on the third write (skills heading=1, skill item=2, failure line=3).
	err := renderSetProvider(&failingWriter{writeAfter: 2}, result)

	var ucErr *usecase.Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, usecase.TypeIO, ucErr.Type)
}

// TestRenderSetProviderUnchangedWriteError surfaces the writer failure on the
// no-change path.
func TestRenderSetProviderUnchangedWriteError(t *testing.T) {
	err := renderSetProvider(&failingWriter{}, &usecase.SetProviderResponse{Provider: nameClaude, Unchanged: true})
	var ucErr *usecase.Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, usecase.TypeIO, ucErr.Type)
}
