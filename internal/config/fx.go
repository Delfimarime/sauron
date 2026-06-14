package config

import (
	"go.uber.org/fx"
)

// NewFxOptions wires the resolved Configuration.
func NewFxOptions() fx.Option {
	return fx.Provide(
		New,
	)
}
