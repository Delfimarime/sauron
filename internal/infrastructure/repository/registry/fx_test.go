package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"

	"github.com/delfimarime/sauron/pkg/sauron/extension"
)

// TestNewFxOptions verifies the registry options compose into a valid container
// and expose the transports as named extension.Registry values.
func TestNewFxOptions(t *testing.T) {
	// Arrange.
	type registries struct {
		fx.In
		Git  extension.Registry `name:"registry.git"`
		HTTP extension.Registry `name:"registry.http"`
	}

	var resolved registries

	// Act.
	app := fx.New(
		NewFxOptions(),
		fx.Populate(&resolved),
	)

	// Assert.
	assert.NoError(t, app.Err())
	assert.NotNil(t, resolved.Git)
	assert.NotNil(t, resolved.HTTP)
}
