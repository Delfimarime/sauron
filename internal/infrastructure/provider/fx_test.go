package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

// TestNewFxOptions verifies the (bootstrap-empty) provider options compose into a
// valid container without error.
func TestNewFxOptions(t *testing.T) {
	// Arrange + Act.
	app := fx.New(NewFxOptions())

	// Assert.
	assert.NoError(t, app.Err())
}
