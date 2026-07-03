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

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// list-artifacts-cmd test literals, named to satisfy goconst across the
// package.
const (
	subcmdSkills = "skills"
	subcmdAgents = "agents"
	flagPage     = "--page"
	nameGoStyle  = "go-style"
)

// twoSkillsTrack is a track.yaml stream carrying two installed skills, used to
// assert listing, ordering, and paging.
const twoSkillsTrack = `apiVersion: sauron.raitonbl.com/v1
kind: Skill
metadata:
  name: go-style
spec:
  version: v1.4.0
  path: skills/go-style
  installedAt: "2026-06-15T00:00:00Z"
  updatedAt: "2026-06-15T00:00:00Z"
---
apiVersion: sauron.raitonbl.com/v1
kind: Skill
metadata:
  name: sql-review
spec:
  version: v2.0.0
  path: skills/sql-review
  installedAt: "2026-06-12T00:00:00Z"
  updatedAt: "2026-06-12T00:00:00Z"
`

// malformedTrack is a track.yaml stream that fails to parse as YAML.
const malformedTrack = "not: [valid"

// tableFields reads the given zero-based line of out as whitespace-separated
// fields, so an assertion on column order/content is not defeated by the
// tabwriter's variable column padding.
func tableFields(t *testing.T, out string, line int) []string {
	t.Helper()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	require.Greaterf(t, len(lines), line, "output has a line %d: %q", line, out)
	return strings.Fields(lines[line])
}

// seedTrack pins SAURON_HOME to a fresh temp dir and writes stream as the
// track.yaml state, so the artifact-listing commands resolve it without
// touching anything durable. An empty stream leaves nothing installed.
func seedTrack(t *testing.T, stream string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("SAURON_HOME", home)
	if stream == "" {
		return
	}
	require.NoError(t, os.WriteFile(filepath.Join(home, "track.yaml"), []byte(stream), 0o644))
}

// runListSkills assembles and runs the skills subcommand, returning stdout and
// the resulting error.
func runListSkills(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := ListSkills()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// runListAgents assembles and runs the agents subcommand, returning stdout and
// the resulting error.
func runListAgents(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := ListAgents()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// TestListGroupAttachesArtifacts asserts the list group attaches the skills and
// agents leaves directly (flat, not nested under a further sub-group).
func TestListGroupAttachesArtifacts(t *testing.T) {
	// Arrange + Act.
	cmd := List()

	// Assert.
	names := map[string]bool{}
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, want := range []string{subcmdSkills, subcmdAgents} {
		assert.Truef(t, names[want], "the %q subcommand is attached", want)
	}
}

// TestListArtifactsFlagSurface asserts each leaf registers the filter, sort,
// paging, and fields flags with their documented defaults, plus an argument
// validator.
func TestListArtifactsFlagSurface(t *testing.T) {
	builders := map[string]func() *cobra.Command{
		subcmdSkills: ListSkills,
		subcmdAgents: ListAgents,
	}

	for name, build := range builders {
		t.Run(name, func(t *testing.T) {
			// Act.
			cmd := build()

			// Assert.
			for _, flag := range []string{flagSearch, flagSort, flagOrder, "page", "limit", fieldsName} {
				assert.NotNilf(t, cmd.Flags().Lookup(flag), "flag %q registered", flag)
			}
			order, err := cmd.Flags().GetString(flagOrder)
			require.NoError(t, err)
			assert.Equal(t, "asc", order)
			page, err := cmd.Flags().GetInt64("page")
			require.NoError(t, err)
			assert.Equal(t, int64(defaultPage), page)
			limit, err := cmd.Flags().GetInt64("limit")
			require.NoError(t, err)
			assert.Equal(t, int64(defaultLimit), limit)
			assert.NotNil(t, cmd.Args, "an argument validator is installed")
		})
	}
}

// TestListArtifactsRejectsBadInput asserts an unexpected argument, an unknown
// flag, an unknown --fields/--sort/--order value, or an out-of-range
// --page/--limit is a usage error (exit 2, FR-008), resolved before track.yaml
// is ever read (no seeding here).
func TestListArtifactsRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: caseUnexpectedArg, args: []string{argExtra}},
		{name: caseUnknown, args: []string{flagUnknown}},
		{name: "rejects a page below 1", args: []string{flagPage, "0"}},
		{name: "rejects a limit below 1", args: []string{"--limit", "0"}},
		{name: "rejects an unknown sort", args: []string{"--sort", nameBogus}},
		{name: "rejects an unknown order", args: []string{"--order", "sideways"}},
		{name: "rejects an unknown field", args: []string{"--fields", nameBogus}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			t.Setenv("SAURON_HOME", t.TempDir())

			// Act.
			_, err := runListSkills(t, tt.args...)

			// Assert.
			require.Error(t, err)
			assert.Equal(t, exitUsage, ExitCode(err))
		})
	}
}

// TestListArtifactsEndToEnd drives the assembled skills/agents subcommands
// through the real fx graph against a seeded track.yaml, covering the default
// name-only view, --fields reordering/selecting columns (name forced first),
// the em-dash placeholder for an empty version, --search/--sort/--order/--page/
// --limit reaching the use case, the empty-result view (no row, only the paging
// line), and a malformed track.yaml surfacing as a runtime error.
func TestListArtifactsEndToEnd(t *testing.T) {
	t.Run("skills: default view shows only name", func(t *testing.T) {
		// Arrange.
		seedTrack(t, twoSkillsTrack)

		// Act.
		out, err := runListSkills(t)

		// Assert.
		require.NoError(t, err)
		assert.Contains(t, out, nameGoStyle)
		assert.Contains(t, out, "sql-review")
		assert.NotContains(t, out, "v1.4.0")
		assert.Contains(t, out, "showing 1")
	})

	t.Run("skills: --fields reorders and selects columns, name forced first", func(t *testing.T) {
		// Arrange.
		seedTrack(t, twoSkillsTrack)

		// Act: request lastUpdatedAt before version; name is still forced first.
		out, err := runListSkills(t, "--fields", "lastUpdatedAt,version")

		// Assert: header and row carry the requested column order (name forced
		// first), matched on whitespace-separated fields since the tabwriter
		// aligns columns with variable padding.
		require.NoError(t, err)
		assert.Equal(t, []string{tableHeaderName, artifactFieldLastUpdatedAt, artifactFieldVersion}, tableFields(t, out, 0))
		assert.Equal(t, []string{nameGoStyle, "2026-06-15T00:00:00Z", "v1.4.0"}, tableFields(t, out, 1))
	})

	t.Run("skills: --page and --limit page the results", func(t *testing.T) {
		// Arrange.
		seedTrack(t, twoSkillsTrack)

		// Act.
		out, err := runListSkills(t, flagPage, "2", "--limit", "1")

		// Assert.
		require.NoError(t, err)
		assert.Contains(t, out, "sql-review")
		assert.NotContains(t, out, nameGoStyle)
		assert.Contains(t, out, "showing 2–2 (page 2, limit 1)")
	})

	t.Run("skills: --search filters by name substring", func(t *testing.T) {
		// Arrange.
		seedTrack(t, twoSkillsTrack)

		// Act.
		out, err := runListSkills(t, "--search", "sql")

		// Assert.
		require.NoError(t, err)
		assert.Contains(t, out, "sql-review")
		assert.NotContains(t, out, nameGoStyle)
	})

	t.Run("skills: --sort lastUpdatedAt --order desc orders before paging", func(t *testing.T) {
		// Arrange.
		seedTrack(t, twoSkillsTrack)

		// Act.
		out, err := runListSkills(t, "--sort", "lastUpdatedAt", "--order", "desc", "--limit", "1")

		// Assert.
		require.NoError(t, err)
		assert.Contains(t, out, nameGoStyle)
		assert.NotContains(t, out, "sql-review")
	})

	t.Run("skills: nothing installed shows no row, only the paging line", func(t *testing.T) {
		// Arrange.
		t.Setenv("SAURON_HOME", t.TempDir())

		// Act.
		out, err := runListSkills(t)

		// Assert.
		require.NoError(t, err)
		assert.Equal(t, "showing 0 results (page 1, limit 20)\n", out)
	})

	t.Run("skills: malformed track.yaml is a runtime error", func(t *testing.T) {
		// Arrange.
		seedTrack(t, malformedTrack)

		// Act.
		_, err := runListSkills(t)

		// Assert.
		require.Error(t, err)
		assert.Equal(t, exitError, ExitCode(err))
		assert.NotEqual(t, exitUsage, ExitCode(err))
	})

	t.Run("agents: default view shows only name", func(t *testing.T) {
		// Arrange.
		home := t.TempDir()
		t.Setenv("SAURON_HOME", home)
		stream := `apiVersion: sauron.raitonbl.com/v1
kind: Agent
metadata:
  name: code-reviewer
spec:
  version: 3af1c2e
  path: agents/code-reviewer
  installedAt: "2026-06-14T00:00:00Z"
  updatedAt: "2026-06-14T00:00:00Z"
`
		require.NoError(t, os.WriteFile(filepath.Join(home, "track.yaml"), []byte(stream), 0o644))

		// Act.
		out, err := runListAgents(t)

		// Assert.
		require.NoError(t, err)
		assert.Contains(t, out, "code-reviewer")
		assert.NotContains(t, out, "3af1c2e")
	})
}

// TestListArtifactsWriteError drives the real command with a failing stdout,
// covering both the table write (a populated page) and the paging-line write
// alone (a page past the seeded data, so the table is empty and never
// written) — both surface as a classified io error.
func TestListArtifactsWriteError(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// args are the extra flags beyond the bare command.
		args []string
	}{
		{name: "table write fails"},
		{name: "paging-line write fails on an empty page", args: []string{flagPage, "99"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			seedTrack(t, twoSkillsTrack)
			cmd := ListSkills()
			cmd.SetOut(&failingWriter{})
			cmd.SetContext(context.Background())
			cmd.SetArgs(tt.args)

			// Act.
			err := cmd.Execute()

			// Assert.
			var ucErr *usecase.Error
			require.ErrorAs(t, err, &ucErr)
			assert.Equal(t, usecase.TypeIO, ucErr.Type)
		})
	}
}

// TestSelectArtifactFields covers the default (name-only, FR-002), the
// selected/reordered request with name forced first, and the unknown-field
// usage error paths of the view's field selector.
func TestSelectArtifactFields(t *testing.T) {
	t.Run("empty request yields name only", func(t *testing.T) {
		got, err := selectArtifactFields(nil)
		require.NoError(t, err)
		assert.Equal(t, []string{tableHeaderName}, got)
	})

	t.Run("selection reorders, name forced first", func(t *testing.T) {
		got, err := selectArtifactFields([]string{artifactFieldLastUpdatedAt, artifactFieldVersion})
		require.NoError(t, err)
		assert.Equal(t, []string{tableHeaderName, artifactFieldLastUpdatedAt, artifactFieldVersion}, got)
	})

	t.Run("unknown field is a usage error", func(t *testing.T) {
		got, err := selectArtifactFields([]string{nameBogus})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, errInvalidFlag)
	})
}

// TestNewListArtifactsInputMapsFlags asserts the kind and the parsed flags
// land on the use case input, and that an empty sort/order resolves to the
// name/asc defaults at the handler boundary before the use case runs.
func TestNewListArtifactsInputMapsFlags(t *testing.T) {
	t.Run("maps every flag", func(t *testing.T) {
		flags := listFlags{Search: "rev", Sort: artifactFieldLastUpdatedAt, Order: orderDesc, paging: pagingFlags{Page: 2, Limit: 5}}

		in, err := newListArtifactsInput(types.KindSkill, &flags)

		require.NoError(t, err)
		assert.Equal(t, types.KindSkill, in.Kind)
		assert.Equal(t, "rev", in.Search)
		assert.Equal(t, artifactFieldLastUpdatedAt, in.Sort)
		assert.Equal(t, orderDesc, in.Order)
		assert.Equal(t, int64(2), in.Page)
		assert.Equal(t, int64(5), in.Limit)
	})

	t.Run("defaults sort to name and order to asc", func(t *testing.T) {
		flags := listFlags{paging: pagingFlags{Page: defaultPage, Limit: defaultLimit}}

		in, err := newListArtifactsInput(types.KindAgent, &flags)

		require.NoError(t, err)
		assert.Equal(t, tableHeaderName, in.Sort)
		assert.Equal(t, "asc", in.Order)
	})

	t.Run("rejects an unknown sort field", func(t *testing.T) {
		flags := listFlags{Sort: nameBogus, paging: pagingFlags{Page: defaultPage, Limit: defaultLimit}}

		_, err := newListArtifactsInput(types.KindSkill, &flags)

		require.Error(t, err)
		assert.ErrorIs(t, err, errInvalidFlag)
	})
}

// TestArtifactRowEmptyVersionRendersEmDash asserts the rendering path this
// command owns (buildTable + the shared cells substitution) renders the
// em-dash placeholder for an artifact with an empty version (FR-002),
// exercised directly against a constructed types.Artifact rather than through
// the track store, whose schema validates spec.version as non-empty.
func TestArtifactRowEmptyVersionRendersEmDash(t *testing.T) {
	// Arrange.
	artifact := types.Artifact{
		Metadata: types.Metadata{Name: nameGoStyle},
		Spec:     types.ArtifactSpec{Version: "", UpdatedAt: "2026-06-15T00:00:00Z"},
	}
	selected, err := selectArtifactFields([]string{artifactFieldVersion})
	require.NoError(t, err)

	// Act.
	rows := buildTable(selected, []types.Artifact{artifact}, func(a types.Artifact) []string {
		return artifactRow(a, selected)
	})
	var buf bytes.Buffer
	require.NoError(t, rows.render(&buf))

	// Assert.
	assert.Equal(t, []string{nameGoStyle, "—"}, tableFields(t, buf.String(), 1))
}
