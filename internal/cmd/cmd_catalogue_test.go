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
)

// shared test-data literals reused across the cmd package's tests, named to
// satisfy goconst across the package.
const (
	argExtra      = "extra"
	flagOrder     = "order"
	flagSearch    = "search"
	flagSort      = "sort"
	fieldsName    = "fields"
	flagTransport = "transport"
	sortName      = "name"
)

// seedCatalogueRegistry pins SAURON_HOME to a fresh temp dir, materializes a
// filesystem-backed registry source holding one agent manifest, and records the
// single registry in settings.yaml — so nothing durable is touched.
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
		"\nspec:\n  transport: filesystem\n  source: " + source + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(home, settingsFile), []byte(stream), 0o644))
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

// TestCatalogueGroup asserts the catalogue group is a pure command group: a bare
// invocation prints help and exits 0 (no RunE), and the per-kind subcommands are
// attached.
func TestCatalogueGroup(t *testing.T) {
	// Arrange.
	cmd := Catalogue()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(nil)

	// Act: invoking the group without a kind noun.
	err := cmd.Execute()

	// Assert: a group with no run behaviour succeeds and prints its help.
	assert.Equal(t, subcmdCatalogue, cmd.Name())
	assert.Nil(t, cmd.RunE, "the group has no run behaviour")
	require.NoError(t, err)
	assert.Equal(t, exitOK, exitCode(err))

	names := map[string]bool{}
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, want := range []string{subcmdSkill, subcmdAgent} {
		assert.Truef(t, names[want], "the %q subcommand is attached", want)
	}
}

// TestListGroupAttachesCatalogue asserts the list group attaches the catalogue
// group.
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
		subcmdSkill: ListCatalogueSkill,
		subcmdAgent: ListCatalogueAgent,
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

// TestNewListCatalogueInputMapsFlags asserts the kind and the parsed flags land
// on the use case input.
func TestNewListCatalogueInputMapsFlags(t *testing.T) {
	// Arrange.
	flags := catalogueFlags{Search: "rev", Sort: sortName, Order: orderDesc, paging: pagingFlags{Page: 2, Limit: 5}}

	// Act.
	input, err := newListCatalogueInput(usecase.CatalogueSkill, &flags)

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, usecase.CatalogueSkill, input.Kind)
	assert.Equal(t, "rev", input.Search)
	assert.Equal(t, sortName, input.Sort)
	assert.Equal(t, orderDesc, input.Order)
	assert.Equal(t, int64(2), input.Page)
	assert.Equal(t, int64(5), input.Limit)
}

// TestNewListCatalogueInputDefaults asserts an empty sort/order resolves to the
// name/asc defaults at the handler boundary before the use case runs.
func TestNewListCatalogueInputDefaults(t *testing.T) {
	// Arrange.
	flags := catalogueFlags{paging: pagingFlags{Page: defaultPage, Limit: defaultLimit}}

	// Act.
	input, err := newListCatalogueInput(usecase.CatalogueAgent, &flags)

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, sortName, input.Sort)
	assert.Equal(t, "asc", input.Order)
}

// TestListCatalogueRejectsBadInput asserts an unexpected argument, an unknown
// flag, or an out-of-range page is rejected as a usage error (exit 2): the first
// two by cobra, the page by the use case.
func TestListCatalogueRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: caseUnexpectedArg, args: []string{argExtra}},
		{name: caseUnknown, args: []string{flagUnknown}},
		{name: "rejects a page below 1", args: []string{"--page", "0"}},
		{name: "rejects an unknown order", args: []string{"--order", "sideways"}},
		{name: "rejects an unknown sort", args: []string{"--sort", "size"}},
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
// graph against a seeded filesystem registry, covering a populated page and the
// no-registry runtime error.
func TestListCatalogueEndToEnd(t *testing.T) {
	t.Run("lists the agents with the paging line", func(t *testing.T) {
		// Arrange.
		seedCatalogueRegistry(t)

		// Act.
		out, err := runListCatalogueAgent(t)

		// Assert.
		require.NoError(t, err)
		for _, want := range []string{"code-reviewer", "agent", "showing 1"} {
			assert.Contains(t, out, want)
		}
	})

	t.Run("no registry configured is a not-found runtime error", func(t *testing.T) {
		// Arrange: an empty home, no settings.yaml.
		t.Setenv("SAURON_HOME", t.TempDir())

		// Act.
		_, err := runListCatalogueAgent(t)

		// Assert.
		require.Error(t, err)
		assert.Equal(t, exitError, exitCode(err))
	})
}

// TestListCatalogueUnreachableSource asserts a registry whose filesystem source
// is absent fails as a runtime error (exit 1), not a usage error.
func TestListCatalogueUnreachableSource(t *testing.T) {
	// Arrange.
	home := t.TempDir()
	t.Setenv("SAURON_HOME", home)
	stream := "apiVersion: sauron.raitonbl.com/v1\nkind: Registry\nmetadata:\n  name: " + acmeName +
		"\nspec:\n  transport: filesystem\n  source: " + filepath.Join(home, "nonexistent") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(home, settingsFile), []byte(stream), 0o644))

	// Act.
	_, err := runListCatalogueAgent(t)

	// Assert.
	require.Error(t, err)
	assert.Equal(t, exitError, exitCode(err))
	assert.NotEqual(t, exitUsage, exitCode(err))
}
