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

// describe-test literals, named to satisfy goconst across the package.
const (
	userRef  = "${env:ACME_USER}"
	tokenRef = "${env:ACME_TOKEN}"
	gitURI   = "git@github.com:acme/artifacts.git"

	createdStamp = "2026-06-21T07:30:00Z"
	updatedStamp = "2026-06-22T08:00:00Z"
)

// describeFixture bundles the describe use case and its mocked store.
type describeFixture struct {
	uc    *DescribeRegistryUseCase
	store *storage.MockBasedRegistriesStore
}

// newDescribeFixture wires a describe use case over a fresh store mock, verifying
// its expectations on cleanup so an unused stub fails the test.
func newDescribeFixture(t *testing.T) *describeFixture {
	t.Helper()
	store := &storage.MockBasedRegistriesStore{}
	t.Cleanup(func() { store.AssertExpectations(t) })
	return &describeFixture{
		store: store,
		uc: NewDescribeRegistryUseCase(DescribeRegistryUseCaseParams{
			Registries: store,
			Logger:     zap.NewNop(),
		}),
	}
}

// fullRegistry is a registry populated across every describable field.
func fullRegistry() *types.Registry {
	return &types.Registry{
		Metadata: types.Metadata{
			CreatedAt:     createdStamp,
			LastUpdatedAt: updatedStamp,
		},
		Spec: types.RegistrySpec{
			Transport: types.TransportGit,
			Source:    gitURI,
			Revision:  "v1.2.0",
			Credentials: &types.Credentials{
				Username: userRef,
				Password: tokenRef,
			},
			Timeout: "45s",
		},
	}
}

// TestDescribeRegistrySuccess asserts the get pipeline returns the configured
// registry for the client to project; field selection is a presentation concern.
func TestDescribeRegistrySuccess(t *testing.T) {
	// Arrange.
	f := newDescribeFixture(t)
	f.store.On("Get", mock.Anything).Return(fullRegistry(), nil)

	// Act.
	registry, err := f.uc.Execute(context.Background(), DescribeRegistryRequest{})

	// Assert: the full record round-trips so a dropped field is caught.
	require.NoError(t, err)
	require.NotNil(t, registry)
	assert.Equal(t, fullRegistry(), registry)
}

// TestDescribeRegistryFailure covers the not-found and io classifications.
func TestDescribeRegistryFailure(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// found is the record Get returns.
		found *types.Registry
		// getErr is the error Get returns.
		getErr error
		// wantType is the expected error classification.
		wantType Type
	}{
		{
			name:     "no registry configured is not found",
			wantType: TypeNotFound,
		},
		{
			name:     "store failure is io",
			getErr:   errors.New("disk gone"),
			wantType: TypeIO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			f := newDescribeFixture(t)
			f.store.On("Get", mock.Anything).Return(tt.found, tt.getErr)

			// Act.
			_, err := f.uc.Execute(context.Background(), DescribeRegistryRequest{})

			// Assert.
			var ucErr *Error
			require.ErrorAs(t, err, &ucErr)
			assert.Equal(t, tt.wantType, ucErr.Type)
		})
	}
}
