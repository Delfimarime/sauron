package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// describe-cmd test literals, named to satisfy goconst across the package.
const (
	flagFields   = "--fields"
	flagUnknown  = "--nope"
	transportSel = "transport,source"
	caseUnknown  = "rejects an unknown flag"

	labelTransport = "transport:"
	labelURI       = "source:"
	labelRef       = "revision:"
	labelAuth      = "credentials:"
	labelTimeout   = "timeout:"
)

// authRegistries is a schema-valid settings.yaml stream carrying an auth block,
// used to assert credential fields render as their stored env references.
const authRegistries = `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
spec:
  transport: git
  source: git@github.com:acme/artifacts.git
  revision: v1.2.0
  credentials:
    username: ${env:ACME_USER}
    password: ${env:ACME_TOKEN}
  timeout: 45s
`

// runDescribeRegistry assembles and runs the subcommand, returning stdout and the
// resulting error.
func runDescribeRegistry(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := DescribeRegistry()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

// TestDescribeGroup asserts the describe group has no run behaviour and attaches
// the registry subcommand.
func TestDescribeGroup(t *testing.T) {
	// Arrange + Act.
	cmd := Describe()

	// Assert.
	assert.Equal(t, "describe", cmd.Name())
	assert.Nil(t, cmd.RunE, "the group has no run behaviour")

	var registry *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == subcmdRegistry {
			registry = sub
		}
	}
	require.NotNil(t, registry, "the registry subcommand is attached")
}

// TestDescribeRegistryFlagSurface asserts the --fields flag and the argument
// validator are installed.
func TestDescribeRegistryFlagSurface(t *testing.T) {
	// Arrange + Act.
	cmd := DescribeRegistry()

	// Assert.
	assert.NotNil(t, cmd.Flags().Lookup(fieldsName), "flag fields registered")
	assert.NotNil(t, cmd.Args, "an argument validator is installed")
}

// TestDescribeRegistryRejectsBadInput asserts an unexpected argument or an
// unknown flag is rejected before the graph is built and maps to the usage exit
// code.
func TestDescribeRegistryRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: caseUnexpectedArg, args: []string{argExtra}},
		{name: caseUnknown, args: []string{flagUnknown}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			_, err := runDescribeRegistry(t, tt.args...)

			// Assert.
			require.Error(t, err)
			assert.Equal(t, exitUsage, exitCode(err))
		})
	}
}

// TestDescribeRegistryEndToEnd drives the assembled subcommand through the real fx
// graph against a seeded settings.yaml, covering the full detail, field
// selection, the credentials block, the not-found error, and the usage error. source is the
// identity and is always present and first; there is no name line.
func TestDescribeRegistryEndToEnd(t *testing.T) {
	tests := []struct {
		name         string
		seed         string
		args         []string
		wantOut      []string
		wantAbsent   []string
		wantErr      bool
		wantNotFound bool
		wantUsage    bool
	}{
		{
			name:    "full detail shows every populated field and the auth block",
			seed:    authRegistries,
			wantOut: []string{labelURI, "git@github.com:acme/artifacts.git", labelTransport, labelRef, "v1.2.0", labelAuth, "${env:ACME_USER}", "${env:ACME_TOKEN}", labelTimeout, "45s"},
		},
		{
			name:       "fields selects and orders, source forced first",
			seed:       authRegistries,
			args:       []string{flagFields, transportSel},
			wantOut:    []string{labelURI, labelTransport},
			wantAbsent: []string{labelRef, labelAuth, labelTimeout},
		},
		{
			name:         "no registry configured is a not-found runtime error",
			seed:         "",
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "invalid field is a usage error",
			seed:      authRegistries,
			args:      []string{flagFields, "bogus"},
			wantErr:   true,
			wantUsage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			seedRegistries(t, tt.seed)

			// Act.
			out, err := runDescribeRegistry(t, tt.args...)

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantUsage {
					assert.Equal(t, exitUsage, exitCode(err))
				}
				if tt.wantNotFound {
					assert.Equal(t, exitError, exitCode(err))
				}
				return
			}
			require.NoError(t, err)
			for _, want := range tt.wantOut {
				assert.Contains(t, out, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, out, absent)
			}
			assert.NotContains(t, strings.ToLower(out), "s3cr3t")
		})
	}
}
