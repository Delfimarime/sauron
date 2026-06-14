// Package provider wires the provider adapters implementing the pkg/provider port.
package provider

import "go.uber.org/fx"

// NewFxOptions wires the provider adapters.
func NewFxOptions() fx.Option {
	return fx.Options()
}
