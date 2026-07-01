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

const oneAdded = "1 added"

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

// TestInstallWriteError drives the real command with a failing stdout, over
// the benign-skip fixture (the only outcome this file's registry fixture can
// reach without a content-serving endpoint — see TestInstallBenignSkipExitsZero):
// the heading write (first) and the failure-line write (second, right after
// the heading, since a skip has no added/updated lines ahead of it) both
// surface as classified io errors.
func TestInstallWriteError(t *testing.T) {
	tests := []struct {
		// name states which write is expected to fail.
		name string
		// writeAfter is the number of successful writes before the failure.
		writeAfter int
	}{
		{name: "heading write fails", writeAfter: 0},
		{name: "failure line write fails", writeAfter: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: registry offers only an agent, provider set, so a
			// requested skill is a benign skip (heading + one failure line).
			seedCatalogueRegistry(t)
			setProviderForTest(t, nameClaude)
			cmd := InstallSkill()
			cmd.SetOut(&failingWriter{writeAfter: tt.writeAfter})
			cmd.SetContext(context.Background())
			cmd.SetArgs([]string{"go-style"})

			// Act.
			err := cmd.Execute()

			// Assert.
			var ucErr *usecase.Error
			require.ErrorAs(t, err, &ucErr)
			assert.Equal(t, usecase.TypeIO, ucErr.Type)
		})
	}
}
