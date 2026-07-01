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
	"github.com/delfimarime/sauron/pkg/sauron/types"
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
