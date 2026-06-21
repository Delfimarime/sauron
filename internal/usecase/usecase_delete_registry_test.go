package usecase

import (
	"bytes"
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

// runDelete executes the use case, returning the captured output and the error.
func runDelete(uc *DeleteRegistryUseCase, name string, dryRun bool) (string, error) {
	var out bytes.Buffer
	request := NewDeleteRegistryRequest(context.Background(), &out)
	request.Name = name
	request.DryRun = dryRun
	err := uc.Execute(request)
	return out.String(), err
}

// TestDeleteRegistryRemovesAndReports removes the registry and reports the summary
// with zero artifacts (the cascade is a no-op).
func TestDeleteRegistryRemovesAndReports(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, testName).Return(acmeRegistry(), nil)
	store.On("Remove", mock.Anything, testName).Return(nil)
	uc := newDeleteUseCase(store)

	// Act.
	out, err := runDelete(uc, testName, false)

	// Assert.
	require.NoError(t, err)
	assert.Contains(t, out, `registry "acme" removed; 0 artifacts removed`)
	store.AssertExpectations(t)
}

// TestDeleteRegistryNotFoundIsSuccess reports nothing was deleted and never removes
// (FR-005): a missing registry is exit 0, not an error.
func TestDeleteRegistryNotFoundIsSuccess(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, "ghost").Return(nil, nil)
	uc := newDeleteUseCase(store)

	// Act.
	out, err := runDelete(uc, "ghost", false)

	// Assert: success, the message, and Remove was never called.
	require.NoError(t, err)
	assert.Contains(t, out, "nothing was deleted")
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
	out, err := runDelete(uc, testName, true)

	// Assert: success, a preview line, and Remove was never called.
	require.NoError(t, err)
	assert.Contains(t, out, "would be removed")
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
	_, err := runDelete(uc, testName, false)

	// Assert.
	var ucErr *Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, TypeIO, ucErr.Type)
}

// TestDeleteRegistryRemoveFailsIsIOError classifies a removal write failure as io.
func TestDeleteRegistryRemoveFailsIsIOError(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, testName).Return(acmeRegistry(), nil)
	store.On("Remove", mock.Anything, testName).Return(errors.New("disk full"))
	uc := newDeleteUseCase(store)

	// Act.
	_, err := runDelete(uc, testName, false)

	// Assert.
	var ucErr *Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, TypeIO, ucErr.Type)
}

// TestDeleteRegistryRendersNonEmptyGroups renders only the non-empty kind groups,
// each entry "-"-prefixed, followed by the summary count.
func TestDeleteRegistryRendersNonEmptyGroups(t *testing.T) {
	// Arrange.
	uc := newDeleteUseCase(&storage.MockBasedRegistriesStore{})
	plan := &DeleteArtifactsByRegistryResponse{
		Skills: []string{"sauron-acme-go-style"},
		Agents: []string{"sauron-acme-code-reviewer"},
	}

	// Act.
	out := uc.renderPlan(plan)

	// Assert: skills and agents groups render, personas (empty) does not.
	assert.Contains(t, out, "skills:\n  - sauron-acme-go-style\n")
	assert.Contains(t, out, "agents:\n  - sauron-acme-code-reviewer\n")
	assert.NotContains(t, out, "personas:")
}

// TestRemovalPlanSummaryReflectsTotal counts artifacts across kinds in the summary.
func TestRemovalPlanSummaryReflectsTotal(t *testing.T) {
	// Arrange: a store that resolves the registry and accepts the removal.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, testName).Return(acmeRegistry(), nil)
	store.On("Remove", mock.Anything, testName).Return(nil)
	uc := newDeleteUseCase(store)
	plan := &DeleteArtifactsByRegistryResponse{Skills: []string{"a", "b"}}

	// Act: the summary helper is exercised directly for a non-empty plan.
	var out bytes.Buffer
	request := NewDeleteRegistryRequest(context.Background(), &out)
	request.Name = testName
	require.NoError(t, uc.reportRemoved(request, plan))

	// Assert.
	assert.Contains(t, out.String(), `registry "acme" removed; 2 artifacts removed`)
}
