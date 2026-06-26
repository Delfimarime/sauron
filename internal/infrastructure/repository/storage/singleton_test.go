package storage

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// providerYAML is a foreign-kind document that shares settings.yaml with the
// Registry; the singleton operations must leave it untouched.
const providerYAML = `apiVersion: sauron.raitonbl.com/v1
kind: Provider
metadata:
  name: claude
`

// TestStoreFirstAbsent returns nil when the file or a matching document is missing.
func TestStoreFirstAbsent(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// seed is written before the lookup (empty: no file).
		seed string
	}{
		{name: caseMissingFile, seed: ""},
		{name: "only a foreign kind", seed: providerYAML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			store, fs := newTestStore(t)
			if tt.seed != "" {
				require.NoError(t, afero.WriteFile(fs, registriesFile, []byte(tt.seed), filePerm))
			}

			// Act.
			node, err := store.First(context.Background(), types.KindRegistry)

			// Assert.
			require.NoError(t, err)
			assert.Nil(t, node)
		})
	}
}

// TestStoreFirstReturnsMatch returns the single valid registry, skipping foreign kinds.
func TestStoreFirstReturnsMatch(t *testing.T) {
	// Arrange: a provider precedes the registry in the shared file.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile,
		[]byte(documentSeparator+providerYAML+documentSeparator+validRegistryYAML), filePerm))

	// Act.
	node, err := store.First(context.Background(), types.KindRegistry)

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, types.KindRegistry, kindOf(node))
}

// TestStoreFirstValidatesOnRead rejects a malformed registry on read.
func TestStoreFirstValidatesOnRead(t *testing.T) {
	// Arrange: an unknown spec field violates additionalProperties: false.
	malformed := `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
spec:
  transport: git
  source: https://example.com/acme.git
  bogus: nope
`
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile, []byte(malformed), filePerm))

	// Act.
	node, err := store.First(context.Background(), types.KindRegistry)

	// Assert.
	require.Error(t, err)
	assert.Nil(t, node)
}

// TestStoreFirstUnknownKind reports an error for a kind with no backing file.
func TestStoreFirstUnknownKind(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)

	// Act.
	_, err := store.First(context.Background(), "Nonexistent")

	// Assert.
	require.ErrorIs(t, err, errUnknownKind)
}

// TestStoreReplaceRoundTrip writes the single registry and reads it back valid.
func TestStoreReplaceRoundTrip(t *testing.T) {
	// Arrange.
	store, fs := newTestStore(t)

	// Act.
	require.NoError(t, store.Replace(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML)))

	// Assert: it reads back and no temp artifact remains.
	got, err := store.First(context.Background(), types.KindRegistry)
	require.NoError(t, err)
	require.NotNil(t, got)

	exists, err := afero.Exists(fs, registriesFile+".tmp")
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestStoreReplaceKeepsOneRegistry replaces the registry in place, never
// accumulating a second document.
func TestStoreReplaceKeepsOneRegistry(t *testing.T) {
	// Arrange: an http registry already configured.
	store, _ := newTestStore(t)
	require.NoError(t, store.Replace(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML)))

	// Act: replace with a different registry.
	require.NoError(t, store.Replace(context.Background(), types.KindRegistry, nodeFromYAML(t, betaRegistryYAML)))

	// Assert: exactly the replacement is present.
	got, err := store.First(context.Background(), types.KindRegistry)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "beta", nameOf(got))
}

// TestStoreReplacePreservesForeignKinds keeps a Provider document when the
// Registry is replaced in the shared file.
func TestStoreReplacePreservesForeignKinds(t *testing.T) {
	// Arrange: a provider already in settings.yaml.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile, []byte(documentSeparator+providerYAML), filePerm))

	// Act.
	require.NoError(t, store.Replace(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML)))

	// Assert: both the provider and the new registry are present.
	raw, err := afero.ReadFile(fs, registriesFile)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(raw), "kind: Provider"))
	assert.True(t, strings.Contains(string(raw), "kind: Registry"))
}

// TestStoreReplaceUnknownKind reports an error for a kind with no backing file.
func TestStoreReplaceUnknownKind(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)

	// Act.
	err := store.Replace(context.Background(), "Nonexistent", nodeFromYAML(t, validRegistryYAML))

	// Assert.
	require.ErrorIs(t, err, errUnknownKind)
}

// TestStoreReplaceLockContended reports an error when the lock is already held.
func TestStoreReplaceLockContended(t *testing.T) {
	// Arrange: a pre-existing lock file blocks acquisition.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, lockFile, nil, lockPerm))

	// Act.
	err := store.Replace(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML))

	// Assert.
	require.Error(t, err)
}

// TestStorePurgeDropsRegistry removes the registry, leaving no match.
func TestStorePurgeDropsRegistry(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)
	require.NoError(t, store.Replace(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML)))

	// Act.
	require.NoError(t, store.Purge(context.Background(), types.KindRegistry))

	// Assert.
	got, err := store.First(context.Background(), types.KindRegistry)
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestStorePurgePreservesForeignKinds keeps a Provider document when the
// Registry is purged from the shared file.
func TestStorePurgePreservesForeignKinds(t *testing.T) {
	// Arrange: a provider and a registry in settings.yaml.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile,
		[]byte(documentSeparator+providerYAML+documentSeparator+validRegistryYAML), filePerm))

	// Act.
	require.NoError(t, store.Purge(context.Background(), types.KindRegistry))

	// Assert: the provider survives, the registry is gone.
	raw, err := afero.ReadFile(fs, registriesFile)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(raw), "kind: Provider"))
	assert.False(t, strings.Contains(string(raw), "kind: Registry"))
}

// TestStorePurgeAbsentIsNoOp returns success and leaves the file untouched when
// no registry is present.
func TestStorePurgeAbsentIsNoOp(t *testing.T) {
	// Arrange: only a foreign kind is present.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile, []byte(documentSeparator+providerYAML), filePerm))
	before, _ := afero.ReadFile(fs, registriesFile)

	// Act.
	err := store.Purge(context.Background(), types.KindRegistry)

	// Assert: success and the file is byte-for-byte unchanged.
	require.NoError(t, err)
	after, _ := afero.ReadFile(fs, registriesFile)
	assert.Equal(t, before, after)
}

// TestStorePurgeUnknownKind reports an error for a kind with no backing file.
func TestStorePurgeUnknownKind(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)

	// Act.
	err := store.Purge(context.Background(), "Nonexistent")

	// Assert.
	require.ErrorIs(t, err, errUnknownKind)
}
