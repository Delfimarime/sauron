//go:build unit

package gherkin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

const twoRegistries = `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
  labels:
    team: platform
spec:
  transport: http
  uri: http://registry-http-default
  auth:
    username: ${env:ACME_USER}
    password: ${env:ACME_TOKEN}
---
apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: local
spec:
  transport: filesystem
  uri: /opt/registry/default
`

func TestDecodeRegistriesMultiDoc(t *testing.T) {
	regs, err := decodeRegistries([]byte(twoRegistries))
	require.NoError(t, err)
	require.Len(t, regs, 2)
	assert.Equal(t, "acme", regs[0].Metadata.Name)
	assert.Equal(t, types.TransportHTTP, regs[0].Spec.Transport)
	assert.Equal(t, "${env:ACME_TOKEN}", regs[0].Spec.Auth.Password)
	assert.Equal(t, types.TransportFilesystem, regs[1].Spec.Transport)
}

func TestDecodeRegistriesSkipsOtherKindsAndEmptyDocs(t *testing.T) {
	stream := "---\n" + twoRegistries + "\n---\napiVersion: sauron.raitonbl.com/v1\nkind: Provider\nmetadata:\n  name: claude\n"
	regs, err := decodeRegistries([]byte(stream))
	require.NoError(t, err)
	assert.Len(t, regs, 2, "empty documents and non-Registry kinds are skipped")
}

func TestFindRegistry(t *testing.T) {
	regs, err := decodeRegistries([]byte(twoRegistries))
	require.NoError(t, err)

	reg, err := findRegistry(regs, "local")
	require.NoError(t, err)
	assert.Equal(t, types.TransportFilesystem, reg.Spec.Transport)

	_, err = findRegistry(regs, "absent")
	assert.Error(t, err)
}

func TestRegistryField(t *testing.T) {
	regs, err := decodeRegistries([]byte(twoRegistries))
	require.NoError(t, err)
	reg := regs[0]

	cases := map[string]string{
		"kind":           "Registry",
		"apiVersion":     "sauron.raitonbl.com/v1",
		"metadata.name":  "acme",
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

// TestConfigurationControllerAssertsThroughRuntime drives a couple of Then steps end
// to end against an in-memory file served by the fake runtime (no real fs).
func TestConfigurationControllerAssertsThroughRuntime(t *testing.T) {
	ctx := context.Background()
	rt := &fakeRuntime{files: map[string][]byte{registriesFile: []byte(twoRegistries)}}
	c := &configurationController{rt: rt}

	require.NoError(t, c.registryCount(ctx, 2))
	require.NoError(t, c.registryExists(ctx, "acme"))
	require.NoError(t, c.registryTransport(ctx, "acme", "http"))
	require.NoError(t, c.registryLabel(ctx, "acme", "team", "platform"))
	require.NoError(t, c.registryPasswordRef(ctx, "acme", "${env:ACME_TOKEN}"))
	require.NoError(t, c.configDoesNotContain(ctx, "s3cr3t"))

	assert.Error(t, c.registryCount(ctx, 99))
	assert.Error(t, c.registryTransport(ctx, "acme", "git"))
	assert.Error(t, c.configDoesNotContain(ctx, "ACME_TOKEN"), "the reference text IS present")
}
