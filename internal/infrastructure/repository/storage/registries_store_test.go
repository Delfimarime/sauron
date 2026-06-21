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

// TestRegistriesStoreFindByNameAbsent returns nil when no registry matches.
func TestRegistriesStoreFindByNameAbsent(t *testing.T) {
	// Arrange.
	registries, _ := newTestRegistriesStore(t)

	// Act.
	got, err := registries.FindByName(context.Background(), acmeName)

	// Assert.
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestRegistriesStoreAddRoundTrip stamps the envelope and round-trips spec.ref.
func TestRegistriesStoreAddRoundTrip(t *testing.T) {
	// Arrange.
	registries, _ := newTestRegistriesStore(t)
	in := types.Registry{
		Metadata: types.Metadata{Name: acmeName},
		Spec: types.RegistrySpec{
			Transport: types.TransportGit,
			URI:       "https://example.com/acme.git",
			Ref:       "release-1.2.3",
		},
	}

	// Act.
	require.NoError(t, registries.Add(context.Background(), in))
	got, err := registries.FindByName(context.Background(), acmeName)

	// Assert: the envelope is stamped and spec.ref survives intact.
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, types.APIVersion, got.APIVersion)
	assert.Equal(t, types.KindRegistry, got.Kind)
	assert.Equal(t, acmeName, got.Metadata.Name)
	assert.Equal(t, types.TransportGit, got.Spec.Transport)
	assert.Equal(t, "release-1.2.3", got.Spec.Ref)
}

// TestRegistriesStoreAddPersistsValidDocument writes a document that passes the
// schema on a subsequent validated read.
func TestRegistriesStoreAddPersistsValidDocument(t *testing.T) {
	// Arrange.
	registries, store := newTestRegistriesStore(t)
	in := types.Registry{
		Metadata: types.Metadata{Name: "beta"},
		Spec: types.RegistrySpec{
			Transport: types.TransportHTTP,
			URI:       "https://example.com/beta",
		},
	}

	// Act.
	require.NoError(t, registries.Add(context.Background(), in))
	node, err := store.FindOne(context.Background(), types.KindRegistry, "beta")

	// Assert: FindOne validates on read, so a non-nil node means it passed.
	require.NoError(t, err)
	require.NotNil(t, node)
}

// TestRegistriesStoreRemoveDropsRegistry adds a registry then removes it, so a
// subsequent lookup returns nil.
func TestRegistriesStoreRemoveDropsRegistry(t *testing.T) {
	// Arrange.
	registries, _ := newTestRegistriesStore(t)
	in := types.Registry{
		Metadata: types.Metadata{Name: acmeName},
		Spec: types.RegistrySpec{
			Transport: types.TransportGit,
			URI:       "https://example.com/acme.git",
		},
	}
	require.NoError(t, registries.Add(context.Background(), in))

	// Act.
	require.NoError(t, registries.Remove(context.Background(), acmeName))

	// Assert.
	got, err := registries.FindByName(context.Background(), acmeName)
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestRegistriesStoreRemoveAbsentIsNoOp removing an unknown registry succeeds.
func TestRegistriesStoreRemoveAbsentIsNoOp(t *testing.T) {
	// Arrange.
	registries, _ := newTestRegistriesStore(t)

	// Act.
	err := registries.Remove(context.Background(), "ghost")

	// Assert.
	require.NoError(t, err)
}

// TestMockBasedRegistriesStore exercises the testify mock.
func TestMockBasedRegistriesStore(t *testing.T) {
	// Arrange.
	m := &MockBasedRegistriesStore{}
	want := &types.Registry{Metadata: types.Metadata{Name: acmeName}}
	ctx := context.Background()
	m.On("FindByName", ctx, acmeName).Return(want, nil)
	m.On("Add", ctx, types.Registry{}).Return(nil)
	m.On("Remove", ctx, acmeName).Return(nil)

	// Act.
	got, findErr := m.FindByName(ctx, acmeName)
	addErr := m.Add(ctx, types.Registry{})
	removeErr := m.Remove(ctx, acmeName)

	// Assert.
	require.NoError(t, findErr)
	require.NoError(t, addErr)
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
