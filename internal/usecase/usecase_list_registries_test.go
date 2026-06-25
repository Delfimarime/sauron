package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// altName is the second registry name reused across the listing assertions.
const altName = "internal"

// reg builds a Registry with the given name, transport, and uri.
func reg(name string, transport types.Transport, uri string) types.Registry {
	return types.Registry{
		Metadata: types.Metadata{Name: name},
		Spec:     types.RegistrySpec{Transport: transport, URI: uri},
	}
}

// newListUseCase wires a list use case over a fresh store mock.
func newListUseCase(store storage.RegistriesStore) *ListRegistriesUseCase {
	return NewListRegistriesUseCase(ListRegistriesUseCaseParams{
		Registries: store,
		Logger:     zap.NewNop(),
	})
}

// TestListRegistriesSuccess asserts the stored registries are returned verbatim;
// filtering and ordering are now view concerns of the client.
func TestListRegistriesSuccess(t *testing.T) {
	// Arrange.
	stored := []types.Registry{
		reg(altName, types.TransportHTTP, "https://reg.example.com/"),
		reg(testName, types.TransportGit, "git@github.com:acme/artifacts.git"),
	}
	store := &storage.MockBasedRegistriesStore{}
	store.On("List", mock.Anything).Return(stored, nil)
	uc := newListUseCase(store)

	// Act.
	result, err := uc.Execute(context.Background(), ListRegistriesInput{})

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, stored, result.Registries)
}

// TestListRegistriesEmpty asserts an empty store yields an empty result, not an
// error.
func TestListRegistriesEmpty(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("List", mock.Anything).Return(nil, nil)
	uc := newListUseCase(store)

	// Act.
	result, err := uc.Execute(context.Background(), ListRegistriesInput{})

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Registries)
}

// TestListRegistriesIOError asserts a failing read classifies as an io failure.
func TestListRegistriesIOError(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("List", mock.Anything).Return(nil, errors.New("registries.yaml is unreadable"))
	uc := newListUseCase(store)

	// Act.
	result, err := uc.Execute(context.Background(), ListRegistriesInput{})

	// Assert.
	assert.Nil(t, result)
	_ = asUseCaseError(t, err, TypeIO)
}
