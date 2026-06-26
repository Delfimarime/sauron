//go:build unit

package gherkin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// providerWithRegistry is the settings.yaml stream after a provider is set alongside
// a registry: decodeProviders must keep the Provider and skip the Registry.
const providerWithRegistry = `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  createdAt: "2026-06-21T07:30:00Z"
  lastUpdatedAt: "2026-06-21T07:30:00Z"
spec:
  transport: filesystem
  source: /opt/registry/default
---
apiVersion: sauron.raitonbl.com/v1
kind: Provider
metadata:
  name: claude
`

// twoProviders is a malformed-for-v1 stream used to prove decodeProviders reads
// every Provider document and oneProvider rejects more than one.
const twoProviders = `apiVersion: sauron.raitonbl.com/v1
kind: Provider
metadata:
  name: claude
---
apiVersion: sauron.raitonbl.com/v1
kind: Provider
metadata:
  name: zencoder
`

func TestDecodeProvidersSkipsRegistryAndEmptyDocs(t *testing.T) {
	providers, err := decodeProviders([]byte(providerWithRegistry))
	require.NoError(t, err)
	require.Len(t, providers, 1, "the Registry document and empty documents are skipped")
	assert.Equal(t, types.ProviderClaude, providers[0].Metadata.Name)
}

func TestOneProvider(t *testing.T) {
	one, err := decodeProviders([]byte(providerWithRegistry))
	require.NoError(t, err)
	provider, err := oneProvider(one)
	require.NoError(t, err)
	assert.Equal(t, types.ProviderClaude, provider.Metadata.Name)

	_, err = oneProvider(nil)
	assert.Error(t, err, "no provider configured is an error")

	two, err := decodeProviders([]byte(twoProviders))
	require.NoError(t, err)
	_, err = oneProvider(two)
	assert.Error(t, err, "more than one provider is an error")
}

// TestSetProviderControllerAssertsThroughRuntime drives the Then steps end to end
// against an in-memory settings.yaml served by the fake runtime (no real fs), and
// the output assertions against a recorded command result.
func TestSetProviderControllerAssertsThroughRuntime(t *testing.T) {
	ctx := context.Background()
	rt := &fakeRuntime{files: map[string][]byte{
		settingsFile: []byte(providerWithRegistry),
	}}
	commands := &commandController{rt: rt, last: &commandResult{code: 0, output: "provider set to \"claude\"\n"}}
	c := &setProviderController{rt: rt, commands: commands}

	require.NoError(t, c.providerIsSetTo(ctx, types.ProviderClaude))
	require.NoError(t, c.reportsProviderSet(ctx, types.ProviderClaude))

	assert.Error(t, c.providerIsSetTo(ctx, types.ProviderZencoder), "the recorded provider is claude")
	assert.Error(t, c.reportsProviderSet(ctx, types.ProviderZencoder), "the output names claude")
	assert.Error(t, c.reportsProviderAlreadySet(ctx, types.ProviderClaude), "the output is the change confirmation, not the no-change notice")
}
