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

// acmeRegistry is a stored registry fixture the delete tests resolve.
func acmeRegistry() *types.Registry {
	return &types.Registry{Metadata: types.Metadata{Name: testName}}
}

// newDeleteUseCase builds the use case over a mock store and the no-op cascade.
func newDeleteUseCase(store storage.RegistriesStore) *DeleteRegistryUseCase {
	return NewDeleteRegistryUseCase(DeleteRegistryUseCaseParams{
		Registries: store,
		Cascade:    NewUninstallByRegistryAction(UninstallByRegistryActionParams{Logger: zap.NewNop()}),
		Logger:     zap.NewNop(),
	})
}

// runDelete executes the use case, returning the result and the error.
func runDelete(uc *DeleteRegistryUseCase, name string, dryRun bool) (*DeleteRegistryResult, error) {
	return uc.Execute(context.Background(), DeleteRegistryInput{Name: name, DryRun: dryRun})
}

// TestDeleteRegistryRemovesAndReports removes the registry and reports the
// applied outcome with the empty (no-op) cascade plan.
func TestDeleteRegistryRemovesAndReports(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, testName).Return(acmeRegistry(), nil)
	store.On("Remove", mock.Anything, testName).Return(nil)
	uc := newDeleteUseCase(store)

	// Act.
	result, err := runDelete(uc, testName, false)

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Existed)
	assert.False(t, result.DryRun)
	require.NotNil(t, result.Plan)
	assert.Equal(t, 0, result.Plan.Total())
	store.AssertExpectations(t)
}

// TestDeleteRegistryNotFoundIsSuccess reports nothing existed and never removes
// (FR-005): a missing registry is exit 0, not an error.
func TestDeleteRegistryNotFoundIsSuccess(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, "ghost").Return(nil, nil)
	uc := newDeleteUseCase(store)

	// Act.
	result, err := runDelete(uc, "ghost", false)

	// Assert: success, Existed false, and Remove was never called.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Existed)
	store.AssertNotCalled(t, "Remove", mock.Anything, mock.Anything)
	store.AssertExpectations(t)
}

// TestDeleteRegistryDryRunWritesNothing previews the plan but never removes (FR-004).
func TestDeleteRegistryDryRunWritesNothing(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, testName).Return(acmeRegistry(), nil)
	uc := newDeleteUseCase(store)

	// Act.
	result, err := runDelete(uc, testName, true)

	// Assert: success, DryRun true, and Remove was never called.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Existed)
	assert.True(t, result.DryRun)
	store.AssertNotCalled(t, "Remove", mock.Anything, mock.Anything)
	store.AssertExpectations(t)
}

// TestDeleteRegistryLookupFailsIsIOError classifies a state-read failure as io.
func TestDeleteRegistryLookupFailsIsIOError(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, testName).Return(nil, errors.New("boom"))
	uc := newDeleteUseCase(store)

	// Act.
	result, err := runDelete(uc, testName, false)

	// Assert.
	assert.Nil(t, result)
	_ = asUseCaseError(t, err, TypeIO)
}

// TestDeleteRegistryRemoveFailsIsIOError classifies a removal write failure as io.
func TestDeleteRegistryRemoveFailsIsIOError(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, testName).Return(acmeRegistry(), nil)
	store.On("Remove", mock.Anything, testName).Return(errors.New("disk full"))
	uc := newDeleteUseCase(store)

	// Act.
	result, err := runDelete(uc, testName, false)

	// Assert.
	assert.Nil(t, result)
	_ = asUseCaseError(t, err, TypeIO)
}
