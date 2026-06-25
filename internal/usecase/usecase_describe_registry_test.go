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
	describeName = "acme"
	userRef      = "${env:ACME_USER}"
	tokenRef     = "${env:ACME_TOKEN}"
	gitURI       = "git@github.com:acme/artifacts.git"

	createdStamp = "2026-06-21T07:30:00Z"
	updatedStamp = "2026-06-22T08:00:00Z"
)

// newDescribeUseCase wires a describe use case over a fresh store mock.
func newDescribeUseCase(store storage.RegistriesStore) *DescribeRegistryUseCase {
	return NewDescribeRegistryUseCase(DescribeRegistryUseCaseParams{
		Registries: store,
		Logger:     zap.NewNop(),
	})
}

// fullRegistry is a registry populated across every describable field.
func fullRegistry() *types.Registry {
	return &types.Registry{
		Metadata: types.Metadata{
			Name:                 describeName,
			CreationTimestamp:    createdStamp,
			LastUpdatedTimestamp: updatedStamp,
		},
		Spec: types.RegistrySpec{
			Transport: types.TransportGit,
			URI:       gitURI,
			Ref:       "v1.2.0",
			Auth: &types.Auth{
				Username: userRef,
				Password: tokenRef,
			},
			Timeout: "45s",
		},
	}
}

// TestDescribeRegistrySuccess asserts the found registry is returned verbatim;
// field selection and rendering are view concerns of the client.
func TestDescribeRegistrySuccess(t *testing.T) {
	// Arrange.
	store := &storage.MockBasedRegistriesStore{}
	store.On("FindByName", mock.Anything, describeName).Return(fullRegistry(), nil)
	uc := newDescribeUseCase(store)

	// Act.
	result, err := uc.Execute(context.Background(), DescribeRegistryInput{Name: describeName})

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, fullRegistry(), result)
}

// TestDescribeRegistryFailure covers the not-found and io classifications.
func TestDescribeRegistryFailure(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// found is the record FindByName returns.
		found *types.Registry
		// findErr is the error FindByName returns.
		findErr error
		// wantType is the expected error classification.
		wantType Type
	}{
		{
			name:     "unknown name is not found",
			wantType: TypeNotFound,
		},
		{
			name:     "store failure is io",
			findErr:  errors.New("disk gone"),
			wantType: TypeIO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			store := &storage.MockBasedRegistriesStore{}
			store.On("FindByName", mock.Anything, describeName).Return(tt.found, tt.findErr)
			uc := newDescribeUseCase(store)

			// Act.
			result, err := uc.Execute(context.Background(), DescribeRegistryInput{Name: describeName})

			// Assert.
			assert.Nil(t, result)
			_ = asUseCaseError(t, err, tt.wantType)
		})
	}
}
