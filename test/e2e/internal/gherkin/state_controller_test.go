//go:build unit

package gherkin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// oneRegistryDoc is the single Registry document settings.yaml holds, alongside the
// Provider that shares the file.
const oneRegistryDoc = `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  creationTimestamp: "2026-06-21T07:30:00Z"
  lastUpdatedTimestamp: "2026-06-21T07:30:00Z"
spec:
  transport: http
  uri: http://registry-http-default
  auth:
    username: ${env:ACME_USER}
    password: ${env:ACME_TOKEN}
---
apiVersion: sauron.raitonbl.com/v1
kind: Provider
metadata:
  name: claude
`

// twoRegistries is a malformed-for-v1 stream (two Registry documents) used to prove
// the decoder reads every Registry document and oneRegistry rejects more than one.
const twoRegistries = `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
spec:
  transport: http
  uri: http://registry-http-default
---
apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: local
spec:
  transport: filesystem
  uri: /opt/registry/default
`

const oneSkillTrack = `apiVersion: sauron.raitonbl.com/v1
kind: Skill
metadata:
  name: sauron-acme-go-style
spec:
  digest: sha256:abc
  path: skills/sauron-acme-go-style
  installedAt: "2026-06-21T07:30:00Z"
  updatedAt: "2026-06-21T07:30:00Z"
`

func TestDecodeRegistriesSkipsProviderAndEmptyDocs(t *testing.T) {
	regs, err := decodeRegistries([]byte(oneRegistryDoc))
	require.NoError(t, err)
	require.Len(t, regs, 1, "the Provider document and empty documents are skipped")
	assert.Equal(t, types.TransportHTTP, regs[0].Spec.Transport)
	assert.Equal(t, "${env:ACME_TOKEN}", regs[0].Spec.Auth.Password)
}

func TestOneRegistry(t *testing.T) {
	one, err := decodeRegistries([]byte(oneRegistryDoc))
	require.NoError(t, err)
	reg, err := oneRegistry(one)
	require.NoError(t, err)
	assert.Equal(t, types.TransportHTTP, reg.Spec.Transport)

	_, err = oneRegistry(nil)
	assert.Error(t, err, "no registry configured is an error")

	two, err := decodeRegistries([]byte(twoRegistries))
	require.NoError(t, err)
	_, err = oneRegistry(two)
	assert.Error(t, err, "more than one registry is an error")
}

func TestRegistryField(t *testing.T) {
	regs, err := decodeRegistries([]byte(oneRegistryDoc))
	require.NoError(t, err)
	reg := regs[0]

	cases := map[string]string{
		"kind":           "Registry",
		"apiVersion":     "sauron.raitonbl.com/v1",
		"spec.transport": "http",
		"spec.uri":       "http://registry-http-default",
	}
	for field, want := range cases {
		got, err := registryField(reg, field)
		require.NoError(t, err, field)
		assert.Equal(t, want, got, field)
	}

	_, err = registryField(reg, "spec.unknown")
	assert.Error(t, err)
}

func TestDecodeSkillsKeepsSkillDocuments(t *testing.T) {
	skills, err := decodeSkills([]byte(oneSkillTrack))
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "sauron-acme-go-style", skills[0].Metadata.Name)
}

// TestStateControllerAssertsThroughRuntime drives a couple of Then steps end to end
// against in-memory files served by the fake runtime (no real fs).
func TestStateControllerAssertsThroughRuntime(t *testing.T) {
	ctx := context.Background()
	rt := &fakeRuntime{files: map[string][]byte{
		settingsFile: []byte(oneRegistryDoc),
		trackFile:    []byte(oneSkillTrack),
	}}
	c := &stateController{rt: rt}

	require.NoError(t, c.exactlyOneRegistry(ctx))
	require.NoError(t, c.registryTransport(ctx, "http"))
	require.NoError(t, c.registryPasswordRef(ctx, "${env:ACME_TOKEN}"))
	require.NoError(t, c.registryHasCreationTimestamp(ctx))
	require.NoError(t, c.configDoesNotContain(ctx, "s3cr3t"))
	require.NoError(t, c.skillStillTracked(ctx, "sauron-acme-go-style"))

	assert.Error(t, c.noRegistry(ctx), "there IS one registry")
	assert.Error(t, c.registryTransport(ctx, "git"))
	assert.Error(t, c.configDoesNotContain(ctx, "ACME_TOKEN"), "the reference text IS present")
	assert.Error(t, c.skillStillTracked(ctx, "absent"))
}
