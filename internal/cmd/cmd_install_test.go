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

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// install-cmd test literals, named to satisfy goconst across the package.
const subcmdInstall = "install"

// runInstallSkill assembles and runs the skill subcommand, returning stdout
// and the resulting error.
func runInstallSkill(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := InstallSkill()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// runInstallAgent assembles and runs the agent subcommand, returning stdout
// and the resulting error.
func runInstallAgent(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := InstallAgent()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// setProviderForTest sets the active provider to name, failing the test on
// error. It is a setup helper shared by install scenarios that require a
// provider to be configured before installing.
func setProviderForTest(t *testing.T, name string) {
	t.Helper()
	sp := SetProvider()
	sp.SetOut(&bytes.Buffer{})
	sp.SetErr(&bytes.Buffer{})
	sp.SetContext(context.Background())
	sp.SetArgs([]string{name})
	require.NoError(t, sp.Execute())
}

// TestInstallGroup asserts the install group is a pure command group: a bare
// invocation prints help and exits 0 (no RunE), and the per-kind subcommands
// are attached.
func TestInstallGroup(t *testing.T) {
	// Arrange.
	cmd := Install()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(nil)

	// Act: invoking the group without a kind noun.
	err := cmd.Execute()

	// Assert: a group with no run behaviour succeeds and prints its help.
	assert.Equal(t, subcmdInstall, cmd.Name())
	assert.Nil(t, cmd.RunE, "the group has no run behaviour")
	require.NoError(t, err)
	assert.Equal(t, exitOK, ExitCode(err))

	names := map[string]bool{}
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, want := range []string{subcmdSkill, subcmdAgent} {
		assert.Truef(t, names[want], "the %q subcommand is attached", want)
	}
}

// TestRootAttachesInstall asserts the root command registers the install group.
func TestRootAttachesInstall(t *testing.T) {
	// Arrange.
	t.Setenv("SAURON_HOME", t.TempDir())

	// Act.
	root, err := New("sauron", "0.0.0", "abc")
	require.NoError(t, err)

	// Assert.
	var install *cobra.Command
	for _, sub := range root.Commands() {
		if sub.Name() == subcmdInstall {
			install = sub
		}
	}
	require.NotNil(t, install, "the install subcommand is attached")
}

// TestInstallFlagSurface asserts each leaf registers an argument validator and
// carries no unexpected flags.
func TestInstallFlagSurface(t *testing.T) {
	builders := map[string]func() *cobra.Command{
		subcmdSkill: InstallSkill,
		subcmdAgent: InstallAgent,
	}

	for name, build := range builders {
		t.Run(name, func(t *testing.T) {
			// Act.
			cmd := build()

			// Assert: an argument validator guards the positional input.
			assert.NotNil(t, cmd.Args, "an argument validator is installed")
		})
	}
}

// TestInstallRejectsMissingArg asserts that a missing artifact name is
// classified as a usage error (exit 2) for both kind subcommands.
func TestInstallRejectsMissingArg(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, ...string) (string, error)
	}{
		{name: "skill rejects missing name", run: runInstallSkill},
		{name: "agent rejects missing name", run: runInstallAgent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			t.Setenv("SAURON_HOME", t.TempDir())

			// Act: no positional argument supplied.
			_, err := tt.run(t)

			// Assert: cobra's MinimumNArgs(1) fires as a usage error.
			require.Error(t, err)
			assert.Equal(t, exitUsage, ExitCode(err))
		})
	}
}

// TestInstallNoProviderIsRuntimeError asserts that running install when no
// provider is set produces a runtime error (exit 1), not a usage error.
func TestInstallNoProviderIsRuntimeError(t *testing.T) {
	// Arrange: a registry is configured (so the registry lookup succeeds) but no
	// provider is set, triggering the not-found runtime error.
	seedCatalogueRegistry(t)

	// Act.
	_, err := runInstallSkill(t, "go-style")

	// Assert.
	require.Error(t, err)
	assert.Equal(t, exitError, ExitCode(err))
	assert.NotEqual(t, exitUsage, ExitCode(err))
}

// TestInstallBenignSkipExitsZero asserts that names the registry does not offer
// are reported as benign skips and the command exits 0 (FR-006). The failure
// line is still rendered so the caller can see which names were skipped.
func TestInstallBenignSkipExitsZero(t *testing.T) {
	// Arrange: registry that offers only an agent (no skills), provider set.
	// seedCatalogueRegistry starts an http registry with agent "code-reviewer"
	// and no skills; "go-style" is not offered → benign skip.
	seedCatalogueRegistry(t)
	setProviderForTest(t, nameClaude)

	// Act: install a skill the registry does not offer.
	out, err := runInstallSkill(t, "go-style")

	// Assert: benign skip exits 0 and the failure line is still rendered.
	require.NoError(t, err)
	assert.Equal(t, exitOK, ExitCode(err))
	assert.Contains(t, out, "go-style")
}

// TestInstallFatalFailureExitsOne asserts that a persist failure (a fetch or
// write that could not complete) exits 1 even when the result is fully rendered.
// The registry offers the skill in the listing but has no content handler;
// the HTTP 404 on content fetch is a commit error → Fatal failure.
func TestInstallFatalFailureExitsOne(t *testing.T) {
	// Arrange: registry that lists "go-style" but has no content endpoint.
	// The startHTTPRegistry helper only serves /skills and /agents; a GET to
	// /skills/go-style returns 404 → fetch fails → result.fail() → Fatal=true.
	home := t.TempDir()
	t.Setenv("SAURON_HOME", home)

	source := startHTTPRegistry(t,
		[]artifactSummary{{Name: "go-style", Version: versionOne, Size: 1024}},
		nil,
	)
	stream := "apiVersion: sauron.raitonbl.com/v1\nkind: Registry\nmetadata:\n  name: " + acmeName +
		"\nspec:\n  transport: http\n  source: " + source + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(home, settingsFile), []byte(stream), 0o644))

	setProviderForTest(t, nameClaude)

	// Act.
	out, err := runInstallSkill(t, "go-style")

	// Assert: fatal failure exits 1 and the failure line is still rendered.
	require.Error(t, err)
	assert.Equal(t, exitError, ExitCode(err))
	assert.NotEqual(t, exitUsage, ExitCode(err))
	assert.Contains(t, out, "go-style")
	assert.Contains(t, out, "skills:")
}

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
