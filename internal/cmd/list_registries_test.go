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
)

// twoRegistries is a schema-valid registries.yaml stream used to seed the
// end-to-end listing tests.
// acmeName is the registry name reused across the listing assertions; colURI is
// the repeated uri column literal.
const (
	acmeName = "acme"
	colURI   = "uri"
)

const twoRegistries = `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
spec:
  transport: git
  uri: git@github.com:acme/artifacts.git
---
apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: internal
spec:
  transport: http
  uri: https://reg.example.com/
`

// seedRegistries pins SAURON_HOME to a fresh temp dir and, when content is
// non-empty, writes it as registries.yaml there — so nothing durable is touched.
func seedRegistries(t *testing.T, content string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("SAURON_HOME", home)
	if content != "" {
		require.NoError(t, os.WriteFile(filepath.Join(home, "registries.yaml"), []byte(content), 0o644))
	}
}

// runListRegistries assembles and runs the subcommand, returning stdout and the
// resulting error.
func runListRegistries(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := ListRegistries()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// TestNewListRegistriesRequestMapsFlags asserts the parsed flags land on the use
// case request.
func TestNewListRegistriesRequestMapsFlags(t *testing.T) {
	// Arrange.
	var stdout bytes.Buffer
	flags := listingFlags{Search: acmeName, Sort: fieldTransport, Order: orderDesc, Fields: []string{sortName, colURI}}

	// Act.
	request := newListRegistriesRequest(context.Background(), &flags, &stdout)

	// Assert.
	require.NotNil(t, request)
	assert.Equal(t, acmeName, request.Search)
	assert.Equal(t, fieldTransport, request.Sort)
	assert.Equal(t, orderDesc, request.Order)
	assert.Equal(t, []string{sortName, colURI}, request.Fields)
	assert.Same(t, &stdout, request.Out())
}

// TestListGroup asserts the list group has no run behaviour and attaches the
// registries subcommand.
func TestListGroup(t *testing.T) {
	// Arrange + Act.
	cmd := List()

	// Assert.
	assert.Equal(t, "list", cmd.Name())
	assert.Nil(t, cmd.RunE, "the group has no run behaviour")

	var registries *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "registries" {
			registries = sub
		}
	}
	require.NotNil(t, registries, "the registries subcommand is attached")
}

// TestListRegistriesFlagSurface asserts the listing flags are registered with
// the documented default for --order.
func TestListRegistriesFlagSurface(t *testing.T) {
	// Arrange + Act.
	cmd := ListRegistries()

	// Assert.
	for _, name := range []string{flagSearch, flagSort, flagOrder, fieldsName} {
		assert.NotNilf(t, cmd.Flags().Lookup(name), "flag %q registered", name)
	}
	order, err := cmd.Flags().GetString(flagOrder)
	require.NoError(t, err)
	assert.Equal(t, "asc", order)
	assert.NotNil(t, cmd.Args, "an argument validator is installed")
}

// TestListRegistriesRejectsBadInput asserts a positional argument or an unknown
// flag is rejected before the graph is built and maps to the usage exit code.
func TestListRegistriesRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "rejects a positional argument", args: []string{argExtra}},
		{name: "rejects an unknown flag", args: []string{"--nope"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			_, err := runListRegistries(t, tt.args...)

			// Assert.
			require.Error(t, err)
			assert.Equal(t, exitUsage, exitCode(err))
		})
	}
}

// TestListRegistriesEndToEnd drives the assembled subcommand through the real fx
// graph against a seeded registries.yaml, covering the default columns, the
// filter, the sort, column selection, the empty result, and the usage error.
func TestListRegistriesEndToEnd(t *testing.T) {
	tests := []struct {
		name       string
		seed       string
		args       []string
		wantOut    []string
		wantAbsent []string
		wantOrder  []string
		wantEmpty  bool
		wantErr    bool
		wantUsage  bool
	}{
		{
			name:      "default columns list all sorted by name",
			seed:      twoRegistries,
			wantOut:   []string{"NAME", "TRANSPORT", "URI", "git@github.com:acme/artifacts.git"},
			wantOrder: []string{acmeName, "internal"},
		},
		{
			name:      "search filters by name case-insensitively",
			seed:      twoRegistries,
			args:      []string{"--search", "ACME"},
			wantOrder: []string{acmeName},
		},
		{
			name:      "sort by transport descending",
			seed:      twoRegistries,
			args:      []string{"--sort", fieldTransport, "--order", orderDesc},
			wantOrder: []string{"internal", acmeName},
		},
		{
			name:       "fields selects and orders columns",
			seed:       twoRegistries,
			args:       []string{"--fields", "name,uri"},
			wantOut:    []string{"NAME", "URI"},
			wantAbsent: []string{"TRANSPORT"},
		},
		{
			name:      "empty result writes nothing",
			seed:      "",
			wantEmpty: true,
		},
		{
			name:      "invalid sort field is a usage error",
			seed:      twoRegistries,
			args:      []string{"--sort", colURI},
			wantErr:   true,
			wantUsage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			seedRegistries(t, tt.seed)

			// Act.
			out, err := runListRegistries(t, tt.args...)

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantUsage {
					assert.Equal(t, exitUsage, exitCode(err))
				}
				return
			}
			require.NoError(t, err)
			if tt.wantEmpty {
				assert.Empty(t, strings.TrimSpace(out))
				return
			}
			for _, want := range tt.wantOut {
				assert.Contains(t, out, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, out, absent)
			}
			if len(tt.wantOrder) > 0 {
				assert.Equal(t, tt.wantOrder, nameColumn(out))
			}
		})
	}
}

// nameColumn returns the first column down the data rows (the registry names),
// skipping the header.
func nameColumn(out string) []string {
	var names []string
	for i, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if i == 0 {
			continue // header row
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		names = append(names, fields[0])
	}
	return names
}
