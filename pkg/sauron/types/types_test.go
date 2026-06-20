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
	personaName  = "backend-dev"
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
					URI:       "git@github.com:acme/artifacts.git",
					Auth:      &types.Auth{Username: "${env:ACME_USER}", Password: "${env:ACME_TOKEN}"},
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
					URI:       "https://example.com/artifacts",
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
			name: "filesystem registry",
			doc: &types.Registry{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindRegistry},
				Metadata: types.Metadata{Name: "local"},
				Spec:     types.RegistrySpec{Transport: types.TransportFilesystem, URI: "/srv/artifacts"},
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
			name: "schedule",
			doc: &types.Schedule{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindSchedule},
				Metadata: types.Metadata{Name: types.ScheduleSync},
				Spec:     types.ScheduleSpec{Cron: "0 */6 * * *"},
			},
			into: &types.Schedule{},
		},
		{
			name: "skill",
			doc: &types.Skill{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindSkill},
				Metadata: types.Metadata{Name: "go-style", Labels: map[string]string{"team": "backend"}},
				Spec: types.ArtifactSpec{
					Registry:    registryAcme,
					Version:     "v1.4.0",
					Digest:      "sha256:abc",
					Path:        "skills/sauron-acme-go-style",
					Provenance:  types.Provenance{Direct: true, Personas: []string{personaName}},
					InstalledAt: timestamp,
					UpdatedAt:   timestamp,
				},
			},
			into: &types.Skill{},
		},
		{
			name: "agent",
			doc: &types.Agent{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindAgent},
				Metadata: types.Metadata{Name: "code-reviewer"},
				Spec: types.ArtifactSpec{
					Registry:    registryAcme,
					Digest:      "sha256:def",
					Path:        "agents/sauron-acme-code-reviewer",
					Provenance:  types.Provenance{Direct: false, Personas: []string{personaName}},
					InstalledAt: timestamp,
					UpdatedAt:   timestamp,
				},
			},
			into: &types.Agent{},
		},
		{
			name: "persona",
			doc: &types.Persona{
				TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindPersona},
				Metadata: types.Metadata{Name: personaName},
				Spec: types.PersonaSpec{
					Registry:    registryAcme,
					Version:     "9f4d2a1",
					Digest:      "sha256:ghi",
					Members:     types.PersonaMembers{Skills: []string{"go-style", "sql-review"}, Agents: []string{"code-reviewer"}},
					InstalledAt: timestamp,
					UpdatedAt:   timestamp,
				},
			},
			into: &types.Persona{},
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

// TestRegistrySpecRef asserts that RegistrySpec.Ref round-trips through both
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
					URI:       "git@github.com:acme/artifacts.git",
					Ref:       tt.ref,
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
				_, present := spec["ref"]
				assert.Equal(t, tt.wantPresent, present)

				into := &types.Registry{}
				require.NoError(t, yaml.Unmarshal(data, into))
				assert.Equal(t, tt.ref, into.Spec.Ref)
			})

			t.Run("json", func(t *testing.T) {
				t.Parallel()

				data, err := json.Marshal(doc)
				require.NoError(t, err)

				var raw map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(data, &raw))
				var spec map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(raw["spec"], &spec))
				_, present := spec["ref"]
				assert.Equal(t, tt.wantPresent, present)

				into := &types.Registry{}
				require.NoError(t, json.Unmarshal(data, into))
				assert.Equal(t, tt.ref, into.Spec.Ref)
			})
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
		Spec:     types.RegistrySpec{Transport: types.TransportGit, URI: "git@github.com:acme/a.git"},
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
