package agent

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/delfimarime/sauron/internal/config"
)

// TestNewFxOptions verifies the provider filesystem is provided under its name
// tag and resolves to a non-nil afero.Fs.
func TestNewFxOptions(t *testing.T) {
	// Arrange.
	type in struct {
		fx.In
		Fs afero.Fs `name:"provider"`
	}
	var resolved in

	// Act.
	app := fx.New(
		fx.Supply(config.Configuration{UserHomeDirectory: t.TempDir()}),
		NewFxOptions(),
		fx.Populate(&resolved),
	)

	// Assert.
	require.NoError(t, app.Err())
	assert.NotNil(t, resolved.Fs)
}
