package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestUninstallByRegistryActionIsNoOp asserts the no-op contract: the action
// returns an empty plan and no error for any registry. 0007 replaces the body.
func TestUninstallByRegistryActionIsNoOp(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// registry is the source whose artifacts the cascade would remove.
		registry string
	}{
		{name: "named registry", registry: testName},
		{name: "empty name", registry: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			action := NewUninstallByRegistryAction(UninstallByRegistryActionParams{Logger: zap.NewNop()})

			// Act.
			plan, err := action.Execute(context.Background(), tt.registry)

			// Assert: empty plan, nil error, zero total.
			require.NoError(t, err)
			require.NotNil(t, plan)
			assert.Empty(t, plan.Skills)
			assert.Empty(t, plan.Agents)
			assert.Empty(t, plan.Personas)
			assert.Equal(t, 0, plan.Total())
		})
	}
}

// TestRemovalPlanTotal sums the artifacts across every kind.
func TestRemovalPlanTotal(t *testing.T) {
	// Arrange.
	plan := DeleteArtifactsByRegistryResponse{
		Skills:   []string{"a", "b"},
		Agents:   []string{"c"},
		Personas: []string{"d", "e", "f"},
	}

	// Act + Assert.
	assert.Equal(t, 6, plan.Total())
}
