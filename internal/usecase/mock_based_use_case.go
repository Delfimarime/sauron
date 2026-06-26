package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockBasedUseCase is a reusable testify mock for any use case fitting the
// generic UseCase[I, P] shape.
type MockBasedUseCase[I, P any] struct {
	mock.Mock
}

// Execute records the call and returns the configured values.
func (m *MockBasedUseCase[I, P]) Execute(ctx context.Context, in I) (*P, error) {
	args := m.Called(ctx, in)

	var result *P
	if v := args.Get(0); v != nil {
		result = v.(*P)
	}

	return result, args.Error(1)
}
