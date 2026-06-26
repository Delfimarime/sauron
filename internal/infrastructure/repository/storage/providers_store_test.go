package storage

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// newTestProvidersStore builds a ProvidersStore over an in-memory filesystem.
func newTestProvidersStore(t *testing.T) (ProvidersStore, *Store) {
	t.Helper()
	store, _ := newTestStore(t)
	return NewProvidersStore(store), store
}

// TestProvidersStoreGetAbsent returns nil when no provider is configured.
func TestProvidersStoreGetAbsent(t *testing.T) {
	// Arrange.
	providers, _ := newTestProvidersStore(t)

	// Act.
	got, err := providers.Get(context.Background())

	// Assert.
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestProvidersStoreSetRoundTrip stamps the envelope and round-trips the name.
func TestProvidersStoreSetRoundTrip(t *testing.T) {
	// Arrange.
	providers, _ := newTestProvidersStore(t)
	in := types.Provider{Metadata: types.Metadata{Name: types.ProviderClaude}}

	// Act.
	require.NoError(t, providers.Set(context.Background(), in))
	got, err := providers.Get(context.Background())

	// Assert: the envelope is stamped and the name survives intact.
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, types.APIVersion, got.APIVersion)
	assert.Equal(t, types.KindProvider, got.Kind)
	assert.Equal(t, types.ProviderClaude, got.Metadata.Name)
}

// TestProvidersStoreSetReplaces keeps exactly one provider: a second Set
// overwrites the first rather than accumulating a second document.
func TestProvidersStoreSetReplaces(t *testing.T) {
	// Arrange: claude already configured.
	providers, _ := newTestProvidersStore(t)
	require.NoError(t, providers.Set(context.Background(), types.Provider{
		Metadata: types.Metadata{Name: types.ProviderClaude},
	}))

	// Act: set a different provider.
	require.NoError(t, providers.Set(context.Background(), types.Provider{
		Metadata: types.Metadata{Name: types.ProviderZencoder},
	}))

	// Assert: only the replacement remains.
	got, err := providers.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, types.ProviderZencoder, got.Metadata.Name)
}

// TestProvidersStoreSetCoexistsWithRegistry keeps the Registry document in
// settings.yaml when the Provider is written into the same file.
func TestProvidersStoreSetCoexistsWithRegistry(t *testing.T) {
	// Arrange: a registry already configured in settings.yaml.
	store, _ := newTestStore(t)
	registries := NewRegistriesStore(store)
	providers := NewProvidersStore(store)
	require.NoError(t, registries.Set(context.Background(), types.Registry{
		Spec: types.RegistrySpec{Transport: types.TransportGit, Source: acmeURI},
	}))

	// Act.
	require.NoError(t, providers.Set(context.Background(), types.Provider{
		Metadata: types.Metadata{Name: types.ProviderZencoder},
	}))

	// Assert: both documents resolve from the shared file.
	gotProvider, err := providers.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, gotProvider)
	assert.Equal(t, types.ProviderZencoder, gotProvider.Metadata.Name)

	gotRegistry, err := registries.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, gotRegistry)
	assert.Equal(t, acmeURI, gotRegistry.Spec.Source)
}

// TestProvidersStoreGetPropagatesReadError surfaces a read failure from the store.
func TestProvidersStoreGetPropagatesReadError(t *testing.T) {
	// Arrange: a malformed stream makes the underlying read fail.
	providers, store := newTestProvidersStore(t)
	require.NoError(t, afero.WriteFile(store.fs, registriesFile, []byte("\tnot: [valid"), filePerm))

	// Act.
	_, err := providers.Get(context.Background())

	// Assert.
	require.Error(t, err)
}

// TestMockBasedProvidersStore exercises the testify mock.
func TestMockBasedProvidersStore(t *testing.T) {
	// Arrange.
	m := &MockBasedProvidersStore{}
	want := &types.Provider{Metadata: types.Metadata{Name: types.ProviderClaude}}
	ctx := context.Background()
	m.On("Get", ctx).Return(want, nil)
	m.On("Set", ctx, types.Provider{}).Return(nil)

	// Act.
	got, getErr := m.Get(ctx)
	setErr := m.Set(ctx, types.Provider{})

	// Assert.
	require.NoError(t, getErr)
	require.NoError(t, setErr)
	assert.Same(t, want, got)
	m.AssertExpectations(t)
}

// TestMockBasedProvidersStoreNilGet returns nil without panicking when Get is
// configured to yield no provider.
func TestMockBasedProvidersStoreNilGet(t *testing.T) {
	// Arrange.
	m := &MockBasedProvidersStore{}
	ctx := context.Background()
	m.On("Get", ctx).Return(nil, nil)

	// Act.
	got, err := m.Get(ctx)

	// Assert.
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestNewProvidersStoreType confirms the constructor returns the interface.
func TestNewProvidersStoreType(t *testing.T) {
	// Arrange.
	fs := afero.NewMemMapFs()
	store, err := NewStore(fs)
	require.NoError(t, err)

	// Act.
	providers := NewProvidersStore(store)

	// Assert.
	require.NotNil(t, providers)
}
