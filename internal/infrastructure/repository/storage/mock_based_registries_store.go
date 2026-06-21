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

// FindByName records the call and returns the configured values.
func (m *MockBasedRegistriesStore) FindByName(ctx context.Context, name string) (*types.Registry, error) {
	args := m.Called(ctx, name)

	var registry *types.Registry
	if v := args.Get(0); v != nil {
		registry = v.(*types.Registry)
	}

	return registry, args.Error(1)
}

// Add records the call and returns the configured error.
func (m *MockBasedRegistriesStore) Add(ctx context.Context, r types.Registry) error {
	return m.Called(ctx, r).Error(0)
}

// List records the call and returns the configured values.
func (m *MockBasedRegistriesStore) List(ctx context.Context) ([]types.Registry, error) {
	args := m.Called(ctx)

	var registries []types.Registry
	if v := args.Get(0); v != nil {
		registries = v.([]types.Registry)
	}

	return registries, args.Error(1)
}
