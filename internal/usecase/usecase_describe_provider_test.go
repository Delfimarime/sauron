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

// describeProviderFixture bundles the describe-provider use case and its mocked
// store.
type describeProviderFixture struct {
	uc    *DescribeProviderUseCase
	store *storage.MockBasedProvidersStore
}

// newDescribeProviderFixture wires a describe-provider use case over a fresh store
// mock.
func newDescribeProviderFixture() *describeProviderFixture {
	store := &storage.MockBasedProvidersStore{}
	return &describeProviderFixture{
		store: store,
		uc: NewDescribeProviderUseCase(DescribeProviderUseCaseParams{
			Providers: store,
			Logger:    zap.NewNop(),
		}),
	}
}

// fullProvider is a provider populated across every describable field.
func fullProvider() *types.Provider {
	return &types.Provider{
		Metadata: types.Metadata{
			Name:          types.ProviderClaude,
			CreatedAt:     createdStamp,
			LastUpdatedAt: updatedStamp,
		},
		Spec: types.ProviderSpec{
			LastSyncedAt:      "2026-06-25T09:15:00Z",
			LastSyncAttemptAt: "2026-06-26T06:00:00Z",
		},
	}
}

// TestDescribeProviderSuccess asserts the get pipeline returns the configured
// provider for the client to project; field selection is a presentation concern.
func TestDescribeProviderSuccess(t *testing.T) {
	// Arrange.
	f := newDescribeProviderFixture()
	f.store.On("Get", mock.Anything).Return(fullProvider(), nil)

	// Act.
	provider, err := f.uc.Execute(context.Background(), DescribeProviderInput{})

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, provider)
	assert.Equal(t, types.ProviderClaude, provider.Metadata.Name)
}

// TestDescribeProviderNoneSet asserts that no provider configured is not an error:
// the use case returns (nil, nil) so the command can report none-set and exit 0.
func TestDescribeProviderNoneSet(t *testing.T) {
	// Arrange.
	f := newDescribeProviderFixture()
	f.store.On("Get", mock.Anything).Return(nil, nil)

	// Act.
	provider, err := f.uc.Execute(context.Background(), DescribeProviderInput{})

	// Assert.
	require.NoError(t, err)
	assert.Nil(t, provider)
}

// TestDescribeProviderReadError asserts a store failure surfaces as an io error.
func TestDescribeProviderReadError(t *testing.T) {
	// Arrange.
	f := newDescribeProviderFixture()
	f.store.On("Get", mock.Anything).Return(nil, errors.New("disk gone"))

	// Act.
	_, err := f.uc.Execute(context.Background(), DescribeProviderInput{})

	// Assert.
	var ucErr *Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, TypeIO, ucErr.Type)
}
