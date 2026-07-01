package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
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
			assert.Equal(t, exitUsage, ExitCode(err))
		})
	}
}

// describeRegistryEndToEndCase is a single row in TestDescribeRegistryEndToEnd.
type describeRegistryEndToEndCase struct {
	name         string
	seed         string
	args         []string
	wantOut      []string
	wantAbsent   []string
	wantErr      bool
	wantNotFound bool
	wantUsage    bool
}

// assertDescribeRegistryEndToEnd checks one table row so TestDescribeRegistryEndToEnd
// stays below the gocognit ≤15 ceiling.
func assertDescribeRegistryEndToEnd(t *testing.T, tt describeRegistryEndToEndCase, out string, err error) {
	t.Helper()
	if tt.wantErr {
		require.Error(t, err)
		if tt.wantUsage {
			assert.Equal(t, exitUsage, ExitCode(err))
		}
		if tt.wantNotFound {
			assert.Equal(t, exitError, ExitCode(err))
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
}

// TestDescribeRegistryEndToEnd drives the assembled subcommand through the real fx
// graph against a seeded settings.yaml, covering the full detail, field
// selection, the credentials block, the not-found error, and the usage error. source is the
// identity and is always present and first; there is no name line.
func TestDescribeRegistryEndToEnd(t *testing.T) {
	tests := []describeRegistryEndToEndCase{
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
			assertDescribeRegistryEndToEnd(t, tt, out, err)
		})
	}
}

// describe-view-test literals, named to satisfy goconst across the package.
const (
	labelTLS     = "tls:"
	labelSSHKey  = "sshKey:"
	labelCreated = "createdAt:"
	labelUpdated = "lastUpdatedAt:"
	vUserRef     = "${env:ACME_USER}"
	vTokenRef    = "${env:ACME_TOKEN}"
	vGitURI      = "git@github.com:acme/artifacts.git"
	vRefV120     = "v1.2.0"
	v45s         = "45s"
	createdStamp = "2026-06-21T07:30:00Z"
	updatedStamp = "2026-06-22T08:00:00Z"
)

// allDescribeFields is the full, ordered field selection a default describe
// yields.
func allDescribeFields() []string {
	return []string{
		describeFieldSource, describeFieldTransport, describeFieldRevision,
		describeFieldCredentials, describeFieldTLS, describeFieldSSHKey, describeFieldTimeout,
		describeFieldCreated, describeFieldUpdated,
	}
}

// fullViewRegistry is a registry populated across every describable field.
func fullViewRegistry() types.Registry {
	return types.Registry{
		Metadata: types.Metadata{
			CreatedAt:     createdStamp,
			LastUpdatedAt: updatedStamp,
		},
		Spec: types.RegistrySpec{
			Transport:   types.TransportGit,
			Source:      vGitURI,
			Revision:    vRefV120,
			Credentials: &types.Credentials{Username: vUserRef, Password: vTokenRef},
			Timeout:     v45s,
		},
	}
}

// TestRenderDescribeRegistry covers the projection + descriptor rendering across
// the default view, field selection, the nested auth/tls blocks, and omission of
// unpopulated fields. uri is the identity and is always present and first.
func TestRenderDescribeRegistry(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// registry is the record to project.
		registry types.Registry
		// fields is the resolved, ordered field selection.
		fields []string
		// wantContains are substrings the output must contain, in order.
		wantContains []string
		// wantAbsent are substrings the output must never contain.
		wantAbsent []string
	}{
		{
			name:     "default shows every populated field",
			registry: fullViewRegistry(),
			fields:   allDescribeFields(),
			wantContains: []string{
				labelURI, vGitURI,
				labelTransport,
				labelRef, vRefV120,
				labelAuth,
				"username:", vUserRef,
				"password:", vTokenRef,
				labelTimeout, v45s,
				labelCreated, createdStamp,
				labelUpdated, updatedStamp,
			},
		},
		{
			name:         "default omits unpopulated fields",
			registry:     types.Registry{Spec: types.RegistrySpec{Transport: types.TransportGit, Source: "u"}},
			fields:       allDescribeFields(),
			wantContains: []string{labelURI, labelTransport},
			wantAbsent:   []string{labelRef, labelAuth, labelTLS, labelSSHKey, labelTimeout, labelCreated, labelUpdated},
		},
		{
			name:         "fields projects and orders, uri forced first",
			registry:     fullViewRegistry(),
			fields:       []string{describeFieldSource, describeFieldTransport, describeFieldRevision},
			wantContains: []string{labelURI, labelTransport, labelRef},
			wantAbsent:   []string{labelAuth, labelTimeout},
		},
		{
			name:         "auth renders the stored env references, never a secret",
			registry:     fullViewRegistry(),
			fields:       []string{describeFieldSource, describeFieldCredentials},
			wantContains: []string{labelAuth, vUserRef, vTokenRef},
			wantAbsent:   []string{"s3cr3t"},
		},
		{
			name: "tls and sshKey render their populated sub-fields",
			registry: types.Registry{
				Spec: types.RegistrySpec{
					Transport: types.TransportHTTP,
					Source:    "u",
					SSHKey:    "/home/dev/.ssh/id_ed25519",
					TLS: &types.TLS{
						SkipVerify: true,
						CACert:     "/etc/ssl/ca.pem",
						ClientCert: "/etc/ssl/client.pem",
						ClientKey:  "/etc/ssl/client.key",
					},
				},
			},
			fields: []string{describeFieldSource, describeFieldTLS, describeFieldSSHKey},
			wantContains: []string{
				labelTLS,
				"skipVerify: true",
				"caCert:", "/etc/ssl/ca.pem",
				"clientCert:", "/etc/ssl/client.pem",
				"clientKey:", "/etc/ssl/client.key",
				labelSSHKey, "/home/dev/.ssh/id_ed25519",
			},
		},
		{
			name: "an empty tls block is omitted",
			registry: types.Registry{
				Spec: types.RegistrySpec{Transport: types.TransportHTTP, Source: "u", TLS: &types.TLS{}},
			},
			fields:       []string{describeFieldSource, describeFieldTLS},
			wantContains: []string{labelURI},
			wantAbsent:   []string{labelTLS},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer
			registry := tt.registry

			// Act.
			err := renderDescribeRegistry(&buf, &registry, tt.fields)

			// Assert.
			require.NoError(t, err)
			out := buf.String()
			lastIndex := -1
			for _, want := range tt.wantContains {
				idx := strings.Index(out, want)
				require.GreaterOrEqualf(t, idx, 0, "output %q missing %q", out, want)
				assert.Greaterf(t, idx, lastIndex, "%q is out of order in %q", want, out)
				lastIndex = idx
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContainsf(t, out, absent, "output unexpectedly contains %q", absent)
			}
		})
	}
}

// TestRenderDescribeRegistryWriteError surfaces a writer failure as an io error.
func TestRenderDescribeRegistryWriteError(t *testing.T) {
	// Arrange.
	registry := fullViewRegistry()

	// Act.
	err := renderDescribeRegistry(&failingWriter{}, &registry, allDescribeFields())

	// Assert.
	var ucErr *usecase.Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, usecase.TypeIO, ucErr.Type)
}

// TestSelectDescribeFields covers the default, identity-first ordering, dedupe,
// and unknown-field paths of the view's field selector. An unknown field is a
// usage error raised at the command boundary.
func TestSelectDescribeFields(t *testing.T) {
	t.Run("empty request yields every field in order", func(t *testing.T) {
		got, err := selectDescribeFields(nil)
		require.NoError(t, err)
		assert.Equal(t, allDescribeFields(), got)
	})

	t.Run("selection forces uri present and first, deduped", func(t *testing.T) {
		got, err := selectDescribeFields([]string{describeFieldTransport, describeFieldRevision, describeFieldTransport})
		require.NoError(t, err)
		assert.Equal(t, []string{describeFieldSource, describeFieldTransport, describeFieldRevision}, got)
	})

	t.Run("unknown field is a usage error", func(t *testing.T) {
		got, err := selectDescribeFields([]string{"bogus"})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, errInvalidFlag)
	})
}
