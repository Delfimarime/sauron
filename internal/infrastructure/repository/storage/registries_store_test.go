package storage

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// newTestRegistriesStore builds a RegistriesStore over an in-memory filesystem.
func newTestRegistriesStore(t *testing.T) (RegistriesStore, *Store) {
	t.Helper()
	store, _ := newTestStore(t)
	return NewRegistriesStore(store), store
}

// TestRegistriesStoreGetAbsent returns nil when no registry is configured.
func TestRegistriesStoreGetAbsent(t *testing.T) {
	// Arrange.
	registries, _ := newTestRegistriesStore(t)

	// Act.
	got, err := registries.Get(context.Background())

	// Assert.
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestRegistriesStoreSetRoundTrip stamps the envelope and round-trips spec.ref.
func TestRegistriesStoreSetRoundTrip(t *testing.T) {
	// Arrange.
	registries, _ := newTestRegistriesStore(t)
	in := types.Registry{
		Spec: types.RegistrySpec{
			Transport: types.TransportGit,
			URI:       acmeURI,
			Ref:       "release-1.2.3",
		},
	}

	// Act.
	require.NoError(t, registries.Set(context.Background(), in))
	got, err := registries.Get(context.Background())

	// Assert: the envelope is stamped and spec.ref survives intact.
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, types.APIVersion, got.APIVersion)
	assert.Equal(t, types.KindRegistry, got.Kind)
	assert.Equal(t, types.TransportGit, got.Spec.Transport)
	assert.Equal(t, acmeURI, got.Spec.URI)
	assert.Equal(t, "release-1.2.3", got.Spec.Ref)
}

// TestRegistriesStoreSetReplaces keeps exactly one registry: a second Set
// overwrites the first rather than accumulating a second document.
func TestRegistriesStoreSetReplaces(t *testing.T) {
	// Arrange: an http registry already configured.
	registries, _ := newTestRegistriesStore(t)
	require.NoError(t, registries.Set(context.Background(), types.Registry{
		Spec: types.RegistrySpec{Transport: types.TransportHTTP, URI: "https://example.com/first"},
	}))

	// Act: set a different git registry.
	require.NoError(t, registries.Set(context.Background(), types.Registry{
		Spec: types.RegistrySpec{Transport: types.TransportGit, URI: acmeURI},
	}))

	// Assert: only the replacement remains.
	got, err := registries.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, types.TransportGit, got.Spec.Transport)
	assert.Equal(t, acmeURI, got.Spec.URI)
}

// TestRegistriesStoreSetPersistsValidDocument writes a document that passes the
// schema on a subsequent validated read.
func TestRegistriesStoreSetPersistsValidDocument(t *testing.T) {
	// Arrange.
	registries, store := newTestRegistriesStore(t)
	in := types.Registry{
		Spec: types.RegistrySpec{
			Transport: types.TransportHTTP,
			URI:       "https://example.com/beta",
		},
	}

	// Act.
	require.NoError(t, registries.Set(context.Background(), in))
	node, err := store.First(context.Background(), types.KindRegistry)

	// Assert: First validates on read, so a non-nil node means it passed.
	require.NoError(t, err)
	require.NotNil(t, node)
}

// TestRegistriesStoreRemoveDropsRegistry sets a registry then removes it, so a
// subsequent get returns nil.
func TestRegistriesStoreRemoveDropsRegistry(t *testing.T) {
	// Arrange.
	registries, _ := newTestRegistriesStore(t)
	require.NoError(t, registries.Set(context.Background(), types.Registry{
		Spec: types.RegistrySpec{Transport: types.TransportGit, URI: acmeURI},
	}))

	// Act.
	require.NoError(t, registries.Remove(context.Background()))

	// Assert.
	got, err := registries.Get(context.Background())
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestRegistriesStoreRemoveAbsentIsNoOp removing with no registry set succeeds.
func TestRegistriesStoreRemoveAbsentIsNoOp(t *testing.T) {
	// Arrange.
	registries, _ := newTestRegistriesStore(t)

	// Act.
	err := registries.Remove(context.Background())

	// Assert.
	require.NoError(t, err)
}

// TestRegistriesStoreGetPropagatesReadError surfaces a read failure from the store.
func TestRegistriesStoreGetPropagatesReadError(t *testing.T) {
	// Arrange: a malformed stream makes the underlying read fail.
	registries, store := newTestRegistriesStore(t)
	require.NoError(t, afero.WriteFile(store.fs, registriesFile, []byte("\tnot: [valid"), filePerm))

	// Act.
	_, err := registries.Get(context.Background())

	// Assert.
	require.Error(t, err)
}

// TestMockBasedRegistriesStore exercises the testify mock.
func TestMockBasedRegistriesStore(t *testing.T) {
	// Arrange.
	m := &MockBasedRegistriesStore{}
	want := &types.Registry{Spec: types.RegistrySpec{URI: acmeURI}}
	ctx := context.Background()
	m.On("Get", ctx).Return(want, nil)
	m.On("Set", ctx, types.Registry{}).Return(nil)
	m.On("Remove", ctx).Return(nil)

	// Act.
	got, getErr := m.Get(ctx)
	setErr := m.Set(ctx, types.Registry{})
	removeErr := m.Remove(ctx)

	// Assert.
	require.NoError(t, getErr)
	require.NoError(t, setErr)
	require.NoError(t, removeErr)
	assert.Same(t, want, got)
	m.AssertExpectations(t)
}

// TestNewRegistriesStoreType confirms the constructor returns the interface.
func TestNewRegistriesStoreType(t *testing.T) {
	// Arrange.
	fs := afero.NewMemMapFs()
	store, err := NewStore(fs)
	require.NoError(t, err)

	// Act.
	registries := NewRegistriesStore(store)

	// Assert.
	require.NotNil(t, registries)
}
