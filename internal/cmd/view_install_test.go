package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

const (
	headSkills = "skills:"
	headAgents = "agents:"
	oneAdded   = "1 added"
)

// installArtifact is a test helper that builds an Artifact with the given
// kind and name, matching what InstallUseCase places in InstallResponse.
func installArtifact(kind, name string) types.Artifact {
	return types.Artifact{
		TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: kind},
		Metadata: types.Metadata{Name: name},
	}
}

// TestRenderInstall covers every plan variant: additions, updates, combined
// added+updated, no-op, and per-name failures; the kind heading, artifact
// sigils, and summary count are exercised for both skill and agent kinds.
func TestRenderInstall(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// kind is the artifact kind the install command was invoked for.
		kind string
		// result is the use-case outcome to render.
		result *usecase.InstallResponse
		// wantContains are substrings the output must contain.
		wantContains []string
		// wantAbsent are substrings the output must not contain.
		wantAbsent []string
	}{
		{
			name: "add renders + and N added",
			kind: types.KindSkill,
			result: &usecase.InstallResponse{
				Added: []types.Artifact{
					installArtifact(types.KindSkill, "go-style"),
					installArtifact(types.KindSkill, "sql-review"),
				},
			},
			wantContains: []string{headSkills, "+ sauron-go-style", "+ sauron-sql-review", "2 added"},
			wantAbsent:   []string{"~", "updated"},
		},
		{
			name: "update renders ~ and N updated",
			kind: types.KindSkill,
			result: &usecase.InstallResponse{
				Updated: []types.Artifact{
					installArtifact(types.KindSkill, "go-style"),
				},
			},
			wantContains: []string{headSkills, "~ sauron-go-style", "1 updated"},
			wantAbsent:   []string{"+", "added"},
		},
		{
			name: "combined renders both + and ~ with combined summary",
			kind: types.KindAgent,
			result: &usecase.InstallResponse{
				Added:   []types.Artifact{installArtifact(types.KindAgent, "code-reviewer")},
				Updated: []types.Artifact{installArtifact(types.KindAgent, "doc-writer")},
			},
			wantContains: []string{
				headAgents, "+ sauron-code-reviewer", "~ sauron-doc-writer", "1 added, 1 updated",
			},
		},
		{
			name:         "no-op renders heading but no summary",
			kind:         types.KindSkill,
			result:       &usecase.InstallResponse{},
			wantContains: []string{headSkills},
			wantAbsent:   []string{"added", "updated", "+", "~"},
		},
		{
			name: "failure is reported with the artifact name",
			kind: types.KindSkill,
			result: &usecase.InstallResponse{
				Added:    []types.Artifact{installArtifact(types.KindSkill, "go-style")},
				Failures: []usecase.InstallFailure{{Name: "unknown-skill", Reason: "not offered"}},
			},
			wantContains: []string{headSkills, "+ sauron-go-style", "unknown-skill", oneAdded},
		},
		{
			name: "agent heading renders agents:",
			kind: types.KindAgent,
			result: &usecase.InstallResponse{
				Added: []types.Artifact{installArtifact(types.KindAgent, "code-reviewer")},
			},
			wantContains: []string{headAgents, "+ sauron-code-reviewer", oneAdded},
			wantAbsent:   []string{headSkills},
		},
		{
			name: "skill heading renders skills:",
			kind: types.KindSkill,
			result: &usecase.InstallResponse{
				Added: []types.Artifact{installArtifact(types.KindSkill, "writer")},
			},
			wantContains: []string{headSkills, "+ sauron-writer", oneAdded},
			wantAbsent:   []string{headAgents},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := renderInstall(&buf, tt.kind, tt.result)

			// Assert.
			require.NoError(t, err)
			out := buf.String()
			for _, want := range tt.wantContains {
				assert.Contains(t, out, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, out, absent)
			}
		})
	}
}

// TestInstallSummary exercises the summary helper for all count combinations
// and the no-op (zero counts) case.
func TestInstallSummary(t *testing.T) {
	tests := []struct {
		name   string
		result *usecase.InstallResponse
		want   string
	}{
		{"zero counts returns empty", &usecase.InstallResponse{}, ""},
		{"one added only", &usecase.InstallResponse{Added: make([]types.Artifact, 1)}, oneAdded},
		{"two added only", &usecase.InstallResponse{Added: make([]types.Artifact, 2)}, "2 added"},
		{"one updated only", &usecase.InstallResponse{Updated: make([]types.Artifact, 1)}, "1 updated"},
		{
			"combined added and updated",
			&usecase.InstallResponse{Added: make([]types.Artifact, 1), Updated: make([]types.Artifact, 1)},
			"1 added, 1 updated",
		},
		{
			"many added and many updated",
			&usecase.InstallResponse{Added: make([]types.Artifact, 3), Updated: make([]types.Artifact, 2)},
			"3 added, 2 updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, installSummary(tt.result))
		})
	}
}

// TestRenderInstallWriteError surfaces writer failures as an io error for each
// write the view performs: heading, added, updated, failure, and summary.
func TestRenderInstallWriteError(t *testing.T) {
	oneArtifact := installArtifact(types.KindSkill, "a")

	tests := []struct {
		// name states which write is expected to fail.
		name string
		// result is the use-case outcome fed to renderInstall.
		result *usecase.InstallResponse
		// writeAfter is the number of successful writes before the failure.
		writeAfter int
	}{
		{
			name:       "heading write fails",
			result:     &usecase.InstallResponse{},
			writeAfter: 0,
		},
		{
			name:       "added line write fails",
			result:     &usecase.InstallResponse{Added: []types.Artifact{oneArtifact}},
			writeAfter: 1,
		},
		{
			name:       "updated line write fails",
			result:     &usecase.InstallResponse{Updated: []types.Artifact{oneArtifact}},
			writeAfter: 1,
		},
		{
			name:       "failure line write fails",
			result:     &usecase.InstallResponse{Failures: []usecase.InstallFailure{{Name: "a", Reason: "r"}}},
			writeAfter: 1,
		},
		{
			name:       "summary write fails",
			result:     &usecase.InstallResponse{Added: []types.Artifact{oneArtifact}},
			writeAfter: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			err := renderInstall(&failingWriter{writeAfter: tt.writeAfter}, types.KindSkill, tt.result)

			// Assert.
			var ucErr *usecase.Error
			require.ErrorAs(t, err, &ucErr)
			assert.Equal(t, usecase.TypeIO, ucErr.Type)
		})
	}
}
