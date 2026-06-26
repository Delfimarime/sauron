package storage

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/internal/config"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// registriesFile is the backing file for the Registry kind.
const registriesFile = "settings.yaml"

// caseMissingFile names the shared "no backing file" table case.
const caseMissingFile = "missing file"

// acmeName is a fixture registry name shared by the storage tests.
const acmeName = "acme"

// acmeURI is the fixture registry URI shared by the storage tests.
const acmeURI = "https://example.com/acme.git"

// validRegistryYAML is a schema-valid Registry document.
const validRegistryYAML = `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
spec:
  transport: git
  source: https://example.com/acme.git
  revision: main
`

// newTestStore builds a Store over an in-memory filesystem.
func newTestStore(t *testing.T) (*Store, afero.Fs) {
	t.Helper()
	fs := afero.NewMemMapFs()
	store, err := NewStore(fs)
	require.NoError(t, err)
	return store, fs
}

// nodeFromYAML parses a single YAML document into a node.
func nodeFromYAML(t *testing.T, raw string) *yaml.Node {
	t.Helper()
	var node yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(raw), &node))
	return &node
}

// TestStoreFindOneAbsent returns nil when the file or the document is missing.
func TestStoreFindOneAbsent(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// seed is written to the registries file before the lookup (empty: no file).
		seed string
		// lookup is the document name searched for.
		lookup string
	}{
		{name: caseMissingFile, seed: "", lookup: acmeName},
		{name: "missing document", seed: validRegistryYAML, lookup: "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			store, fs := newTestStore(t)
			if tt.seed != "" {
				require.NoError(t, afero.WriteFile(fs, registriesFile, []byte(tt.seed), filePerm))
			}

			// Act.
			node, err := store.FindOne(context.Background(), types.KindRegistry, tt.lookup)

			// Assert.
			require.NoError(t, err)
			assert.Nil(t, node)
		})
	}
}

// TestStoreFindOneValidatesOnRead rejects a malformed document on read.
func TestStoreFindOneValidatesOnRead(t *testing.T) {
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
	node, err := store.FindOne(context.Background(), types.KindRegistry, acmeName)

	// Assert.
	require.Error(t, err)
	assert.Nil(t, node)
}

// TestStoreFindOneReturnsMatch returns the matching, valid document.
func TestStoreFindOneReturnsMatch(t *testing.T) {
	// Arrange.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile, []byte(validRegistryYAML), filePerm))

	// Act.
	node, err := store.FindOne(context.Background(), types.KindRegistry, acmeName)

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, acmeName, nameOf(node))
}

// TestStoreUnknownKind reports an error for a kind with no backing file.
func TestStoreUnknownKind(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)

	// Act.
	_, err := store.FindOne(context.Background(), "Nonexistent", "x")

	// Assert.
	require.ErrorIs(t, err, errUnknownKind)
}

// TestStoreAppendRoundTrip appends a document and reads it back.
func TestStoreAppendRoundTrip(t *testing.T) {
	// Arrange.
	store, fs := newTestStore(t)
	first := nodeFromYAML(t, validRegistryYAML)
	second := nodeFromYAML(t, `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: beta
spec:
  transport: http
  source: https://example.com/beta
`)

	// Act.
	require.NoError(t, store.Append(context.Background(), types.KindRegistry, first))
	require.NoError(t, store.Append(context.Background(), types.KindRegistry, second))

	// Assert: both documents are retrievable, and no temp artifact remains.
	got, err := store.FindOne(context.Background(), types.KindRegistry, acmeName)
	require.NoError(t, err)
	require.NotNil(t, got)

	got, err = store.FindOne(context.Background(), types.KindRegistry, "beta")
	require.NoError(t, err)
	require.NotNil(t, got)

	exists, err := afero.Exists(fs, registriesFile+".tmp")
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestStoreAppendSerializes runs concurrent appends and asserts every document
// survives — the lock prevents lost updates.
func TestStoreAppendSerializes(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)
	const writers = 8

	// Act: append distinct documents concurrently.
	var wg sync.WaitGroup
	wg.Add(writers)
	for i := range writers {
		go func(i int) {
			defer wg.Done()
			doc := nodeFromYAML(t, registryYAML(i))
			assert.NoError(t, store.Append(context.Background(), types.KindRegistry, doc))
		}(i)
	}
	wg.Wait()

	// Assert: each writer's document is present.
	for i := range writers {
		got, err := store.FindOne(context.Background(), types.KindRegistry, registryName(i))
		require.NoError(t, err)
		require.NotNil(t, got, "document %d lost", i)
	}
}

// TestStoreFindOneMalformedStream surfaces a YAML parse error from the file.
func TestStoreFindOneMalformedStream(t *testing.T) {
	// Arrange: invalid YAML in the registries file.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile, []byte("\tnot: [valid"), filePerm))

	// Act.
	_, err := store.FindOne(context.Background(), types.KindRegistry, acmeName)

	// Assert.
	require.Error(t, err)
}

// TestStoreAppendReadOnlyFails reports an error when the commit cannot be written.
func TestStoreAppendReadOnlyFails(t *testing.T) {
	// Arrange: a read-only filesystem rejects the temp write.
	store, err := NewStore(afero.NewReadOnlyFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	// Act.
	err = store.Append(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML))

	// Assert.
	require.Error(t, err)
}

// TestStoreLockContended reports an error when the lock is already held on disk.
func TestStoreLockContended(t *testing.T) {
	// Arrange: a pre-existing lock file blocks acquisition.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, lockFile, nil, lockPerm))

	// Act.
	err := store.Append(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML))

	// Assert.
	require.Error(t, err)
}

// TestStoreRemoveDropsMatch removes the matching document and leaves the rest,
// keeping every other document retrievable.
func TestStoreRemoveDropsMatch(t *testing.T) {
	// Arrange: two documents in the stream.
	store, _ := newTestStore(t)
	require.NoError(t, store.Append(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML)))
	require.NoError(t, store.Append(context.Background(), types.KindRegistry, nodeFromYAML(t, `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: beta
spec:
  transport: http
  source: https://example.com/beta
`)))

	// Act.
	require.NoError(t, store.Remove(context.Background(), types.KindRegistry, acmeName))

	// Assert: acme is gone, beta survives and still reads back valid.
	gone, err := store.FindOne(context.Background(), types.KindRegistry, acmeName)
	require.NoError(t, err)
	assert.Nil(t, gone)

	beta, err := store.FindOne(context.Background(), types.KindRegistry, "beta")
	require.NoError(t, err)
	require.NotNil(t, beta)
}

// TestStoreRemoveLastEmptiesFile removes the only document, leaving an empty stream.
func TestStoreRemoveLastEmptiesFile(t *testing.T) {
	// Arrange.
	store, fs := newTestStore(t)
	require.NoError(t, store.Append(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML)))

	// Act.
	require.NoError(t, store.Remove(context.Background(), types.KindRegistry, acmeName))

	// Assert: the document is gone and no temp artifact remains.
	gone, err := store.FindOne(context.Background(), types.KindRegistry, acmeName)
	require.NoError(t, err)
	assert.Nil(t, gone)

	exists, err := afero.Exists(fs, registriesFile+".tmp")
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestStoreRemoveAbsentIsNoOp returns success and leaves the file untouched when no
// document matches (FR-005 idempotency) — including a missing file.
func TestStoreRemoveAbsentIsNoOp(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// seed is written before the removal (empty: no file).
		seed string
		// remove is the document name removed.
		remove string
	}{
		{name: caseMissingFile, seed: "", remove: acmeName},
		{name: "no matching document", seed: validRegistryYAML, remove: "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			store, fs := newTestStore(t)
			if tt.seed != "" {
				require.NoError(t, afero.WriteFile(fs, registriesFile, []byte(tt.seed), filePerm))
			}
			before, _ := afero.ReadFile(fs, registriesFile)

			// Act.
			err := store.Remove(context.Background(), types.KindRegistry, tt.remove)

			// Assert: success and the file is byte-for-byte unchanged.
			require.NoError(t, err)
			after, _ := afero.ReadFile(fs, registriesFile)
			assert.Equal(t, before, after)
		})
	}
}

// TestStoreRemoveUnknownKind reports an error for a kind with no backing file.
func TestStoreRemoveUnknownKind(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)

	// Act.
	err := store.Remove(context.Background(), "Nonexistent", "x")

	// Assert.
	require.ErrorIs(t, err, errUnknownKind)
}

// TestStoreRemoveLockContended reports an error when the lock is already held.
func TestStoreRemoveLockContended(t *testing.T) {
	// Arrange: a pre-existing lock file blocks acquisition.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile, []byte(validRegistryYAML), filePerm))
	require.NoError(t, afero.WriteFile(fs, lockFile, nil, lockPerm))

	// Act.
	err := store.Remove(context.Background(), types.KindRegistry, acmeName)

	// Assert.
	require.Error(t, err)
}

// TestStoreRemoveMalformedStream surfaces a YAML parse error from the file.
func TestStoreRemoveMalformedStream(t *testing.T) {
	// Arrange.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile, []byte("\tnot: [valid"), filePerm))

	// Act.
	err := store.Remove(context.Background(), types.KindRegistry, acmeName)

	// Assert.
	require.Error(t, err)
}

// registryName names the i-th concurrent test registry.
func registryName(i int) string {
	return "reg" + string(rune('a'+i))
}

// registryYAML renders the i-th concurrent test registry document.
func registryYAML(i int) string {
	return "apiVersion: sauron.raitonbl.com/v1\n" +
		"kind: Registry\n" +
		"metadata:\n" +
		"  name: " + registryName(i) + "\n" +
		"spec:\n" +
		"  transport: git\n" +
		"  source: https://example.com/" + registryName(i) + ".git\n"
}

// TestNewStore asserts the store retains the injected filesystem.
func TestNewStore(t *testing.T) {
	// Arrange.
	fs := afero.NewMemMapFs()

	// Act.
	store, err := NewStore(fs)

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Same(t, fs, store.fs)
}

// TestNewFxOptions resolves a Store and RegistriesStore through the container to
// exercise the wiring.
func TestNewFxOptions(t *testing.T) {
	// Arrange + Act.
	var store *Store
	var registries RegistriesStore
	app := fx.New(
		fx.Supply(config.Configuration{HomeDirectory: t.TempDir()}),
		NewFxOptions(),
		fx.Populate(&store, &registries),
	)

	// Assert.
	require.NoError(t, app.Err())
	require.NotNil(t, store.fs)
	require.NotNil(t, registries)
}

// TestNewFilesystem asserts paths resolve under the configured home, without touching the real filesystem.
func TestNewFilesystem(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// probe is the path resolved through the returned fs.
		probe string
	}{
		// A flat path resolves directly under home.
		{name: "roots a file under home", probe: "settings.yaml"},
		// A nested path resolves under home too.
		{name: "roots a nested path under home", probe: "sub/track.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: t.TempDir is used only as a home path string; nothing is written.
			home := t.TempDir()
			fs := newFilesystem(config.Configuration{HomeDirectory: home})

			// Act: RealPath resolves the path without any I/O.
			base, ok := fs.(*afero.BasePathFs)
			require.True(t, ok)
			resolved, err := base.RealPath(tt.probe)

			// Assert: the resolved path sits under the configured home.
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(home, tt.probe), resolved)
		})
	}
}
