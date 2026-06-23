package agent

import "go.uber.org/fx"

// NewFxOptions wires the provider adapters.
func NewFxOptions() fx.Option {
	return fx.Options()
}
