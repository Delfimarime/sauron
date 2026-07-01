package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
)

// trackedSkillStream is a schema-valid track.yaml holding a single installed
// skill whose on-disk file is intentionally never created, so a provider switch
// strands it on migration.
const trackedSkillStream = "---\n" +
	"apiVersion: sauron.raitonbl.com/v1\n" +
	"kind: Skill\n" +
	"metadata:\n" +
	"  name: go-style\n" +
	"spec:\n" +
	"  version: v1.0.0\n" +
	"  path: sauron-go-style\n" +
	"  installedAt: \"2024-01-01T00:00:00Z\"\n" +
	"  updatedAt: \"2024-01-01T00:00:00Z\"\n"

// runSetProvider assembles and runs the provider subcommand, returning stdout and
// the resulting error.
func runSetProvider(t *testing.T, name string) (string, error) {
	t.Helper()
	cmd := SetProvider()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{name})
	err := cmd.Execute()
	return stdout.String(), err
}

// TestSetProviderHandlerRejectsUnknownName asserts an unsupported provider name
// surfaces as a usage error mapped to exit code 2.
func TestSetProviderHandlerRejectsUnknownName(t *testing.T) {
	// Arrange.
	t.Setenv("SAURON_HOME", t.TempDir())

	// Act.
	err := setProvider(context.Background(), []string{nameBogus}, &bytes.Buffer{})

	// Assert.
	require.Error(t, err)
	assert.Equal(t, exitUsage, ExitCode(err))
}

// TestSetProviderCommand exercises the assembled subcommand: ExactArgs(1) and the
// unsupported-name path, all mapped to the usage exit code. A temporary
// SAURON_HOME keeps nothing durable.
func TestSetProviderCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "rejects a missing argument", args: []string{}},
		{name: caseUnexpectedArg, args: []string{nameClaude, "extra"}},
		{name: "rejects an unsupported name", args: []string{nameBogus}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			t.Setenv("SAURON_HOME", t.TempDir())
			cmd := SetProvider()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs(tt.args)
			cmd.SetContext(context.Background())

			// Act.
			err := cmd.Execute()

			// Assert.
			require.Error(t, err)
			assert.Equal(t, exitUsage, ExitCode(err))
		})
	}
}

// TestSetProviderEndToEnd drives the assembled subcommand through the real fx
// graph: the first invocation records the provider, a second invocation of the
// same provider reports no change. State lives in a temp SAURON_HOME.
func TestSetProviderEndToEnd(t *testing.T) {
	// Arrange.
	t.Setenv("SAURON_HOME", t.TempDir())

	// Act: first set records the provider.
	var first bytes.Buffer
	set := SetProvider()
	set.SetOut(&first)
	set.SetErr(&bytes.Buffer{})
	set.SetContext(context.Background())
	set.SetArgs([]string{nameClaude})
	require.NoError(t, set.Execute())

	// Assert: with nothing installed the summary reads cleanly.
	assert.Equal(t, "provider set to \"claude\"\n", first.String())

	// Act: re-setting the active provider changes nothing.
	var second bytes.Buffer
	reset := SetProvider()
	reset.SetOut(&second)
	reset.SetErr(&bytes.Buffer{})
	reset.SetContext(context.Background())
	reset.SetArgs([]string{nameClaude})
	require.NoError(t, reset.Execute())

	// Assert.
	assert.Equal(t, "provider already set to \"claude\"\n", second.String())
}

// TestSetProviderCleanSwitchExitsZero asserts a real provider change with nothing
// installed migrates nothing and exits 0.
func TestSetProviderCleanSwitchExitsZero(t *testing.T) {
	// Arrange: HOME redirected so the provider filesystem stays off the real FS.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SAURON_HOME", t.TempDir())
	_, err := runSetProvider(t, nameClaude)
	require.NoError(t, err)

	// Act: switch to a different provider with no installed artifacts.
	out, switchErr := runSetProvider(t, "zencoder")

	// Assert: a clean switch exits 0 and confirms the new provider.
	require.NoError(t, switchErr)
	assert.Equal(t, exitOK, ExitCode(switchErr))
	assert.Equal(t, "provider set to \"zencoder\"\n", out)
}

// TestSetProviderMigrationFailureExitsOne asserts that a stranded artifact (its
// file missing under the old provider directory, so the migration rename fails)
// exits 1 even though the provider is still persisted and the failure is rendered.
func TestSetProviderMigrationFailureExitsOne(t *testing.T) {
	// Arrange: HOME redirected so the provider filesystem stays off the real FS;
	// the active provider is claude and one skill is recorded as installed, but its
	// file is never created so the rename to zencoder's directory fails.
	t.Setenv("HOME", t.TempDir())
	home := t.TempDir()
	t.Setenv("SAURON_HOME", home)

	_, err := runSetProvider(t, nameClaude)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(home, "track.yaml"), []byte(trackedSkillStream), 0o644))

	// Act: switching strands the skill whose file is absent.
	out, switchErr := runSetProvider(t, "zencoder")

	// Assert: a failed migration exits 1 (a runtime error, not a usage error) with
	// the failure rendered and the new provider still confirmed (FR-005).
	require.Error(t, switchErr)
	assert.Equal(t, exitError, ExitCode(switchErr))
	assert.NotEqual(t, exitUsage, ExitCode(switchErr))
	assert.Contains(t, out, "go-style")
	assert.Contains(t, out, "provider set to \"zencoder\"")
}

// TestSetGroupAttachesProvider asserts the provider subcommand is attached to the
// set group.
func TestSetGroupAttachesProvider(t *testing.T) {
	// Arrange + Act.
	cmd := Set()

	// Assert.
	var provider bool
	for _, sub := range cmd.Commands() {
		if sub.Name() == subcmdProvider {
			provider = true
		}
	}
	assert.True(t, provider, "the provider subcommand is attached")
}

// set-provider-view-test literals, named to satisfy goconst across the package.
const (
	skillAcmeGoStyle  = "sauron-acme-go-style"
	agentAcmeReviewer = "sauron-acme-code-reviewer"
)

// TestRenderGroupInto asserts one named plan group: nothing written when
// empty, the heading plus one `~` line per name otherwise. This is what
// setProvider's handler composes inline for skills: and agents: — there is no
// longer a single renderSetProvider entrypoint to test the full assembly
// against; TestSetProviderMigrationFailureExitsOne and
// TestSetProviderEndToEnd already pin the full assembly (group + failure line
// + summary, in order) through the real command.
func TestRenderGroupInto(t *testing.T) {
	tests := []struct {
		name  string
		label string
		names []string
		want  string
	}{
		{name: "empty group writes nothing", label: "skills", names: nil, want: ""},
		{
			name:  "populated group writes the heading and each name",
			label: "skills",
			names: []string{skillAcmeGoStyle},
			want:  "skills:\n  ~ " + skillAcmeGoStyle + "\n",
		},
		{
			name:  "multiple names each render on their own line",
			label: "agents",
			names: []string{agentAcmeReviewer, "sauron-x"},
			want:  "agents:\n  ~ " + agentAcmeReviewer + "\n  ~ sauron-x\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer
			ew := newErrWriter(&buf)

			// Act.
			renderGroupInto(ew, tt.label, tt.names)

			// Assert.
			require.NoError(t, ew.toIOError("test"))
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestSummaryLine asserts the closing confirmation, with and without a
// migrated count.
func TestSummaryLine(t *testing.T) {
	tests := []struct {
		name   string
		result *usecase.SetProviderResponse
		want   string
	}{
		{
			name:   "no migration reports the plain confirmation",
			result: &usecase.SetProviderResponse{Provider: nameClaude},
			want:   "provider set to \"claude\"\n",
		},
		{
			name:   "a migration appends the artifact count",
			result: &usecase.SetProviderResponse{Provider: nameZencoder, Migrated: 2},
			want:   "provider set to \"zencoder\"; 2 artifacts migrated\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, summaryLine(tt.result))
		})
	}
}

// TestSetProviderWriteError drives the real command with a failing stdout,
// over the stranded-skill migration-failure fixture (skills heading, skill
// item, failure line, summary — see TestSetProviderMigrationFailureExitsOne)
// and the unchanged path, covering every reachable write point without a
// content-fetch fixture for a clean multi-kind migration.
func TestSetProviderWriteError(t *testing.T) {
	t.Run("change path", func(t *testing.T) {
		for _, writeAfter := range []int{0, 1, 2, 3} {
			// Arrange: HOME redirected so the provider filesystem stays off the
			// real FS; claude is active with one stranded skill (its file never
			// created, so the switch fails to migrate it).
			t.Setenv("HOME", t.TempDir())
			home := t.TempDir()
			t.Setenv("SAURON_HOME", home)
			_, err := runSetProvider(t, nameClaude)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(home, "track.yaml"), []byte(trackedSkillStream), 0o644))
			cmd := SetProvider()
			cmd.SetOut(&failingWriter{writeAfter: writeAfter})
			cmd.SetContext(context.Background())
			cmd.SetArgs([]string{"zencoder"})

			// Act.
			execErr := cmd.Execute()

			// Assert.
			var ucErr *usecase.Error
			require.ErrorAsf(t, execErr, &ucErr, "writeAfter=%d", writeAfter)
			assert.Equalf(t, usecase.TypeIO, ucErr.Type, "writeAfter=%d", writeAfter)
		}
	})

	t.Run("unchanged path", func(t *testing.T) {
		// Arrange.
		t.Setenv("SAURON_HOME", t.TempDir())
		_, err := runSetProvider(t, nameClaude)
		require.NoError(t, err)
		cmd := SetProvider()
		cmd.SetOut(&failingWriter{})
		cmd.SetContext(context.Background())
		cmd.SetArgs([]string{nameClaude})

		// Act.
		execErr := cmd.Execute()

		// Assert.
		var ucErr *usecase.Error
		require.ErrorAs(t, execErr, &ucErr)
		assert.Equal(t, usecase.TypeIO, ucErr.Type)
	})
}
