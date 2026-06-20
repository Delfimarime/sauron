package extension

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// MockBasedRegistry is a testify mock implementing Registry.
type MockBasedRegistry struct {
	mock.Mock
}

// Validate records the call and returns the configured error.
func (m *MockBasedRegistry) Validate(opts ...Option) error {
	return m.Called(opts).Error(0)
}

// Open records the call and returns the configured values.
func (m *MockBasedRegistry) Open(ctx context.Context, opts ...Option) (source.FileSystem, error) {
	args := m.Called(ctx, opts)

	var fs source.FileSystem
	if v := args.Get(0); v != nil {
		fs = v.(source.FileSystem)
	}

	return fs, args.Error(1)
}
