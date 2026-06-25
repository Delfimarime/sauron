package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// registry-test literals, named to satisfy goconst across the package.
const (
	acmeURI     = "git@github.com:acme/artifacts.git"
	internalURI = "https://reg.example.com/"
	internal    = "internal"
	valTokenRef = "${env:ACME_TOKEN}"
	valRefV120  = "v1.2.0"
	valTimeout  = "45s"
	valBogus    = "bogus"

	fldName      = "name:"
	fldTransport = "transport:"
	fldURI       = "uri:"
	fldRef       = "ref:"
	fldAuth      = "auth:"
	fldTimeout   = "timeout:"
	fldTLS       = "tls:"
)

// listing builds a registry with the given name, transport, and uri.
func listing(name string, transport types.Transport, uri string) types.Registry {
	return types.Registry{
		Metadata: types.Metadata{Name: name},
		Spec:     types.RegistrySpec{Transport: transport, URI: uri},
	}
}

// twoViewRegistries is the fixture reused across the listing assertions.
func twoViewRegistries() []types.Registry {
	return []types.Registry{
		listing(internal, types.TransportHTTP, internalURI),
		listing(rowAcme, types.TransportGit, acmeURI),
	}
}

// TestRenderRegistryList pins the filter, sort, projection, and empty-listing
// rules to their exact rendered bytes.
func TestRenderRegistryList(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// registries is the stored input.
		registries []types.Registry
		// opts are the view options applied.
		opts RegistryListOptions
		// want is the exact expected output.
		want string
	}{
		{
			name:       "default columns sorted by name ascending",
			registries: twoViewRegistries(),
			want: "NAME      TRANSPORT  URI\n" +
				"acme      git        " + acmeURI + "\n" +
				"internal  http       " + internalURI + "\n",
		},
		{
			name:       "search filters by name case-insensitively",
			registries: twoViewRegistries(),
			opts:       RegistryListOptions{Search: "ACME"},
			want: "NAME  TRANSPORT  URI\n" +
				"acme  git        " + acmeURI + "\n",
		},
		{
			name:       "sort by transport descending",
			registries: twoViewRegistries(),
			opts:       RegistryListOptions{Sort: viewLabelTransport, Order: orderDesc},
			want: "NAME      TRANSPORT  URI\n" +
				"internal  http       " + internalURI + "\n" +
				"acme      git        " + acmeURI + "\n",
		},
		{
			name:       "fields selects and forces name first",
			registries: twoViewRegistries(),
			opts:       RegistryListOptions{Fields: []string{viewLabelURI}},
			want: "NAME      URI\n" +
				"acme      " + acmeURI + "\n" +
				"internal  " + internalURI + "\n",
		},
		{
			name:       "absent optional column renders the placeholder",
			registries: []types.Registry{listing(rowAcme, types.TransportHTTP, "https://a/")},
			opts:       RegistryListOptions{Fields: []string{fieldRef}},
			want: "NAME  REF\n" +
				"acme  —\n",
		},
		{
			name:       "empty listing renders nothing",
			registries: nil,
			want:       "",
		},
		{
			name:       "search with no match renders nothing",
			registries: twoViewRegistries(),
			opts:       RegistryListOptions{Search: "absent"},
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := RenderRegistryList(&buf, tt.registries, tt.opts)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestRegistryListOptionsValidate covers the out-of-set field, sort, and order
// values that fail validation.
func TestRegistryListOptionsValidate(t *testing.T) {
	tests := []struct {
		name string
		opts RegistryListOptions
		ok   bool
	}{
		{name: "defaults validate", opts: RegistryListOptions{}, ok: true},
		{name: "known fields validate", opts: RegistryListOptions{Fields: []string{viewLabelURI}}, ok: true},
		{name: "unknown field", opts: RegistryListOptions{Fields: []string{valBogus}}},
		{name: "unknown sort", opts: RegistryListOptions{Sort: viewLabelURI}},
		{name: "unknown order", opts: RegistryListOptions{Order: "sideways"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.ok {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

// fullDetail is a registry populated across every describable field.
func fullDetail() types.Registry {
	return types.Registry{
		Metadata: types.Metadata{
			Name:                 rowAcme,
			CreationTimestamp:    "2026-06-21T07:30:00Z",
			LastUpdatedTimestamp: "2026-06-22T08:00:00Z",
		},
		Spec: types.RegistrySpec{
			Transport: types.TransportGit,
			URI:       acmeURI,
			Ref:       valRefV120,
			Auth:      &types.Auth{Username: valUserRef, Password: valTokenRef},
			Timeout:   valTimeout,
		},
	}
}

// TestRenderRegistryDetail covers the default view, field projection, and the
// nested auth/tls blocks, pinning that credential values stay the stored env
// references.
func TestRenderRegistryDetail(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// opts are the view options applied.
		opts RegistryDetailOptions
		// registry is the input record.
		registry types.Registry
		// wantContains are substrings the descriptor must contain, in order.
		wantContains []string
		// wantAbsent are substrings the descriptor must never contain.
		wantAbsent []string
	}{
		{
			name:     "default shows every populated field",
			registry: fullDetail(),
			wantContains: []string{
				fldName, rowAcme,
				fldTransport, valGit,
				fldURI, acmeURI,
				fldRef, valRefV120,
				fldAuth,
				"username:", valUserRef,
				"password:", valTokenRef,
				fldTimeout, valTimeout,
				"creationTimestamp:", "2026-06-21T07:30:00Z",
				"lastUpdatedTimestamp:", "2026-06-22T08:00:00Z",
			},
		},
		{
			name:         "default omits unpopulated fields",
			registry:     types.Registry{Metadata: types.Metadata{Name: rowAcme}, Spec: types.RegistrySpec{Transport: types.TransportGit, URI: "u"}},
			wantContains: []string{fldName, fldTransport, fldURI},
			wantAbsent:   []string{fldRef, fldAuth, fldTLS, "sshKey:", fldTimeout, "creationTimestamp:"},
		},
		{
			name:         "fields projects and orders, name forced first",
			opts:         RegistryDetailOptions{Fields: []string{viewLabelTransport, viewLabelURI}},
			registry:     fullDetail(),
			wantContains: []string{fldName, fldTransport, fldURI},
			wantAbsent:   []string{fldRef, fldAuth, fldTimeout},
		},
		{
			name:         "auth renders the stored env references, never a secret",
			opts:         RegistryDetailOptions{Fields: []string{viewLabelAuth}},
			registry:     fullDetail(),
			wantContains: []string{fldAuth, valUserRef, valTokenRef},
			wantAbsent:   []string{"s3cr3t"},
		},
		{
			name: "tls and sshKey render their populated sub-fields",
			opts: RegistryDetailOptions{Fields: []string{"tls", "sshKey"}},
			registry: types.Registry{
				Metadata: types.Metadata{Name: rowAcme},
				Spec: types.RegistrySpec{
					Transport: types.TransportHTTP,
					SSHKey:    "/home/dev/.ssh/id_ed25519",
					TLS: &types.TLS{
						SkipVerify: true,
						CACert:     "/etc/ssl/ca.pem",
						ClientCert: "/etc/ssl/client.pem",
						ClientKey:  "/etc/ssl/client.key",
					},
				},
			},
			wantContains: []string{
				fldTLS,
				"skipVerify: true",
				"caCert:", "/etc/ssl/ca.pem",
				"clientCert:", "/etc/ssl/client.pem",
				"clientKey:", "/etc/ssl/client.key",
				"sshKey:", "/home/dev/.ssh/id_ed25519",
			},
		},
		{
			name: "an empty tls block is omitted",
			opts: RegistryDetailOptions{Fields: []string{"tls"}},
			registry: types.Registry{
				Metadata: types.Metadata{Name: rowAcme},
				Spec:     types.RegistrySpec{Transport: types.TransportHTTP, TLS: &types.TLS{}},
			},
			wantContains: []string{fldName},
			wantAbsent:   []string{fldTLS},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := RenderRegistryDetail(&buf, tt.registry, tt.opts)

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

// TestRegistryDetailOptionsValidate covers the unknown-field rejection.
func TestRegistryDetailOptionsValidate(t *testing.T) {
	assert.NoError(t, RegistryDetailOptions{}.Validate())
	assert.NoError(t, RegistryDetailOptions{Fields: []string{viewLabelTransport}}.Validate())
	assert.Error(t, RegistryDetailOptions{Fields: []string{valBogus}}.Validate())
}
