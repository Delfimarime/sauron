package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

const (
	registryAcme = "acme"
	timestamp    = "2026-06-15T10:00:00Z"
)

// TestRoundTrip asserts every document type survives a YAML marshal/unmarshal
// cycle unchanged, using representative documents. It is purely in-memory — no filesystem.
func TestRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		doc  any
		into any
	}{
		{
			name: "git registry with auth",
			doc: &types.Registry{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindRegistry},
				Metadata: types.Metadata{Name: registryAcme},
				Spec: types.RegistrySpec{
					Transport: types.TransportGit,
					Source:       "git@github.com:acme/artifacts.git",
					Credentials:      &types.Credentials{Username: "${env:ACME_USER}", Password: "${env:ACME_TOKEN}"},
					SSHKey:    "/path/id_ed25519",
					Timeout:   "30s",
				},
			},
			into: &types.Registry{},
		},
		{
			name: "http registry with tls and labels",
			doc: &types.Registry{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindRegistry},
				Metadata: types.Metadata{Name: "mirror", Labels: map[string]string{"team": "backend"}},
				Spec: types.RegistrySpec{
					Transport: types.TransportHTTP,
					Source:       "https://example.com/artifacts",
					TLS: &types.TLS{
						SkipVerify: true,
						CACert:     "/path/ca.pem",
						ClientCert: "/path/client.pem",
						ClientKey:  "/path/client.key",
					},
				},
			},
			into: &types.Registry{},
		},
		{
			name: "http registry",
			doc: &types.Registry{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindRegistry},
				Metadata: types.Metadata{Name: "local"},
				Spec:     types.RegistrySpec{Transport: types.TransportHTTP, Source: "https://acme.example/artifacts"},
			},
			into: &types.Registry{},
		},
		{
			name: "provider",
			doc: &types.Provider{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindProvider},
				Metadata: types.Metadata{Name: types.ProviderClaude},
			},
			into: &types.Provider{},
		},
		{
			name: "skill artifact",
			doc: &types.Artifact{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindSkill},
				Metadata: types.Metadata{Name: "go-style", Labels: map[string]string{"team": "backend"}},
				Spec: types.ArtifactSpec{
					Version:     "v1.4.0",
					Path:        "sauron-go-style",
					InstalledAt: timestamp,
					UpdatedAt:   timestamp,
				},
			},
			into: &types.Artifact{},
		},
		{
			name: "agent artifact",
			doc: &types.Artifact{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindAgent},
				Metadata: types.Metadata{Name: "code-reviewer"},
				Spec: types.ArtifactSpec{
					Version:     "v2.0.0",
					Path:        "sauron-code-reviewer",
					InstalledAt: timestamp,
					UpdatedAt:   timestamp,
				},
			},
			into: &types.Artifact{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := yaml.Marshal(tt.doc)
			require.NoError(t, err)

			require.NoError(t, yaml.Unmarshal(data, tt.into))
			assert.Equal(t, tt.doc, tt.into)
		})
	}
}

// TestRegistrySpecRef asserts that RegistrySpec.Revision round-trips through both
// YAML and JSON, is omitted from the encoded form when empty, and is preserved
// when set. It is purely in-memory — no filesystem.
func TestRegistrySpecRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ref         string
		wantPresent bool
	}{
		{name: "ref omitted when empty", ref: "", wantPresent: false},
		{name: "ref preserved when set", ref: "v1.4.0", wantPresent: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc := &types.Registry{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindRegistry},
				Metadata: types.Metadata{Name: registryAcme},
				Spec: types.RegistrySpec{
					Transport: types.TransportGit,
					Source:       "git@github.com:acme/artifacts.git",
					Revision:       tt.ref,
				},
			}

			t.Run("yaml", func(t *testing.T) {
				t.Parallel()

				data, err := yaml.Marshal(doc)
				require.NoError(t, err)

				var raw map[string]any
				require.NoError(t, yaml.Unmarshal(data, &raw))
				spec, ok := raw["spec"].(map[string]any)
				require.True(t, ok)
				_, present := spec["revision"]
				assert.Equal(t, tt.wantPresent, present)

				into := &types.Registry{}
				require.NoError(t, yaml.Unmarshal(data, into))
				assert.Equal(t, tt.ref, into.Spec.Revision)
			})

			t.Run("json", func(t *testing.T) {
				t.Parallel()

				data, err := json.Marshal(doc)
				require.NoError(t, err)

				var raw map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(data, &raw))
				var spec map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(raw["spec"], &spec))
				_, present := spec["revision"]
				assert.Equal(t, tt.wantPresent, present)

				into := &types.Registry{}
				require.NoError(t, json.Unmarshal(data, into))
				assert.Equal(t, tt.ref, into.Spec.Revision)
			})
		})
	}
}

// TestProviderSpecSyncTimestamps asserts that ProviderSpec's sync timestamps
// round-trip through YAML, are omitted from the encoded form when the spec is
// empty (a provider that has never synced), and are preserved when set. It is
// purely in-memory — no filesystem.
func TestProviderSpecSyncTimestamps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		spec        types.ProviderSpec
		wantPresent bool
	}{
		{name: "spec omitted when never synced", spec: types.ProviderSpec{}, wantPresent: false},
		{
			name: "spec preserved when synced",
			spec: types.ProviderSpec{
				LastSyncedAt:      "2026-06-25T09:15:00Z",
				LastSyncAttemptAt: "2026-06-26T06:00:00Z",
			},
			wantPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc := &types.Provider{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindProvider},
				Metadata: types.Metadata{Name: types.ProviderClaude},
				Spec:     tt.spec,
			}

			data, err := yaml.Marshal(doc)
			require.NoError(t, err)

			var raw map[string]any
			require.NoError(t, yaml.Unmarshal(data, &raw))
			_, present := raw["spec"]
			assert.Equal(t, tt.wantPresent, present)

			into := &types.Provider{}
			require.NoError(t, yaml.Unmarshal(data, into))
			assert.Equal(t, tt.spec, into.Spec)
		})
	}
}

// TestEnvelopeKeysAtTopLevel pins the load-bearing detail that the embedded
// TypeMeta serializes apiVersion/kind at the document's top level (yaml.v3
// requires the ",inline" tag for this), alongside metadata and spec.
func TestEnvelopeKeysAtTopLevel(t *testing.T) {
	t.Parallel()

	reg := &types.Registry{
		TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindRegistry},
		Metadata: types.Metadata{Name: registryAcme},
		Spec:     types.RegistrySpec{Transport: types.TransportGit, Source: "git@github.com:acme/a.git"},
	}

	data, err := yaml.Marshal(reg)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, yaml.Unmarshal(data, &raw))

	assert.Equal(t, types.APIVersion, raw["apiVersion"])
	assert.Equal(t, types.KindRegistry, raw["kind"])
	assert.Contains(t, raw, "metadata")
	assert.Contains(t, raw, "spec")
}
