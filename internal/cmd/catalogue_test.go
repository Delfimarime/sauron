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
)

// catalogue-cmd test literals, named to satisfy goconst across the package.
const (
	subcmdCatalogue = "catalogue"
	subcmdSkill     = "skill"
	subcmdAgent     = "agent"
	subcmdPersona   = "persona"
)

// shared test-data literals reused across the cmd package's tests, named to
// satisfy goconst across the package.
const (
	argGhost   = "ghost"
	argExtra   = "extra"
	flagOrder  = "order"
	flagSearch = "search"
	flagSort   = "sort"
	fieldsName = "fields"
	flagKind   = "kind"
	orderDesc  = "desc"
	sortName   = "name"
)

// seedCatalogueRegistry pins SAURON_HOME to a fresh temp dir, materializes a
// filesystem-backed registry source holding one agent manifest, and records the
// registry in registries.yaml — so nothing durable is touched.
func seedCatalogueRegistry(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("SAURON_HOME", home)

	source := filepath.Join(home, "source")
	agents := filepath.Join(source, ".agents")
	require.NoError(t, os.MkdirAll(agents, 0o755))
	manifest := "apiVersion: sauron.raitonbl.com/v1\nkind: Agent\nmetadata:\n  name: code-reviewer\n"
	require.NoError(t, os.WriteFile(filepath.Join(agents, "code-reviewer.yaml"), []byte(manifest), 0o644))

	stream := "apiVersion: sauron.raitonbl.com/v1\nkind: Registry\nmetadata:\n  name: " + acmeName +
		"\nspec:\n  transport: filesystem\n  uri: " + source + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(home, "registries.yaml"), []byte(stream), 0o644))
}

// runListCatalogueAgent assembles and runs the agent subcommand, returning stdout
// and the resulting error.
func runListCatalogueAgent(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := ListCatalogueAgent()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// TestCatalogueGroup asserts the catalogue group has no run behaviour and attaches
// the three per-kind subcommands.
func TestCatalogueGroup(t *testing.T) {
	// Arrange + Act.
	cmd := Catalogue()

	// Assert.
	assert.Equal(t, subcmdCatalogue, cmd.Name())
	assert.Nil(t, cmd.RunE, "the group has no run behaviour")

	names := map[string]bool{}
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, want := range []string{subcmdSkill, subcmdAgent, subcmdPersona} {
		assert.Truef(t, names[want], "the %q subcommand is attached", want)
	}
}

// TestListGroupAttachesCatalogue asserts the list group attaches the catalogue
// group alongside registries.
func TestListGroupAttachesCatalogue(t *testing.T) {
	// Arrange + Act.
	cmd := List()

	// Assert.
	var catalogue *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == subcmdCatalogue {
			catalogue = sub
		}
	}
	require.NotNil(t, catalogue, "the catalogue subcommand is attached")
}

// TestCatalogueFlagSurface asserts each leaf registers the filter, sort, and
// paging flags with their documented defaults, plus an argument validator.
func TestCatalogueFlagSurface(t *testing.T) {
	builders := map[string]func() *cobra.Command{
		subcmdSkill:   ListCatalogueSkill,
		subcmdAgent:   ListCatalogueAgent,
		subcmdPersona: ListCataloguePersona,
	}

	for name, build := range builders {
		t.Run(name, func(t *testing.T) {
			// Act.
			cmd := build()

			// Assert.
			for _, flag := range []string{flagSearch, flagSort, flagOrder, "page", "limit"} {
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

// TestNewListCatalogueRequestMapsArgs asserts the kind, the positional registry,
// and the parsed flags land on the use case request.
func TestNewListCatalogueRequestMapsArgs(t *testing.T) {
	// Arrange.
	var stdout bytes.Buffer
	flags := catalogueFlags{Search: "rev", Sort: sortName, Order: orderDesc, paging: pagingFlags{Page: 2, Limit: 5}}

	// Act.
	request := newListCatalogueRequest(context.Background(), usecase.CatalogueSkill, &flags, []string{acmeName}, &stdout)

	// Assert.
	require.NotNil(t, request)
	assert.Equal(t, usecase.CatalogueSkill, request.Kind)
	assert.Equal(t, acmeName, request.Registry)
	assert.Equal(t, "rev", request.Search)
	assert.Equal(t, sortName, request.Sort)
	assert.Equal(t, orderDesc, request.Order)
	assert.Equal(t, int64(2), request.Page)
	assert.Equal(t, int64(5), request.Limit)
	assert.Same(t, &stdout, request.Out())
}

// TestListCatalogueRejectsBadInput asserts a missing registry, an unknown flag,
// or an out-of-range page is rejected as a usage error (exit 2): the first two by
// cobra, the page by the use case.
func TestListCatalogueRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "rejects a missing registry", args: nil},
		{name: "rejects an extra argument", args: []string{acmeName, argExtra}},
		{name: caseUnknown, args: []string{acmeName, flagUnknown}},
		{name: "rejects a page below 1", args: []string{acmeName, "--page", "0"}},
		{name: "rejects an unknown order", args: []string{acmeName, "--order", "sideways"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			seedCatalogueRegistry(t)

			// Act.
			_, err := runListCatalogueAgent(t, tt.args...)

			// Assert.
			require.Error(t, err)
			assert.Equal(t, exitUsage, exitCode(err))
		})
	}
}

// TestListCatalogueEndToEnd drives the assembled subcommand through the real fx
// graph against a seeded filesystem registry, covering a populated page, the
// not-found runtime error, and the unreachable-source runtime error.
func TestListCatalogueEndToEnd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantOut   []string
		wantErr   bool
		wantError bool
	}{
		{
			name:    "lists the agents with the paging line",
			args:    []string{acmeName},
			wantOut: []string{"code-reviewer", "agent", "showing 1"},
		},
		{
			name:      "unknown registry is a not-found runtime error",
			args:      []string{argGhost},
			wantErr:   true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			seedCatalogueRegistry(t)

			// Act.
			out, err := runListCatalogueAgent(t, tt.args...)

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantError {
					assert.Equal(t, exitError, exitCode(err))
				}
				return
			}
			require.NoError(t, err)
			for _, want := range tt.wantOut {
				assert.Contains(t, out, want)
			}
		})
	}
}

// TestListCatalogueUnreachableSource asserts a registry whose filesystem source
// is absent fails as a runtime error (exit 1), not a usage error.
func TestListCatalogueUnreachableSource(t *testing.T) {
	// Arrange.
	home := t.TempDir()
	t.Setenv("SAURON_HOME", home)
	stream := "apiVersion: sauron.raitonbl.com/v1\nkind: Registry\nmetadata:\n  name: " + acmeName +
		"\nspec:\n  transport: filesystem\n  uri: " + filepath.Join(home, "nonexistent") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(home, "registries.yaml"), []byte(stream), 0o644))

	// Act.
	_, err := runListCatalogueAgent(t, acmeName)

	// Assert.
	require.Error(t, err)
	assert.Equal(t, exitError, exitCode(err))
	assert.NotEqual(t, exitUsage, exitCode(err))
}
