// Package telemetry owns the ECS-encoded zap logger and its fx wiring.
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
