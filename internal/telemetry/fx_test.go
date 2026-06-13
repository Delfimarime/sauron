package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// TestNewLogger verifies the constructor returns a usable ECS-encoded logger.
func TestNewLogger(t *testing.T) {
	// Arrange + Act.
	log := NewLogger()

	// Assert.
	require.NotNil(t, log)
	assert.NotPanics(t, func() { log.Info("smoke", zap.String(FieldEventAction, "test")) })
}

// TestNewFxOptions verifies the *zap.Logger is provided into the container.
func TestNewFxOptions(t *testing.T) {
	// Arrange.
	var log *zap.Logger

	// Act.
	app := fx.New(NewFxOptions(), fx.Populate(&log))

	// Assert.
	require.NoError(t, app.Err())
	assert.NotNil(t, log)
}
