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

// acmeRegistry is a stored registry fixture the unset tests resolve.
func acmeRegistry() *types.Registry {
	return &types.Registry{Spec: types.RegistrySpec{URI: testHTTPURI}}
}

// newUnsetUseCase builds the use case over a mock store.
func newUnsetUseCase(store storage.RegistriesStore) *UnsetRegistryUseCase {
	return NewUnsetRegistryUseCase(UnsetRegistryUseCaseParams{
		Registries: store,
		Logger:     zap.NewNop(),
	})
}

// runUnset executes the use case, returning the result and the error.
func runUnset(uc *UnsetRegistryUseCase, dryRun bool) (*UnsetRegistryResult, error) {
	return uc.Execute(context.Background(), UnsetRegistryInput{DryRun: dryRun})
}

// TestUnsetRegistryRemovesAndReports removes the registry and reports the removed
// outcome.
func TestUnsetRegistryRemovesAndReports(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("Get", mock.Anything).Return(acmeRegistry(), nil)
	store.On("Remove", mock.Anything).Return(nil)
	uc := newUnsetUseCase(store)

	// Act.
	result, err := runUnset(uc, false)

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, UnsetRemoved, result.Outcome)
	store.AssertExpectations(t)
}

// TestUnsetRegistryNotConfiguredIsSuccess reports the nothing outcome and never
// removes: no configured registry is exit 0, not an error.
func TestUnsetRegistryNotConfiguredIsSuccess(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("Get", mock.Anything).Return(nil, nil)
	uc := newUnsetUseCase(store)

	// Act.
	result, err := runUnset(uc, false)

	// Assert: success, the nothing outcome, and Remove was never called.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, UnsetNothing, result.Outcome)
	store.AssertNotCalled(t, "Remove", mock.Anything)
	store.AssertExpectations(t)
}

// TestUnsetRegistryDryRunWritesNothing previews the removal but never removes.
func TestUnsetRegistryDryRunWritesNothing(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("Get", mock.Anything).Return(acmeRegistry(), nil)
	uc := newUnsetUseCase(store)

	// Act.
	result, err := runUnset(uc, true)

	// Assert: success, the preview outcome, and Remove was never called.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, UnsetPreview, result.Outcome)
	store.AssertNotCalled(t, "Remove", mock.Anything)
	store.AssertExpectations(t)
}

// TestUnsetRegistryReadFailsIsIOError classifies a state-read failure as io.
func TestUnsetRegistryReadFailsIsIOError(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("Get", mock.Anything).Return(nil, errors.New("boom"))
	uc := newUnsetUseCase(store)

	// Act.
	_, err := runUnset(uc, false)

	// Assert.
	var ucErr *Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, TypeIO, ucErr.Type)
}

// TestUnsetRegistryRemoveFailsIsIOError classifies a removal write failure as io.
func TestUnsetRegistryRemoveFailsIsIOError(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("Get", mock.Anything).Return(acmeRegistry(), nil)
	store.On("Remove", mock.Anything).Return(errors.New("disk full"))
	uc := newUnsetUseCase(store)

	// Act.
	_, err := runUnset(uc, false)

	// Assert.
	var ucErr *Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, TypeIO, ucErr.Type)
}
