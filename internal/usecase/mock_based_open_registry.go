package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// MockBasedOpenRegistry is a testify mock implementing OpenRegistry.
type MockBasedOpenRegistry struct {
	mock.Mock
}

// Execute records the call and returns the configured values.
func (m *MockBasedOpenRegistry) Execute(ctx context.Context, registry types.Registry) (source.FileSystem, error) {
	args := m.Called(ctx, registry)

	var fs source.FileSystem
	if v := args.Get(0); v != nil {
		fs = v.(source.FileSystem)
	}

	return fs, args.Error(1)
}
