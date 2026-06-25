package storage

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// MockBasedRegistriesStore is a testify mock implementing RegistriesStore.
type MockBasedRegistriesStore struct {
	mock.Mock
}

// Get records the call and returns the configured values.
func (m *MockBasedRegistriesStore) Get(ctx context.Context) (*types.Registry, error) {
	args := m.Called(ctx)

	var registry *types.Registry
	if v := args.Get(0); v != nil {
		registry = v.(*types.Registry)
	}

	return registry, args.Error(1)
}

// Set records the call and returns the configured error.
func (m *MockBasedRegistriesStore) Set(ctx context.Context, r types.Registry) error {
	return m.Called(ctx, r).Error(0)
}

// Remove records the call and returns the configured error.
func (m *MockBasedRegistriesStore) Remove(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
