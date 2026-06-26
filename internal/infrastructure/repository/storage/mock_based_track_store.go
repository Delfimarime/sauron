package storage

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// MockBasedTrackStore is a testify mock implementing TrackStore.
type MockBasedTrackStore struct {
	mock.Mock
}

// List records the call and returns the configured values.
func (m *MockBasedTrackStore) List(ctx context.Context) ([]types.Artifact, error) {
	args := m.Called(ctx)

	var artifacts []types.Artifact
	if v := args.Get(0); v != nil {
		artifacts = v.([]types.Artifact)
	}

	return artifacts, args.Error(1)
}

// Update records the call and returns the configured error.
func (m *MockBasedTrackStore) Update(ctx context.Context, artifact types.Artifact) error {
	return m.Called(ctx, artifact).Error(0)
}
