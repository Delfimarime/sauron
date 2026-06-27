package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

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
