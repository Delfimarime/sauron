package storage

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// MockBasedProvidersStore is a testify mock implementing ProvidersStore.
type MockBasedProvidersStore struct {
	mock.Mock
}

// Get records the call and returns the configured values.
func (m *MockBasedProvidersStore) Get(ctx context.Context) (*types.Provider, error) {
	args := m.Called(ctx)

	var provider *types.Provider
	if v := args.Get(0); v != nil {
		provider = v.(*types.Provider)
	}

	return provider, args.Error(1)
}

// Set records the call and returns the configured error.
func (m *MockBasedProvidersStore) Set(ctx context.Context, p types.Provider) error {
	return m.Called(ctx, p).Error(0)
}
