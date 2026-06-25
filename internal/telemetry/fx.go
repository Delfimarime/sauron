package telemetry

import (
	"go.uber.org/fx"
)

// NewFxOptions wires the zap logger.
func NewFxOptions() fx.Option {
	return fx.Provide(
		NewLogger,
	)
}
