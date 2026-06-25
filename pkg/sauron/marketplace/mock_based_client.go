package marketplace

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockBasedClient is a testify mock implementing Client.
type MockBasedClient struct {
	mock.Mock
}

// Skills records the call and returns the configured artifact client.
func (m *MockBasedClient) Skills() ArtifactClient {
	return m.artifactClient(m.Called())
}

// Agents records the call and returns the configured artifact client.
func (m *MockBasedClient) Agents() ArtifactClient {
	return m.artifactClient(m.Called())
}

// artifactClient unwraps an ArtifactClient from recorded arguments.
func (m *MockBasedClient) artifactClient(args mock.Arguments) ArtifactClient {
	if v := args.Get(0); v != nil {
		return v.(ArtifactClient)
	}
	return nil
}

// MockBasedArtifactClient is a testify mock implementing ArtifactClient.
type MockBasedArtifactClient struct {
	mock.Mock
}

// List records the call and returns the configured values.
func (m *MockBasedArtifactClient) List(ctx context.Context, opts ...ListOption) (*ArtifactList, error) {
	args := m.Called(ctx, opts)

	var list *ArtifactList
	if v := args.Get(0); v != nil {
		list = v.(*ArtifactList)
	}

	return list, args.Error(1)
}
